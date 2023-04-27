// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	v1 "kpt.dev/configsync/pkg/api/configmanagement/v1"
	"kpt.dev/configsync/pkg/api/configsync"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	hubv1 "kpt.dev/configsync/pkg/api/hub/v1"
	"kpt.dev/configsync/pkg/core"
	"kpt.dev/configsync/pkg/declared"
	"kpt.dev/configsync/pkg/kinds"
	"kpt.dev/configsync/pkg/metadata"
	"kpt.dev/configsync/pkg/metrics"
	"kpt.dev/configsync/pkg/reconcilermanager"
	"kpt.dev/configsync/pkg/reposync"
	"kpt.dev/configsync/pkg/status"
	"kpt.dev/configsync/pkg/util"
	"kpt.dev/configsync/pkg/util/compare"
	"kpt.dev/configsync/pkg/validate/raw/validate"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// RepoSyncReconciler reconciles a RepoSync object.
type RepoSyncReconciler struct {
	reconcilerBase
	// repoSyncs is a cache of the reconciled RepoSync objects.
	repoSyncs map[types.NamespacedName]struct{}

	lock sync.Mutex
}

// NewRepoSyncReconciler returns a new RepoSyncReconciler.
func NewRepoSyncReconciler(clusterName string, reconcilerPollingPeriod, hydrationPollingPeriod time.Duration, client client.Client, dynamicClient dynamic.Interface, log logr.Logger, scheme *runtime.Scheme) *RepoSyncReconciler {
	return &RepoSyncReconciler{
		reconcilerBase: reconcilerBase{
			clusterName:             clusterName,
			client:                  client,
			dynamicClient:           dynamicClient,
			log:                     log,
			scheme:                  scheme,
			reconcilerPollingPeriod: reconcilerPollingPeriod,
			hydrationPollingPeriod:  hydrationPollingPeriod,
			syncKind:                configsync.RepoSyncKind,
		},
		repoSyncs: make(map[types.NamespacedName]struct{}),
	}
}

// +kubebuilder:rbac:groups=configsync.gke.io,resources=reposyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=configsync.gke.io,resources=reposyncs/status,verbs=get;update;patch

// Reconcile the RepoSync resource.
func (r *RepoSyncReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	rsRef := req.NamespacedName
	log := r.log.WithValues("reposync", rsRef.String())
	start := time.Now()
	var err error
	rs := &v1beta1.RepoSync{}
	reconcilerRef := types.NamespacedName{
		Namespace: v1.NSConfigManagementSystem,
		Name:      core.NsReconcilerName(rsRef.Namespace, rsRef.Name),
	}

	if err = r.client.Get(ctx, rsRef, rs); err != nil {
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		if apierrors.IsNotFound(err) {
			r.clearLastReconciled(rsRef)
			if _, ok := r.repoSyncs[rsRef]; !ok {
				log.Error(err, "Sync object not managed by this reconciler-manager",
					logFieldObject, rsRef.String(),
					logFieldKind, r.syncKind)
				// return `controllerruntime.Result{}, nil` here to make sure the request will not be requeued.
				return controllerruntime.Result{}, nil
			}
			// Namespace controller resources are cleaned up when reposync no longer present.
			//
			// Note: Update cleanup resources in cleanupNSControllerResources(...) when
			// resources created by namespace controller changes.
			cleanupErr := r.cleanupNSControllerResources(ctx, rsRef, reconcilerRef)
			// if cleanupErr != nil, the request will be requeued.
			return controllerruntime.Result{}, cleanupErr
		}
		return controllerruntime.Result{}, status.APIServerError(err, "failed to get RepoSync")
	}

	if !rs.DeletionTimestamp.IsZero() {
		// Deletion requested.
		// Cleanup is handled above, after the RepoSync is NotFound.
		log.V(3).Info("Deletion timestamp detected",
			logFieldObject, rsRef.String(),
			logFieldKind, r.syncKind)
	}

	// The caching client sometimes returns an old R*Sync, because the watch
	// hasn't received the update event yet. If so, error and retry.
	// This is an optimization to avoid unnecessary API calls.
	if r.isLastReconciled(rsRef, rs.ResourceVersion) {
		return controllerruntime.Result{}, fmt.Errorf("ResourceVersion already reconciled: %s", rs.ResourceVersion)
	}

	currentRS := rs.DeepCopy()

	if err = r.validateNamespaceName(rsRef.Namespace); err != nil {
		log.Error(err, "Namespace invalid",
			logFieldObject, rsRef.String(),
			logFieldKind, r.syncKind)
		reposync.SetStalled(rs, "Validation", err)
		// Validation errors should not trigger retry (return error),
		// unless the status update also fails.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		// Use the validation error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, updateErr
	}

	r.repoSyncs[rsRef] = struct{}{}
	if errs := validation.IsDNS1123Subdomain(reconcilerRef.Name); errs != nil {
		err = errors.Errorf("Invalid reconciler name %q: %s.", reconcilerRef.Name, strings.Join(errs, ", "))
		log.Error(err, "Name or namespace invalid",
			logFieldObject, rsRef.String(),
			logFieldKind, r.syncKind)
		reposync.SetStalled(rs, "Validation", err)
		// Validation errors should not trigger retry (return error),
		// unless the status update also fails.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		// Use the validation error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, updateErr
	}

	if err = r.validateSpec(ctx, rs, reconcilerRef.Name); err != nil {
		log.Error(err, "Spec invalid",
			logFieldObject, rsRef.String(),
			logFieldKind, r.syncKind)
		reposync.SetStalled(rs, "Validation", err)
		// Validation errors should not trigger retry (return error),
		// unless the status update also fails.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		// Use the validation error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, updateErr
	}

	// Create secret in config-management-system namespace using the
	// existing secret in the reposync.namespace.
	if sRef, err := upsertAuthSecret(ctx, log, rs, r.client, reconcilerRef); err != nil {
		var authType configsync.AuthType
		if rs.Spec.SourceType == string(v1beta1.GitSource) {
			authType = rs.Spec.Auth
		} else if rs.Spec.SourceType == string(v1beta1.HelmSource) {
			authType = rs.Spec.Helm.Auth
		}
		log.Error(err, "Managed object upsert failed",
			logFieldObject, sRef.String(),
			logFieldKind, "Secret",
			"type", "auth",
			"auth", authType)
		reposync.SetStalled(rs, "Secret", err)
		// Upsert errors should always trigger retry (return error),
		// even if status update is successful.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		if updateErr != nil {
			log.Error(updateErr, "Object status update failed",
				logFieldObject, rsRef.String(),
				logFieldKind, r.syncKind)
		}
		// Use the upsert error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, errors.Wrap(err, "Secret reconcile failed")
	}

	// Create secret in config-management-system namespace using the
	// existing secret in the reposync.namespace.
	if sRef, err := upsertCACertSecret(ctx, log, rs, r.client, reconcilerRef); err != nil {
		log.Error(err, "Managed object upsert failed",
			logFieldObject, sRef.String(),
			logFieldKind, "Secret",
			"type", "caCert")
		reposync.SetStalled(rs, "Secret", err)
		// Upsert errors should always trigger retry (return error),
		// even if status update is successful.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		if updateErr != nil {
			log.Error(updateErr, "Object status update failed",
				logFieldObject, rsRef.String(),
				logFieldKind, r.syncKind)
		}
		// Use the upsert error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, errors.Wrap(err, "Secret reconcile failed")
	}

	labelMap := map[string]string{
		metadata.SyncNamespaceLabel: rs.Namespace,
		metadata.SyncNameLabel:      rs.Name,
		metadata.SyncKindLabel:      r.syncKind,
	}

	// Overwrite reconciler pod ServiceAccount.
	var auth configsync.AuthType
	var gcpSAEmail string
	switch v1beta1.SourceType(rs.Spec.SourceType) {
	case v1beta1.GitSource:
		auth = rs.Spec.Auth
		gcpSAEmail = rs.Spec.GCPServiceAccountEmail
	case v1beta1.OciSource:
		auth = rs.Spec.Oci.Auth
		gcpSAEmail = rs.Spec.Oci.GCPServiceAccountEmail
	case v1beta1.HelmSource:
		auth = rs.Spec.Helm.Auth
		gcpSAEmail = rs.Spec.Helm.GCPServiceAccountEmail
	}
	if saRef, err := r.upsertServiceAccount(ctx, reconcilerRef, auth, gcpSAEmail, labelMap); err != nil {
		log.Error(err, "Managed object upsert failed",
			logFieldObject, saRef.String(),
			logFieldKind, "ServiceAccount",
			"sourceType", rs.Spec.SourceType,
			"auth", auth)
		reposync.SetStalled(rs, "ServiceAccount", err)
		// Upsert errors should always trigger retry (return error),
		// even if status update is successful.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		if updateErr != nil {
			log.Error(updateErr, "Object status update failed",
				logFieldObject, rsRef.String(),
				logFieldKind, r.syncKind)
		}
		// Use the upsert error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, errors.Wrap(err, "ServiceAccount reconcile failed")
	}

	// Overwrite reconciler rolebinding.
	if rbRef, err := r.upsertRoleBinding(ctx, reconcilerRef, rsRef); err != nil {
		log.Error(err, "Managed object upsert failed",
			logFieldObject, rbRef.String(),
			logFieldKind, "RoleBinding")
		reposync.SetStalled(rs, "RoleBinding", err)
		// Upsert errors should always trigger retry (return error),
		// even if status update is successful.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		if updateErr != nil {
			log.Error(updateErr, "Object status update failed",
				logFieldObject, rsRef.String(),
				logFieldKind, r.syncKind)
		}
		// Use the upsert error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, errors.Wrap(err, "RoleBinding reconcile failed")
	}

	notificationEnabled, err := util.NotificationEnabled(ctx, r.client, rs.Namespace, rs.Annotations, rs.Spec.NotificationConfig)
	if err != nil {
		log.Error(err, "Check notification configuration failed")
		reposync.SetStalled(rs, "Notification", err)
		// check notification subscription errors should always trigger retry (return error),
		// even if status update is successful.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		if updateErr != nil {
			log.Error(updateErr, "Object status update failed",
				logFieldObject, rsRef.String(),
				logFieldKind, r.syncKind)
		}
		// Use the upsert error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, errors.Wrap(err, "Check notification configuration failed")
	}
	if notificationEnabled {
		if _, err = r.upsertNotificationCR(ctx, rs); err != nil {
			log.Error(err, "Managed object upsert failed",
				logFieldObject, rsRef.String(),
				logFieldKind, "Notification")
			reposync.SetStalled(rs, "Notification", err)
			// Upsert errors should always trigger retry (return error),
			// even if status update is successful.
			_, updateErr := r.updateStatus(ctx, currentRS, rs)
			if updateErr != nil {
				log.Error(updateErr, "Object status update failed",
					logFieldObject, rsRef.String(),
					logFieldKind, r.syncKind)
			}
			// Use the upsert error for metric tagging.
			metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
			return controllerruntime.Result{}, errors.Wrap(err, "Notification upsert failed")
		}
	} else {
		// delete Notification CR, if it exists
		if err := r.deleteNotificationCR(ctx, rsRef); err != nil {
			return controllerruntime.Result{}, status.APIServerError(err, "failed to delete Notification")
		}
	}

	containerEnvs := r.populateContainerEnvs(ctx, rs, reconcilerRef.Name, notificationEnabled)
	mut := r.mutationsFor(ctx, rs, notificationEnabled, containerEnvs)

	// Upsert Namespace reconciler deployment.
	deployObj, op, err := r.upsertDeployment(ctx, reconcilerRef, labelMap, mut)
	if err != nil {
		log.Error(err, "Managed object get failed",
			logFieldObject, reconcilerRef.String(),
			logFieldKind, "Deployment")
		reposync.SetStalled(rs, "Deployment", err)
		// Upsert errors should always trigger retry (return error),
		// even if status update is successful.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		if updateErr != nil {
			log.Error(updateErr, "Object status update failed",
				logFieldObject, rsRef.String(),
				logFieldKind, r.syncKind)
		}
		// Use the upsert error for metric tagging.
		metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
		return controllerruntime.Result{}, errors.Wrap(err, "Deployment reconcile failed")
	}
	rs.Status.Reconciler = reconcilerRef.Name

	// Get the latest deployment to check the status.
	// For other operations, upsertDeployment will have returned the latest already.
	if op == controllerutil.OperationResultNone {
		deployObj, err = r.deployment(ctx, reconcilerRef)
		if err != nil {
			log.Error(err, "Managed object get failed",
				logFieldObject, reconcilerRef.String(),
				logFieldKind, "Deployment")
			reposync.SetStalled(rs, "Deployment", err)
			// Get errors should always trigger retry (return error),
			// even if status update is successful.
			_, updateErr := r.updateStatus(ctx, currentRS, rs)
			if updateErr != nil {
				log.Error(updateErr, "Object status update failed",
					logFieldObject, rsRef.String(),
					logFieldKind, r.syncKind)
			}
			return controllerruntime.Result{}, err
		}
	}

	result, err := kstatus.Compute(deployObj)
	if err != nil {
		log.Error(err, "Managed object status check failed",
			logFieldObject, reconcilerRef.String(),
			logFieldKind, "Deployment")
		reposync.SetStalled(rs, "Deployment", err)
		// Get errors should always trigger retry (return error),
		// even if status update is successful.
		_, updateErr := r.updateStatus(ctx, currentRS, rs)
		if updateErr != nil {
			log.Error(updateErr, "Object status update failed",
				logFieldObject, rsRef.String(),
				logFieldKind, r.syncKind)
		}
		return controllerruntime.Result{}, err
	}

	log.V(3).Info("RepoSync reconciler Deployment status",
		logFieldObject, reconcilerRef.String(),
		"resourceVersion", deployObj.GetResourceVersion(),
		"status", result.Status,
		"message", result.Message)

	// Update RepoSync status based on reconciler deployment condition result.
	switch result.Status {
	case kstatus.InProgressStatus:
		// inProgressStatus indicates that the deployment is not yet
		// available. Hence update the Reconciling status condition.
		reposync.SetReconciling(rs, "Deployment", result.Message)
		// Clear Stalled condition.
		reposync.ClearCondition(rs, v1beta1.RepoSyncStalled)
	case kstatus.FailedStatus:
		// statusFailed indicates that the deployment failed to reconcile. Update
		// Reconciling status condition with appropriate message specifying the
		// reason for failure.
		reposync.SetReconciling(rs, "Deployment", result.Message)
		// Set Stalled condition with the deployment statusFailed.
		reposync.SetStalled(rs, "Deployment", errors.New(string(result.Status)))
	case kstatus.CurrentStatus:
		// currentStatus indicates that the deployment is available, which qualifies
		// to clear the Reconciling status condition in RepoSync.
		reposync.ClearCondition(rs, v1beta1.RepoSyncReconciling)
		// Since there were no errors, we can clear any previous Stalled condition.
		reposync.ClearCondition(rs, v1beta1.RepoSyncStalled)
	}

	updated, err := r.updateStatus(ctx, currentRS, rs)
	// Use the status update error for metric tagging, if no other errors.
	metrics.RecordReconcileDuration(ctx, metrics.StatusTagKey(err), start)
	if err != nil {
		return controllerruntime.Result{}, err
	}

	if updated && result.Status == kstatus.CurrentStatus {
		r.log.Info("Sync object reconcile successful",
			logFieldObject, rsRef.String(),
			logFieldKind, r.syncKind)
	}
	return controllerruntime.Result{}, nil
}

// SetupWithManager registers RepoSync controller with reconciler-manager.
func (r *RepoSyncReconciler) SetupWithManager(mgr controllerruntime.Manager, watchFleetMembership bool) error {
	// Index the `gitSecretRefName` field, so that we will be able to lookup RepoSync be a referenced `SecretRef` name.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.RepoSync{}, gitSecretRefField, func(rawObj client.Object) []string {
		rs := rawObj.(*v1beta1.RepoSync)
		if rs.Spec.Git == nil || v1beta1.GetSecretName(rs.Spec.Git.SecretRef) == "" {
			return nil
		}
		return []string{rs.Spec.Git.SecretRef.Name}
	}); err != nil {
		return err
	}
	// Index the `caCertSecretRefField` field, so that we will be able to lookup RepoSync be a referenced `caCertSecretRefField` name.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.RepoSync{}, caCertSecretRefField, func(rawObj client.Object) []string {
		rs := rawObj.(*v1beta1.RepoSync)
		if rs.Spec.Git == nil || v1beta1.GetSecretName(rs.Spec.Git.CACertSecretRef) == "" {
			return nil
		}
		return []string{rs.Spec.Git.CACertSecretRef.Name}
	}); err != nil {
		return err
	}
	// Index the `helmSecretRefName` field, so that we will be able to lookup RepoSync be a referenced `SecretRef` name.
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.RepoSync{}, helmSecretRefField, func(rawObj client.Object) []string {
		rs := rawObj.(*v1beta1.RepoSync)
		if rs.Spec.Helm == nil || v1beta1.GetSecretName(rs.Spec.Helm.SecretRef) == "" {
			return nil
		}
		return []string{rs.Spec.Helm.SecretRef.Name}
	}); err != nil {
		return err
	}

	controllerBuilder := controllerruntime.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		For(&v1beta1.RepoSync{}).
		// Custom Watch to trigger Reconcile for objects created by RepoSync controller.
		Watches(&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.mapSecretToRepoSyncs),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &appsv1.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(r.mapObjectToRepoSync),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.ServiceAccount{}},
			handler.EnqueueRequestsFromMapFunc(r.mapObjectToRepoSync),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &rbacv1.RoleBinding{}},
			handler.EnqueueRequestsFromMapFunc(r.mapObjectToRepoSync),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}))

	if watchFleetMembership {
		// Custom Watch for membership to trigger reconciliation.
		controllerBuilder.Watches(&source.Kind{Type: &hubv1.Membership{}},
			handler.EnqueueRequestsFromMapFunc(r.mapMembershipToRepoSyncs()),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}))
	}
	return controllerBuilder.Complete(r)
}

func (r *RepoSyncReconciler) mapMembershipToRepoSyncs() func(client.Object) []reconcile.Request {
	return func(o client.Object) []reconcile.Request {
		// Clear the membership if the cluster is unregistered
		if err := r.client.Get(context.Background(), types.NamespacedName{Name: fleetMembershipName}, &hubv1.Membership{}); err != nil {
			if apierrors.IsNotFound(err) {
				klog.Infof("Clear membership cache because %v", err)
				r.membership = nil
				return r.requeueAllRepoSyncs()
			}
			klog.Errorf("failed to get membership: %v", err)
			return nil
		}

		m, isMembership := o.(*hubv1.Membership)
		if !isMembership {
			klog.Error("object is not a type of membership, gvk ", o.GetObjectKind().GroupVersionKind())
			return nil
		}
		if m.Name != fleetMembershipName {
			klog.Error("membership is %s, not 'membership'", m.Name)
			return nil
		}
		r.membership = m
		return r.requeueAllRepoSyncs()
	}
}

func (r *RepoSyncReconciler) requeueAllRepoSyncs() []reconcile.Request {
	allRepoSyncs := &v1beta1.RepoSyncList{}
	if err := r.client.List(context.Background(), allRepoSyncs); err != nil {
		klog.Error("failed to list all RepoSyncs: %v", err)
		return nil
	}

	requests := make([]reconcile.Request, len(allRepoSyncs.Items))
	for i, rs := range allRepoSyncs.Items {
		requests[i] = reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(&rs),
		}
	}
	if len(requests) > 0 {
		klog.Infof("Changes to membership trigger reconciliations for %d RepoSync objects.", len(allRepoSyncs.Items))
	}
	return requests
}

// mapSecretToRepoSyncs define a mapping from the Secret object to its attached
// RepoSync objects via the `spec.git.secretRef.name` field .
// The update to the Secret object will trigger a reconciliation of the RepoSync objects.
func (r *RepoSyncReconciler) mapSecretToRepoSyncs(secret client.Object) []reconcile.Request {
	// map the copied ns-reconciler Secret in the config-management-system to RepoSync request.
	if secret.GetNamespace() == configsync.ControllerNamespace {
		// Ignore secrets in the config-management-system namespace that don't start with ns-reconciler.
		if !strings.HasPrefix(secret.GetName(), core.NsReconcilerPrefix) {
			return nil
		}
		if err := r.addTypeInformationToObject(secret); err != nil {
			klog.Errorf("failed to add type information to object (name: %s, namespace: %s): %v", secret.GetName(), secret.GetNamespace(), err)
			return nil
		}
		allRepoSyncs := &v1beta1.RepoSyncList{}
		if err := r.client.List(context.Background(), allRepoSyncs); err != nil {
			klog.Error("failed to list all RepoSyncs for object (name: %s, namespace: %s): %v", secret.GetName(), secret.GetNamespace(), err)
			return nil
		}
		for _, rs := range allRepoSyncs.Items {
			// It is a one-to-one mapping between the copied ns-reconciler Secret and the RepoSync object,
			// so requeue the mapped RepoSync object and then return.
			reconcilerName := core.NsReconcilerName(rs.GetNamespace(), rs.GetName())
			if isUpsertedSecret(&rs, secret.GetName()) {
				return requeueRepoSyncRequest(secret, &rs)
			}
			isSAToken := strings.HasPrefix(secret.GetName(), reconcilerName+"-token-")
			if isSAToken {
				saRef := client.ObjectKey{
					Name:      reconcilerName,
					Namespace: configsync.ControllerNamespace,
				}
				serviceAccount := &corev1.ServiceAccount{}
				if err := r.client.Get(context.Background(), saRef, serviceAccount); err != nil {
					klog.Error("failed to get ServiceAccount %q in the %q namespace: %v", secret.GetName(), secret.GetNamespace(), err)
					return nil
				}
				for _, s := range serviceAccount.Secrets {
					if s.Name == secret.GetName() {
						return requeueRepoSyncRequest(secret, &rs)
					}
				}
			}
		}
		return nil
	}

	// map the user-managed ns-reconciler Secret in the RepoSync's namespace to RepoSync request.
	// The user-managed ns-reconciler Secret might be shared among multiple RepoSync objects in the same namespace,
	// so requeue all the attached RepoSync objects.
	attachedRepoSyncs := &v1beta1.RepoSyncList{}
	secretFields := []string{gitSecretRefField, caCertSecretRefField, helmSecretRefField}
	for _, secretField := range secretFields {
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(secretField, secret.GetName()),
			Namespace:     secret.GetNamespace(),
		}
		fetchedRepoSyncs := &v1beta1.RepoSyncList{}
		if err := r.client.List(context.Background(), fetchedRepoSyncs, listOps); err != nil {
			klog.Errorf("failed to list attached RepoSyncs for secret (name: %s, namespace: %s): %v", secret.GetName(), secret.GetNamespace(), err)
			return nil
		}
		attachedRepoSyncs.Items = append(attachedRepoSyncs.Items, fetchedRepoSyncs.Items...)
	}
	requests := make([]reconcile.Request, len(attachedRepoSyncs.Items))
	attachedRSNames := make([]string, len(attachedRepoSyncs.Items))
	for i, rs := range attachedRepoSyncs.Items {
		attachedRSNames[i] = rs.GetName()
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      rs.GetName(),
				Namespace: rs.GetNamespace(),
			},
		}
	}
	if len(requests) > 0 {
		klog.Infof("Changes to Secret (name: %s, namespace: %s) triggers a reconciliation for the RepoSync object %q in the same namespace.", secret.GetName(), secret.GetNamespace(), strings.Join(attachedRSNames, ", "))
	}
	return requests
}

// mapObjectToRepoSync define a mapping from an object in 'config-management-system'
// namespace to a RepoSync to be reconciled.
func (r *RepoSyncReconciler) mapObjectToRepoSync(obj client.Object) []reconcile.Request {
	// Ignore changes from other namespaces because all the generated resources
	// exist in the config-management-system namespace.
	if obj.GetNamespace() != configsync.ControllerNamespace {
		return nil
	}

	// Ignore changes from resources without the ns-reconciler prefix or configsync.gke.io:ns-reconciler
	// because all the generated resources have the prefix.
	nsRoleBindingName := RepoSyncPermissionsName()
	if !strings.HasPrefix(obj.GetName(), core.NsReconcilerPrefix) && obj.GetName() != nsRoleBindingName {
		return nil
	}

	if err := r.addTypeInformationToObject(obj); err != nil {
		klog.Errorf("failed to add type information to object (name: %s, namespace: %s): %v", obj.GetName(), obj.GetNamespace(), err)
		return nil
	}

	allRepoSyncs := &v1beta1.RepoSyncList{}
	if err := r.client.List(context.Background(), allRepoSyncs); err != nil {
		klog.Error("failed to list all RepoSyncs for object (name: %s, namespace: %s): %v", obj.GetName(), obj.GetNamespace(), err)
		return nil
	}

	// Most of the resources are mapped to a single RepoSync object except RoleBinding.
	// All RepoSync objects share the same RoleBinding object, so requeue all RepoSync objects the RoleBinding is changed.
	// For other resources, requeue the mapping RepoSync object and then return.
	var requests []reconcile.Request
	var attachedRSNames []string
	for _, rs := range allRepoSyncs.Items {
		reconcilerName := core.NsReconcilerName(rs.GetNamespace(), rs.GetName())
		switch obj.(type) {
		case *rbacv1.RoleBinding:
			if obj.GetName() == nsRoleBindingName {
				requests = append(requests, reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&rs),
				})
				attachedRSNames = append(attachedRSNames, rs.GetName())
			}
		default: // Deployment and ServiceAccount
			if obj.GetName() == reconcilerName {
				return requeueRepoSyncRequest(obj, &rs)
			}
		}
	}
	if len(requests) > 0 {
		klog.Infof("Changes to %s (name: %s, namespace: %s) triggers a reconciliation for the RepoSync objects %q in the same namespace.",
			obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName(), obj.GetNamespace(), strings.Join(attachedRSNames, ", "))
	}
	return requests
}

// addTypeInformationToObject looks up and adds GVK to a runtime.Object based upon the loaded Scheme
func (r *RepoSyncReconciler) addTypeInformationToObject(obj runtime.Object) error {
	gvk, err := kinds.Lookup(obj, r.scheme)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return nil
}

func requeueRepoSyncRequest(obj client.Object, rs *v1beta1.RepoSync) []reconcile.Request {
	klog.Infof("Changes to %s (name: %s, namespace: %s) triggers a reconciliation for the RepoSync object %q in the same namespace.",
		obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName(), obj.GetNamespace(), rs.GetName())
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      rs.GetName(),
				Namespace: rs.GetNamespace(),
			},
		},
	}
}

func (r *RepoSyncReconciler) populateContainerEnvs(ctx context.Context, rs *v1beta1.RepoSync, reconcilerName string, notificationEnabled bool) map[string][]corev1.EnvVar {
	result := map[string][]corev1.EnvVar{
		reconcilermanager.HydrationController: hydrationEnvs(rs.Spec.SourceType, rs.Spec.Git, rs.Spec.Oci, declared.Scope(rs.Namespace), reconcilerName, r.hydrationPollingPeriod.String()),
		reconcilermanager.Reconciler:          reconcilerEnvs(r.clusterName, rs.Name, reconcilerName, declared.Scope(rs.Namespace), rs.Spec.SourceType, rs.Spec.Git, rs.Spec.Oci, reposync.GetHelmBase(rs.Spec.Helm), r.reconcilerPollingPeriod.String(), rs.Spec.SafeOverride().StatusMode, v1beta1.GetReconcileTimeout(rs.Spec.SafeOverride().ReconcileTimeout), v1beta1.GetAPIServerTimeout(rs.Spec.SafeOverride().APIServerTimeout)),
	}
	if notificationEnabled {
		result[reconcilermanager.Notification] = notificationEnvs(r.syncKind, rs.Namespace, rs.Name, rs.Spec.NotificationConfig)
	}
	switch v1beta1.SourceType(rs.Spec.SourceType) {
	case v1beta1.GitSource:
		result[reconcilermanager.GitSync] = gitSyncEnvs(ctx, options{
			ref:             rs.Spec.Git.Revision,
			branch:          rs.Spec.Git.Branch,
			repo:            rs.Spec.Git.Repo,
			secretType:      rs.Spec.Git.Auth,
			period:          v1beta1.GetPeriodSecs(rs.Spec.Git.Period),
			proxy:           rs.Spec.Proxy,
			depth:           rs.Spec.SafeOverride().GitSyncDepth,
			noSSLVerify:     rs.Spec.Git.NoSSLVerify,
			caCertSecretRef: v1beta1.GetSecretName(rs.Spec.Git.CACertSecretRef),
		})
	case v1beta1.OciSource:
		result[reconcilermanager.OciSync] = ociSyncEnvs(rs.Spec.Oci.Image, rs.Spec.Oci.Auth, v1beta1.GetPeriodSecs(rs.Spec.Oci.Period))
	case v1beta1.HelmSource:
		result[reconcilermanager.HelmSync] = helmSyncEnvs(&rs.Spec.Helm.HelmBase, rs.Namespace, "")
	}
	return result
}

func (r *RepoSyncReconciler) validateSpec(ctx context.Context, rs *v1beta1.RepoSync, reconcilerName string) error {
	switch v1beta1.SourceType(rs.Spec.SourceType) {
	case v1beta1.GitSource:
		return r.validateGitSpec(ctx, rs, reconcilerName)
	case v1beta1.OciSource:
		return validate.OciSpec(rs.Spec.Oci, rs)
	case v1beta1.HelmSource:
		return validate.HelmSpec(reposync.GetHelmBase(rs.Spec.Helm), rs)
	default:
		return validate.InvalidSourceType(rs)
	}
}

func (r *RepoSyncReconciler) validateGitSpec(ctx context.Context, rs *v1beta1.RepoSync, reconcilerName string) error {
	if err := validate.GitSpec(rs.Spec.Git, rs); err != nil {
		return err
	}
	if err := r.validateCACertSecret(ctx, rs.Namespace, v1beta1.GetSecretName(rs.Spec.Git.CACertSecretRef)); err != nil {
		return err
	}
	return r.validateNamespaceSecret(ctx, rs, reconcilerName)
}

// validateNamespaceSecret verify that any necessary Secret is present before creating ConfigMaps and Deployments.
func (r *RepoSyncReconciler) validateNamespaceSecret(ctx context.Context, repoSync *v1beta1.RepoSync, reconcilerName string) error {
	var authType configsync.AuthType
	var namespaceSecretName string
	if repoSync.Spec.SourceType == string(v1beta1.GitSource) {
		authType = repoSync.Spec.Auth
		namespaceSecretName = v1beta1.GetSecretName(repoSync.Spec.SecretRef)
	} else if repoSync.Spec.SourceType == string(v1beta1.HelmSource) {
		authType = repoSync.Spec.Helm.Auth
		namespaceSecretName = v1beta1.GetSecretName(repoSync.Spec.Helm.SecretRef)
	}
	if SkipForAuth(authType) {
		// There is no Secret to check for the Config object.
		return nil
	}

	secretName := ReconcilerResourceName(reconcilerName, namespaceSecretName)
	if errs := validation.IsDNS1123Subdomain(secretName); errs != nil {
		return errors.Errorf("The managed secret name %q is invalid: %s. To fix it, update '.spec.git.secretRef.name'", secretName, strings.Join(errs, ", "))
	}

	secret, err := validateSecretExist(ctx,
		namespaceSecretName,
		repoSync.Namespace,
		r.client)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return errors.Errorf("Secret %s not found: create one to allow client authentication", namespaceSecretName)
		}
		return errors.Wrapf(err, "Secret %s get failed", namespaceSecretName)
	}
	return validateSecretData(authType, secret)
}

func (r *RepoSyncReconciler) validateNamespaceName(namespaceName string) error {
	if namespaceName == configsync.ControllerNamespace {
		return fmt.Errorf("RepoSync objects are not allowed in the %s namespace", configsync.ControllerNamespace)
	}
	return nil
}

func (r *RepoSyncReconciler) upsertRoleBinding(ctx context.Context, reconcilerRef, rsRef types.NamespacedName) (client.ObjectKey, error) {
	rbRef := client.ObjectKey{
		Namespace: rsRef.Namespace,
		Name:      RepoSyncPermissionsName(),
	}
	childRB := &rbacv1.RoleBinding{}
	childRB.Name = rbRef.Name
	childRB.Namespace = rbRef.Namespace

	op, err := controllerruntime.CreateOrUpdate(ctx, r.client, childRB, func() error {
		childRB.RoleRef = rolereference(RepoSyncPermissionsName(), "ClusterRole")
		childRB.Subjects = addSubject(childRB.Subjects, r.serviceAccountSubject(reconcilerRef))
		return nil
	})
	if err != nil {
		return rbRef, err
	}
	if op != controllerutil.OperationResultNone {
		r.log.Info("Managed object upsert successful",
			logFieldObject, rbRef.String(),
			logFieldKind, "RoleBinding",
			logFieldOperation, op)
	}
	return rbRef, nil
}

func (r *RepoSyncReconciler) updateStatus(ctx context.Context, currentRS, rs *v1beta1.RepoSync) (bool, error) {
	rs.Status.ObservedGeneration = rs.Generation

	// Avoid unnecessary status updates.
	if cmp.Equal(currentRS.Status, rs.Status, compare.IgnoreTimestampUpdates) {
		klog.V(3).Infof("Skipping status update for RepoSync %s (ResourceVersion: %s)",
			client.ObjectKeyFromObject(rs), rs.ResourceVersion)
		return false, nil
	}

	if klog.V(5).Enabled() {
		klog.Infof("Updating status for RepoSync %s (ResourceVersion: %s):\nDiff (- Expected, + Actual):\n%s",
			client.ObjectKeyFromObject(rs), rs.ResourceVersion,
			cmp.Diff(currentRS.Status, rs.Status))
	}

	resourceVersion := rs.ResourceVersion

	err := r.client.Status().Update(ctx, rs)
	if err != nil {
		return false, err
	}

	// Register the latest ResourceVersion as reconciled.
	r.setLastReconciled(core.ObjectNamespacedName(rs), resourceVersion)
	return true, nil
}

func (r *RepoSyncReconciler) mutationsFor(ctx context.Context, rs *v1beta1.RepoSync, notificationEnabled bool, containerEnvs map[string][]corev1.EnvVar) mutateFn {
	return func(obj client.Object) error {
		d, ok := obj.(*appsv1.Deployment)
		if !ok {
			return errors.Errorf("expected appsv1 Deployment, got: %T", obj)
		}
		reconcilerName := core.NsReconcilerName(rs.Namespace, rs.Name)

		// Only inject the FWI credentials when the auth type is gcpserviceaccount and the membership info is available.
		var auth configsync.AuthType
		var gcpSAEmail string
		var secretRefName string
		var caCertSecretRefName string
		switch v1beta1.SourceType(rs.Spec.SourceType) {
		case v1beta1.GitSource:
			auth = rs.Spec.Auth
			gcpSAEmail = rs.Spec.GCPServiceAccountEmail
			secretRefName = v1beta1.GetSecretName(rs.Spec.SecretRef)
			caCertSecretRefName = v1beta1.GetSecretName(rs.Spec.Git.CACertSecretRef)
		case v1beta1.OciSource:
			auth = rs.Spec.Oci.Auth
			gcpSAEmail = rs.Spec.Oci.GCPServiceAccountEmail
		case v1beta1.HelmSource:
			auth = rs.Spec.Helm.Auth
			gcpSAEmail = rs.Spec.Helm.GCPServiceAccountEmail
			secretRefName = v1beta1.GetSecretName(rs.Spec.Helm.SecretRef)
		}
		injectFWICreds := useFWIAuth(auth, r.membership)
		if injectFWICreds {
			if err := r.injectFleetWorkloadIdentityCredentials(&d.Spec.Template, gcpSAEmail); err != nil {
				return err
			}
		}

		// Add unique reconciler label
		core.SetLabel(&d.Spec.Template, metadata.ReconcilerLabel, reconcilerName)

		templateSpec := &d.Spec.Template.Spec
		// Update ServiceAccountName. eg. ns-reconciler-<namespace>
		templateSpec.ServiceAccountName = reconcilerName
		// The Deployment object fetched from the API server has the field defined.
		// Update DeprecatedServiceAccount to avoid discrepancy in equality check.
		templateSpec.DeprecatedServiceAccount = reconcilerName
		// Mutate secret.secretname to secret reference specified in RepoSync CR.
		// Secret reference is the name of the secret used by git-sync or helm-sync container to
		// authenticate with the git or helm repository using the authorization method specified
		// in the RepoSync CR.
		secretName := ReconcilerResourceName(reconcilerName, secretRefName)
		if useCACert(caCertSecretRefName) {
			caCertSecretRefName = ReconcilerResourceName(reconcilerName, caCertSecretRefName)
		}
		templateSpec.Volumes = filterVolumes(templateSpec.Volumes, auth, secretName, caCertSecretRefName, rs.Spec.SourceType, r.membership)
		var updatedContainers []corev1.Container
		// Mutate spec.Containers to update name, configmap references and volumemounts.
		for _, container := range templateSpec.Containers {
			addContainer := true
			switch container.Name {
			case reconcilermanager.Reconciler:
				container.Env = append(container.Env, containerEnvs[container.Name]...)
				mutateContainerResource(&container, rs.Spec.Override)
			case reconcilermanager.HydrationController:
				container.Env = append(container.Env, containerEnvs[container.Name]...)
				if rs.Spec.SafeOverride().EnableShellInRendering == nil || !*rs.Spec.SafeOverride().EnableShellInRendering {
					container.Image = strings.ReplaceAll(container.Image, reconcilermanager.HydrationControllerWithShell, reconcilermanager.HydrationController)
				} else {
					container.Image = strings.ReplaceAll(container.Image, reconcilermanager.HydrationController+":", reconcilermanager.HydrationControllerWithShell+":")
				}
				mutateContainerResource(&container, rs.Spec.Override)
			case reconcilermanager.OciSync:
				// Don't add the oci-sync container when sourceType is NOT oci.
				if v1beta1.SourceType(rs.Spec.SourceType) != v1beta1.OciSource {
					addContainer = false
				} else {
					container.Env = append(container.Env, containerEnvs[container.Name]...)
					injectFWICredsToContainer(&container, injectFWICreds)
					mutateContainerResource(&container, rs.Spec.Override)
				}
			case reconcilermanager.HelmSync:
				// Don't add the helm-sync container when sourceType is NOT helm.
				if v1beta1.SourceType(rs.Spec.SourceType) != v1beta1.HelmSource {
					addContainer = false
				} else {
					container.Env = append(container.Env, containerEnvs[container.Name]...)
					container.VolumeMounts = volumeMounts(rs.Spec.Helm.Auth, "", rs.Spec.SourceType, container.VolumeMounts)
					if authTypeToken(rs.Spec.Helm.Auth) {
						container.Env = append(container.Env, helmSyncTokenAuthEnv(secretName)...)
					}
					injectFWICredsToContainer(&container, injectFWICreds)
					mutateContainerResource(&container, rs.Spec.Override)
				}
			case reconcilermanager.GitSync:
				// Don't add the git-sync container when sourceType is NOT git.
				if v1beta1.SourceType(rs.Spec.SourceType) != v1beta1.GitSource {
					addContainer = false
				} else {
					container.Env = append(container.Env, containerEnvs[container.Name]...)
					// Don't mount git-creds volume if auth is 'none' or 'gcenode'.
					container.VolumeMounts = volumeMounts(rs.Spec.Auth, caCertSecretRefName, rs.Spec.SourceType, container.VolumeMounts)
					// Update Environment variables for `token` Auth, which
					// passes the credentials as the Username and Password.
					if authTypeToken(rs.Spec.Auth) {
						container.Env = append(container.Env, gitSyncTokenAuthEnv(secretName)...)
					}
					sRef := client.ObjectKey{Namespace: rs.Namespace, Name: v1beta1.GetSecretName(rs.Spec.SecretRef)}
					keys := GetSecretKeys(ctx, r.client, sRef)
					container.Env = append(container.Env, gitSyncHTTPSProxyEnv(secretName, keys)...)
					mutateContainerResource(&container, rs.Spec.Override)
				}
			case reconcilermanager.Notification:
				if !notificationEnabled {
					addContainer = false
				} else {
					container.Env = append(container.Env, containerEnvs[container.Name]...)
				}
			case metrics.OtelAgentName:
				// The no-op case to avoid unknown container error after
				// first-ever reconcile.
			default:
				return errors.Errorf("unknown container in reconciler deployment template: %q", container.Name)
			}
			if addContainer {
				updatedContainers = append(updatedContainers, container)
			}
		}

		// Add container spec for the "gcenode-askpass-sidecar" (defined as
		// a constant) to the reconciler Deployment when `.spec.sourceType` is `git`,
		// and `.spec.git.auth` is either `gcenode` or `gcpserviceaccount`.
		// The container is added first time when the reconciler deployment is created.
		if v1beta1.SourceType(rs.Spec.SourceType) == v1beta1.GitSource &&
			(auth == configsync.AuthGCPServiceAccount || auth == configsync.AuthGCENode) {
			updatedContainers = append(updatedContainers, gceNodeAskPassSidecar(gcpSAEmail, injectFWICreds))
		}

		templateSpec.Containers = updatedContainers
		return nil
	}
}
