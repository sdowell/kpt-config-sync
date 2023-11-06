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

package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	jserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"kpt.dev/configsync/e2e"
	"kpt.dev/configsync/e2e/nomostest"
	"kpt.dev/configsync/e2e/nomostest/ntopts"
	nomostesting "kpt.dev/configsync/e2e/nomostest/testing"
	"kpt.dev/configsync/e2e/nomostest/testpredicates"
	v1 "kpt.dev/configsync/pkg/api/configmanagement/v1"
	"kpt.dev/configsync/pkg/api/configsync"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	"kpt.dev/configsync/pkg/core"
	"kpt.dev/configsync/pkg/kinds"
	"kpt.dev/configsync/pkg/metadata"
	"kpt.dev/configsync/pkg/metrics"
	"kpt.dev/configsync/pkg/reconcilermanager"
	"kpt.dev/configsync/pkg/reconcilermanager/controllers"
	"kpt.dev/configsync/pkg/testing/fake"
	"kpt.dev/configsync/pkg/util/log"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestReconcilerManagerNormalTeardown validates that when a RootSync or
// RepoSync is deleted, the reconciler-manager finalizer handles deletion of the
// reconciler and its dependencies managed by the reconciler-manager.
func TestReconcilerManagerNormalTeardown(t *testing.T) {
	testNamespace := "teardown"
	nt := nomostest.New(t, nomostesting.ACMController,
		ntopts.WithDelegatedControl, ntopts.Unstructured,
		ntopts.NamespaceRepo(testNamespace, configsync.RepoSyncName))

	t.Log("Validate the reconciler-manager deployment")
	reconcilerManager := &appsv1.Deployment{}
	setNN(reconcilerManager, client.ObjectKey{Name: reconcilermanager.ManagerName, Namespace: v1.NSConfigManagementSystem})
	err := nt.Validate(reconcilerManager.Name, reconcilerManager.Namespace, reconcilerManager)
	require.NoError(t, err)

	t.Log("Validate the RootSync")
	rootSync := &v1beta1.RootSync{}
	setNN(rootSync, client.ObjectKey{Name: configsync.RootSyncName, Namespace: v1.NSConfigManagementSystem})
	err = nt.Watcher.WatchObject(kinds.RootSyncV1Beta1(), rootSync.Name, rootSync.Namespace, []testpredicates.Predicate{
		testpredicates.StatusEquals(nt.Scheme, kstatus.CurrentStatus),
		testpredicates.HasFinalizer(metadata.ReconcilerManagerFinalizer),
	})
	require.NoError(t, err)

	t.Log("Validate the RootSync reconciler and its dependencies")
	rootSyncDependencies := validateRootSyncDependencies(nt, rootSync.Name)

	t.Log("Validate the RepoSync")
	repoSync := &v1beta1.RepoSync{}
	setNN(repoSync, client.ObjectKey{Name: configsync.RepoSyncName, Namespace: testNamespace})
	err = nt.Watcher.WatchObject(kinds.RepoSyncV1Beta1(), repoSync.Name, repoSync.Namespace, []testpredicates.Predicate{
		testpredicates.StatusEquals(nt.Scheme, kstatus.CurrentStatus),
		testpredicates.HasFinalizer(metadata.ReconcilerManagerFinalizer),
	})
	require.NoError(t, err)

	t.Log("Validate the RepoSync reconciler and its dependencies")
	repoSyncDependencies := validateRepoSyncDependencies(nt, repoSync.Namespace, repoSync.Name)

	t.Log("Delete the RootSync and wait for it to be not found")
	err = nomostest.DeleteObjectsAndWait(nt, rootSync)
	require.NoError(t, err)

	t.Log("Validate the RootSync reconciler and its dependencies were deleted")
	for _, obj := range rootSyncDependencies {
		validateObjectNotFound(nt, obj)
	}

	t.Log("Delete the RepoSync and wait for it to be not found")
	err = nomostest.DeleteObjectsAndWait(nt, repoSync)
	require.NoError(t, err)

	t.Log("Validate the RepoSync reconciler and its dependencies were deleted")
	for _, obj := range repoSyncDependencies {
		validateObjectNotFound(nt, obj)
	}
}

// TestReconcilerManagerTeardownInvalidRSyncs validates that the
// reconciler-manager finalizer can handle deletion of the reconciler and its
// dependencies managed by the reconciler-manager even when the RootSync or
// RepoSync is invalid.
func TestReconcilerManagerTeardownInvalidRSyncs(t *testing.T) {
	testNamespace := "invalid-teardown"
	nt := nomostest.New(t, nomostesting.ACMController,
		ntopts.WithDelegatedControl, ntopts.Unstructured,
		ntopts.NamespaceRepo(testNamespace, configsync.RepoSyncName))

	t.Log("Validate the reconciler-manager deployment")
	reconcilerManager := &appsv1.Deployment{}
	setNN(reconcilerManager, client.ObjectKey{Name: reconcilermanager.ManagerName, Namespace: v1.NSConfigManagementSystem})
	err := nt.Validate(reconcilerManager.Name, reconcilerManager.Namespace, reconcilerManager)
	require.NoError(t, err)

	t.Log("Validate the RootSync")
	rootSync := &v1beta1.RootSync{}
	setNN(rootSync, client.ObjectKey{Name: configsync.RootSyncName, Namespace: v1.NSConfigManagementSystem})
	err = nt.Watcher.WatchObject(kinds.RootSyncV1Beta1(), rootSync.Name, rootSync.Namespace, []testpredicates.Predicate{
		testpredicates.StatusEquals(nt.Scheme, kstatus.CurrentStatus),
		testpredicates.HasFinalizer(metadata.ReconcilerManagerFinalizer),
	})
	require.NoError(t, err)

	t.Log("Validate the RootSync reconciler and its dependencies")
	rootSyncDependencies := validateRootSyncDependencies(nt, rootSync.Name)

	t.Log("Reset spec.git.auth to make RootSync invalid")
	t.Cleanup(func() {
		if err := nt.KubeClient.Get(rootSync.Name, rootSync.Namespace, &v1beta1.RootSync{}); err != nil {
			if !apierrors.IsNotFound(err) {
				t.Fatal(err)
			}
		} else {
			nt.MustMergePatch(rootSync, fmt.Sprintf(`{"spec":{"git":{"auth": "%s"}}}`, rootSync.Spec.Auth))
		}
	})
	nt.MustMergePatch(rootSync, `{"spec":{"git":{"auth": "token"}}}`)
	nt.WaitForRootSyncStalledError(rootSync.Namespace, rootSync.Name,
		"Validation", `git secretType was set as "token" but token key is not present in git-creds secret`)

	t.Log("Validate the RepoSync")
	repoSync := &v1beta1.RepoSync{}
	setNN(repoSync, client.ObjectKey{Name: configsync.RepoSyncName, Namespace: testNamespace})
	err = nt.Watcher.WatchObject(kinds.RepoSyncV1Beta1(), repoSync.Name, repoSync.Namespace, []testpredicates.Predicate{
		testpredicates.StatusEquals(nt.Scheme, kstatus.CurrentStatus),
		testpredicates.HasFinalizer(metadata.ReconcilerManagerFinalizer),
	})
	require.NoError(t, err)

	t.Log("Validate the RepoSync reconciler and its dependencies")
	repoSyncDependencies := validateRepoSyncDependencies(nt, repoSync.Namespace, repoSync.Name)

	t.Log("Reset spec.git.auth to make RepoSync invalid")
	t.Cleanup(func() {
		if err := nt.KubeClient.Get(repoSync.Name, repoSync.Namespace, &v1beta1.RepoSync{}); err != nil {
			if !apierrors.IsNotFound(err) {
				t.Fatal(err)
			}
		} else {
			nt.MustMergePatch(repoSync, fmt.Sprintf(`{"spec":{"git":{"auth": "%s"}}}`, repoSync.Spec.Auth))
		}
	})
	nt.MustMergePatch(repoSync, `{"spec":{"git":{"auth": "token"}}}`)
	nt.WaitForRepoSyncStalledError(repoSync.Namespace, repoSync.Name,
		"Validation", `git secretType was set as "token" but token key is not present in ssh-key secret`)

	t.Log("Delete the RootSync and wait for it to be not found")
	err = nomostest.DeleteObjectsAndWait(nt, rootSync)
	require.NoError(t, err)

	t.Log("Validate the RootSync reconciler and its dependencies were deleted")
	for _, obj := range rootSyncDependencies {
		validateObjectNotFound(nt, obj)
	}

	t.Log("Delete the RepoSync and wait for it to be not found")
	err = nomostest.DeleteObjectsAndWait(nt, repoSync)
	require.NoError(t, err)

	t.Log("Validate the RepoSync reconciler and its dependencies were deleted")
	for _, obj := range repoSyncDependencies {
		validateObjectNotFound(nt, obj)
	}
}

func validateRootSyncDependencies(nt *nomostest.NT, rsName string) []client.Object {
	var rootSyncDependencies []client.Object

	rootSyncReconciler := &appsv1.Deployment{}
	setNN(rootSyncReconciler, core.RootReconcilerObjectKey(rsName))
	rootSyncDependencies = append(rootSyncDependencies, rootSyncReconciler)

	// Note: reconciler-manager no longer applies ConfigMaps to configure the
	// reconciler. So we don't need to validate their deletion. The deletion
	// only happens when upgrading from a very old unsupported version.

	rootSyncCRB := &rbacv1.ClusterRoleBinding{}
	setNN(rootSyncCRB, client.ObjectKey{Name: controllers.RootSyncPermissionsName()})
	rootSyncDependencies = append(rootSyncDependencies, rootSyncCRB)

	rootSyncSA := &corev1.ServiceAccount{}
	setNN(rootSyncSA, client.ObjectKeyFromObject(rootSyncReconciler))
	rootSyncDependencies = append(rootSyncDependencies, rootSyncSA)

	for _, obj := range rootSyncDependencies {
		err := nt.Validate(obj.GetName(), obj.GetNamespace(), obj)
		require.NoError(nt.T, err)
	}

	return rootSyncDependencies
}

func validateRepoSyncDependencies(nt *nomostest.NT, ns, rsName string) []client.Object {
	var repoSyncDependencies []client.Object
	repoSyncReconciler := &appsv1.Deployment{}
	setNN(repoSyncReconciler, core.NsReconcilerObjectKey(ns, rsName))
	repoSyncDependencies = append(repoSyncDependencies, repoSyncReconciler)

	// Note: reconciler-manager no longer applies ConfigMaps to configure the
	// reconciler. So we don't need to validate their deletion. The deletion
	// only happens when upgrading from a very old unsupported version.

	repoSyncRB := &rbacv1.RoleBinding{}
	setNN(repoSyncRB, client.ObjectKey{
		Name:      controllers.RepoSyncPermissionsName(),
		Namespace: ns,
	})
	repoSyncDependencies = append(repoSyncDependencies, repoSyncRB)

	repoSyncSA := &corev1.ServiceAccount{}
	setNN(repoSyncSA, client.ObjectKeyFromObject(repoSyncReconciler))
	repoSyncDependencies = append(repoSyncDependencies, repoSyncSA)

	// See nomostest.CreateNamespaceSecret for creation of user secrets.
	// This is a managed secret with a derivative name.
	repoSyncAuthSecret := &corev1.Secret{}
	setNN(repoSyncAuthSecret, client.ObjectKey{
		Name:      controllers.ReconcilerResourceName(repoSyncReconciler.Name, nomostest.NamespaceAuthSecretName),
		Namespace: repoSyncReconciler.Namespace,
	})
	repoSyncDependencies = append(repoSyncDependencies, repoSyncAuthSecret)

	// See nomostest.CreateNamespaceSecret for creation of user secrets.
	// This is a managed secret with a derivative name.
	// For local kind clusters, the CA Certs are provided to authenticate the git server.
	if nt.GitProvider.Type() == e2e.Local {
		repoSyncCACertSecret := &corev1.Secret{}
		setNN(repoSyncCACertSecret, client.ObjectKey{
			Name:      controllers.ReconcilerResourceName(repoSyncReconciler.Name, nomostest.NamespaceAuthSecretName),
			Namespace: repoSyncReconciler.Namespace,
		})
		repoSyncDependencies = append(repoSyncDependencies, repoSyncCACertSecret)
	}

	for _, obj := range repoSyncDependencies {
		err := nt.Validate(obj.GetName(), obj.GetNamespace(), obj)
		require.NoError(nt.T, err)
	}
	return repoSyncDependencies
}

func validateObjectNotFound(nt *nomostest.NT, o client.Object) {
	gvk, err := kinds.Lookup(o, nt.Scheme)
	require.NoError(nt.T, err)
	rObj, err := kinds.NewObjectForGVK(gvk, nt.Scheme)
	require.NoError(nt.T, err)
	cObj, err := kinds.ObjectAsClientObject(rObj)
	require.NoError(nt.T, err)
	err = nt.ValidateNotFound(o.GetName(), o.GetNamespace(), cObj)
	require.NoError(nt.T, err)
}

func setNN(obj client.Object, nn types.NamespacedName) {
	obj.SetName(nn.Name)
	obj.SetNamespace(nn.Namespace)
}

func TestManagingReconciler(t *testing.T) {
	nt := nomostest.New(t, nomostesting.ACMController)

	reconcilerDeployment := &appsv1.Deployment{}
	if err := nt.Validate(nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem, reconcilerDeployment); err != nil {
		nt.T.Fatal(err)
	}
	generation := reconcilerDeployment.Generation
	originalImagePullPolicy := testpredicates.ContainerByName(reconcilerDeployment, reconcilermanager.Reconciler).ImagePullPolicy
	updatedImagePullPolicy := corev1.PullAlways
	require.NotEqual(nt.T, updatedImagePullPolicy, originalImagePullPolicy)
	managedReplicas := *reconcilerDeployment.Spec.Replicas
	originalTolerations := reconcilerDeployment.Spec.Template.Spec.Tolerations

	// test case 1: The reconciler-manager should manage most of the fields with one exception:
	// - changes to the container resource requirements should be ignored when the autopilot annotation is set.
	nt.T.Log("Manually update the ImagePullPolicy")
	mustUpdateRootReconciler(nt, func(d *appsv1.Deployment) {
		for i, container := range d.Spec.Template.Spec.Containers {
			if container.Name == reconcilermanager.Reconciler {
				d.Spec.Template.Spec.Containers[i].ImagePullPolicy = updatedImagePullPolicy
			}
		}
	})
	nt.T.Log("Verify the ImagePullPolicy should be reverted by the reconciler-manager")
	generation += 2 // generation bumped by 2 because the change will be first applied then reverted by the reconciler-manager
	err := nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{
			testpredicates.HasGenerationAtLeast(generation),
			testpredicates.DeploymentContainerPullPolicyEquals(reconcilermanager.Reconciler, originalImagePullPolicy),
		})
	if err != nil {
		nt.T.Fatal(err)
	}

	// test case 2: the reconciler-manager should manage the replicas field, so that the reconciler can be resumed after pause.
	nt.T.Log("Manually update the replicas")
	newReplicas := managedReplicas - 1
	nt.MustMergePatch(reconcilerDeployment, fmt.Sprintf(`{"spec": {"replicas": %d}}`, newReplicas))
	nt.T.Log("Verify the reconciler-manager should revert the replicas change")
	generation += 2 // generation bumped by 2 because the change will be first applied then reverted by the reconciler-manager
	err = nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{testpredicates.HasGenerationAtLeast(generation), hasReplicas(managedReplicas)})
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = getDeploymentGeneration(nt, nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem)

	// test case 3:  the reconciler-manager should not revert the change to the fields that are not owned by reconciler-manager
	nt.T.Log("Manually update fields that are not owned by reconciler-manager")
	var modifiedTolerations []corev1.Toleration
	mustUpdateRootReconciler(nt, func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].TerminationMessagePath = "dev/termination-message"
		d.Spec.Template.Spec.Containers[0].Stdin = true
		d.Spec.Template.Spec.Tolerations = append(d.Spec.Template.Spec.Tolerations, corev1.Toleration{Key: "kubernetes.io/arch", Effect: "NoSchedule", Operator: "Exists"})
		modifiedTolerations = d.Spec.Template.Spec.Tolerations
		d.Spec.Template.Spec.PriorityClassName = "system-node-critical"
	})

	nt.T.Log("Verify the reconciler-manager does not revert the change")
	generation++ // generation bumped by 1 because reconicler-manager should not revert this change
	err = nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{
			testpredicates.HasGenerationAtLeast(generation),
			firstContainerTerminationMessagePathEquals("dev/termination-message"),
			firstContainerStdinEquals(true),
			hasTolerations(modifiedTolerations),
			hasPriorityClassName("system-node-critical"),
		})
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = getDeploymentGeneration(nt, nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem)
	// change the fields back to default values
	mustUpdateRootReconciler(nt, func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].TerminationMessagePath = "dev/termination-log"
		d.Spec.Template.Spec.Containers[0].Stdin = false
		d.Spec.Template.Spec.Tolerations = originalTolerations
		d.Spec.Template.Spec.PriorityClassName = ""
	})
	generation++ // generation bumped by 1 because reconciler-manager should not revert this change
	err = nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{
			testpredicates.HasGenerationAtLeast(generation),
			firstContainerTerminationMessagePathEquals("dev/termination-log"),
			firstContainerStdinEquals(false),
			hasTolerations(originalTolerations),
			hasPriorityClassName(""),
		})
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = getDeploymentGeneration(nt, nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem)

	// test case 4: the reconciler-manager should update the reconciler Deployment if the manifest in the ConfigMap has been changed.
	nt.T.Log("Update the Deployment manifest in the ConfigMap")
	nt.T.Cleanup(func() {
		resetReconcilerDeploymentManifests(nt, reconcilermanager.Reconciler, originalImagePullPolicy)
	})
	mustUpdateReconcilerTemplateConfigMap(nt, func(d *appsv1.Deployment) {
		for i, container := range d.Spec.Template.Spec.Containers {
			if container.Name == reconcilermanager.Reconciler {
				d.Spec.Template.Spec.Containers[i].ImagePullPolicy = updatedImagePullPolicy
			}
		}
	})
	nt.T.Log("Restart the reconciler-manager to pick up the manifests change")
	nomostest.DeletePodByLabel(nt, "app", reconcilermanager.ManagerName, true)
	nt.T.Log("Verify the reconciler Deployment has been updated to the new manifest")
	generation++ // generation bumped by 1 to apply the new change in the default manifests declared in the Config Map
	err = nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{
			testpredicates.HasGenerationAtLeast(generation),
			testpredicates.DeploymentContainerPullPolicyEquals(reconcilermanager.Reconciler, updatedImagePullPolicy),
		})
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = getDeploymentGeneration(nt, nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem)

	// test case 5: the reconciler-manager should delete the git-creds volume if not needed
	currentVolumesCount := len(reconcilerDeployment.Spec.Template.Spec.Volumes)
	rs := fake.RootSyncObjectV1Beta1(configsync.RootSyncName)
	nt.T.Log("Switch the auth type from ssh to none")
	nt.MustMergePatch(rs, `{"spec": {"git": {"auth": "none", "secretRef": {"name":""}}}}`)
	nt.T.Log("Verify the git-creds volume is gone")
	generation++ // generation bumped by 1 to delete the git-creds volume
	err = nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{testpredicates.HasGenerationAtLeast(generation), gitCredsVolumeDeleted(currentVolumesCount)})
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = getDeploymentGeneration(nt, nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem)

	// test case 6: the reconciler-manager should add the gcenode-askpass-sidecar container when needed
	nt.T.Log("Switch the auth type from none to gcpserviceaccount")
	nt.MustMergePatch(rs, `{"spec":{"git":{"auth":"gcpserviceaccount","secretRef":{"name":""},"gcpServiceAccountEmail":"test-gcp-sa-email@test-project.iam.gserviceaccount.com"}}}`)
	nt.T.Log("Verify the gcenode-askpass-sidecar container should be added")
	generation++ // generation bumped by 1 to apply the new sidecar container
	err = nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{testpredicates.HasGenerationAtLeast(generation), templateForGcpServiceAccountAuthType()})
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = getDeploymentGeneration(nt, nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem)

	// test case 7: the reconciler-manager should mount the git-creds volumes again if the auth type requires a git secret
	nt.T.Log("Switch the auth type gcpserviceaccount to ssh")
	nt.MustMergePatch(rs, `{"spec":{"git":{"auth":"ssh","secretRef":{"name":"git-creds"}}}}`)
	nt.T.Log("Verify the git-creds volume exists and the gcenode-askpass-sidecar container is gone")
	generation++ // generation bumped by 1 to add the git-cred volume again
	err = nt.Watcher.WatchObject(kinds.Deployment(), nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{testpredicates.HasGenerationAtLeast(generation), templateForSSHAuthType()})
	if err != nil {
		nt.T.Fatal(err)
	}
}

type updateFunc func(deployment *appsv1.Deployment)

func mustUpdateRootReconciler(nt *nomostest.NT, f updateFunc) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		d := &appsv1.Deployment{}
		if err := nt.KubeClient.Get(nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem, d); err != nil {
			return err
		}
		f(d)
		return nt.KubeClient.Update(d)
	})
	if err != nil {
		nt.T.Fatal(err)
	}
}

func mustUpdateReconcilerTemplateConfigMap(nt *nomostest.NT, f updateFunc) {
	decoder := serializer.NewCodecFactory(nt.Scheme).UniversalDeserializer()
	yamlSerializer := jserializer.NewYAMLSerializer(jserializer.DefaultMetaFactory, nt.Scheme, nt.Scheme)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get ConfigMap
		rmConfigMap := &corev1.ConfigMap{}
		if err := nt.KubeClient.Get(controllers.ReconcilerTemplateConfigMapName, configsync.ControllerNamespace, rmConfigMap); err != nil {
			return err
		}

		// Decode ConfigMap data entry as Deployment
		yamlString := rmConfigMap.Data[controllers.ReconcilerTemplateConfigMapKey]
		rmObj := &appsv1.Deployment{}
		if _, _, err := decoder.Decode([]byte(yamlString), nil, rmObj); err != nil {
			return err
		}

		// Mutate Deployment
		f(rmObj)

		// Encode Deployment and update ConfigMap data entry
		yamlBytes, err := runtime.Encode(yamlSerializer, rmObj)
		if err != nil {
			return err
		}
		rmConfigMap.Data[controllers.ReconcilerTemplateConfigMapKey] = string(yamlBytes)

		// Update ConfigMap
		return nt.KubeClient.Update(rmConfigMap)
	})
	if err != nil {
		nt.T.Fatal(err)
	}
}

func resetReconcilerDeploymentManifests(nt *nomostest.NT, containerName string, pullPolicy corev1.PullPolicy) {
	nt.T.Log("Reset the Deployment manifest in the ConfigMap")
	if err := nomostest.ResetReconcilerManagerConfigMap(nt); err != nil {
		nt.T.Fatalf("failed to reset configmap: %v", err)
	}

	nt.T.Log("Restart the reconciler-manager to pick up the manifests change")
	nomostest.DeletePodByLabel(nt, "app", reconcilermanager.ManagerName, true)

	nt.T.Log("Verify the reconciler Deployment has been reverted to the original manifest")
	err := nt.Watcher.WatchObject(kinds.Deployment(),
		nomostest.DefaultRootReconcilerName, v1.NSConfigManagementSystem,
		[]testpredicates.Predicate{
			testpredicates.DeploymentContainerPullPolicyEquals(containerName, pullPolicy),
		},
	)
	if err != nil {
		nt.T.Fatal(err)
	}
}

func firstContainerTerminationMessagePathEquals(terminationMessagePath string) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		if d.Spec.Template.Spec.Containers[0].TerminationMessagePath != terminationMessagePath {
			return fmt.Errorf("expected first container terminationMessagePath is: %s, got: %s", terminationMessagePath, d.Spec.Template.Spec.Containers[0].TerminationMessagePath)
		}
		return nil
	}
}

func firstContainerStdinEquals(stdin bool) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		if d.Spec.Template.Spec.Containers[0].Stdin != stdin {
			return fmt.Errorf("expected first container stdin is: %t, got: %t", stdin, d.Spec.Template.Spec.Containers[0].Stdin)
		}
		return nil
	}
}

func hasTolerations(tolerations []corev1.Toleration) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		for i, toleration := range d.Spec.Template.Spec.Tolerations {
			if !equality.Semantic.DeepEqual(toleration, tolerations[i]) {
				return fmt.Errorf("expected toleration is: %s, got: %s", tolerations[i].String(), toleration.String())
			}
		}
		return nil
	}
}

func hasPriorityClassName(priorityClassName string) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		if d.Spec.Template.Spec.PriorityClassName != priorityClassName {
			return fmt.Errorf("expected priorityClassName is: %s, got: %s", priorityClassName, d.Spec.Template.Spec.PriorityClassName)
		}
		return nil
	}
}

func hasReplicas(replicas int32) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		if *d.Spec.Replicas != replicas {
			return fmt.Errorf("expected replicas: %d, got: %d", replicas, *d.Spec.Replicas)
		}
		return nil
	}
}

func gitCredsVolumeDeleted(volumesCount int) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		if len(d.Spec.Template.Spec.Volumes) != volumesCount-1 {
			return fmt.Errorf("expected volumes count: %d, got: %d",
				volumesCount-1, len(d.Spec.Template.Spec.Volumes))
		}
		for _, volume := range d.Spec.Template.Spec.Volumes {
			if volume.Name == controllers.GitCredentialVolume {
				return fmt.Errorf("the git-creds volume should be gone for `none` auth type")
			}
		}
		return nil
	}
}

func templateForGcpServiceAccountAuthType() testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		for _, volume := range d.Spec.Template.Spec.Volumes {
			if volume.Name == controllers.GitCredentialVolume {
				return fmt.Errorf("the git-creds volume should not exist for `gcpserviceaccount` auth type")
			}
		}
		for _, container := range d.Spec.Template.Spec.Containers {
			if container.Name == reconcilermanager.GCENodeAskpassSidecar {
				return nil
			}
		}
		return fmt.Errorf("the %s container has not be created yet", reconcilermanager.GCENodeAskpassSidecar)
	}
}

func templateForSSHAuthType() testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		for _, container := range d.Spec.Template.Spec.Containers {
			if container.Name == reconcilermanager.GCENodeAskpassSidecar {
				return fmt.Errorf("the gcenode-askpass-sidecar container should not exist for `ssh` auth type")
			}
		}
		for _, volume := range d.Spec.Template.Spec.Volumes {
			if volume.Name == controllers.GitCredentialVolume {
				return nil
			}
		}
		return fmt.Errorf("the git-creds volume has not be created yet")
	}
}

func firstContainerNameEquals(expected string) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		d, ok := o.(*appsv1.Deployment)
		if !ok {
			return testpredicates.WrongTypeErr(d, &appsv1.Deployment{})
		}
		found := d.Spec.Template.Spec.Containers[0].Name
		if found != expected {
			return fmt.Errorf("expected name of the first container: %q, got: %q",
				expected, found)
		}
		return nil
	}
}

func totalExpectedContainerResources(resourceMap map[string]v1beta1.ContainerResourcesSpec) v1beta1.ContainerResourcesSpec {
	totals := v1beta1.ContainerResourcesSpec{
		CPURequest:    resource.MustParse("0m"),
		CPULimit:      resource.MustParse("0m"),
		MemoryRequest: resource.MustParse("0m"),
		MemoryLimit:   resource.MustParse("0m"),
	}
	for _, container := range resourceMap {
		totals.CPURequest.Add(container.CPURequest)
		totals.CPULimit.Add(container.CPULimit)
		totals.MemoryRequest.Add(container.MemoryRequest)
		totals.MemoryLimit.Add(container.MemoryLimit)
	}
	return totals
}

func TestAutopilotReconcilerAdjustment(t *testing.T) {
	nt := nomostest.New(t, nomostesting.ACMController, ntopts.Unstructured)

	rootSyncNN := nomostest.RootSyncNN(configsync.RootSyncName)
	reconcilerNN := core.RootReconcilerObjectKey(rootSyncNN.Name)

	// Get RootSync
	rootSyncObj := &v1beta1.RootSync{}
	err := nt.Validate(rootSyncNN.Name, rootSyncNN.Namespace, rootSyncObj,
		// Confirm there are no resource overrides
		testpredicates.RootSyncSpecOverrideEquals(
			&v1beta1.RootSyncOverrideSpec{
				OverrideSpec: v1beta1.OverrideSpec{
					ReconcileTimeout: &metav1.Duration{Duration: *nt.DefaultReconcileTimeout},
				},
			},
		),
	)
	if err != nil {
		nt.T.Fatal(err)
	}

	// Get reconciler Deployment
	reconcilerDeployment := &appsv1.Deployment{}
	if err := nt.Validate(reconcilerNN.Name, reconcilerNN.Namespace, reconcilerDeployment); err != nil {
		nt.T.Fatal(err)
	}
	firstContainerName := reconcilerDeployment.Spec.Template.Spec.Containers[0].Name
	generation := reconcilerDeployment.Generation

	// default container resource requests defined in code:
	// pkg/reconcilermanager/controllers/reconciler_container_resources.go
	var expectedResources map[string]v1beta1.ContainerResourcesSpec
	if nt.IsGKEAutopilot {
		expectedResources = controllers.ReconcilerContainerResourceDefaultsForAutopilot()
	} else {
		expectedResources = controllers.ReconcilerContainerResourceDefaults()
	}
	// Filter container map down to just expected containers
	expectedResources = filterResourceMap(expectedResources,
		reconcilermanager.Reconciler,
		reconcilermanager.GitSync,
		metrics.OtelAgentName)
	nt.T.Logf("expectedResources: %s", log.AsJSON(expectedResources))

	if _, found := expectedResources[firstContainerName]; !found {
		nt.T.Fatalf("expected the default resource map to include %q, but it was missing: %+v", firstContainerName, expectedResources)
	}

	if nt.IsGKEAutopilot {
		simulateAutopilotResourceAdjustment(nt, expectedResources, firstContainerName)
	}

	nt.T.Log("Validating container resources - 1")
	reconcilerDeployment = &appsv1.Deployment{}
	err = nt.Validate(reconcilerNN.Name, reconcilerNN.Namespace, reconcilerDeployment,
		testpredicates.HasGenerationAtLeast(generation),
		testpredicates.DeploymentContainerResourcesAllEqual(nt.Scheme, nt.Logger, expectedResources),
		firstContainerNameEquals(firstContainerName),
	)
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = reconcilerDeployment.GetGeneration()

	nt.T.Log("Increasing CPU and Memory request on the RootSync spec.override")
	updated := expectedResources[firstContainerName]
	updated.CPURequest.Add(resource.MustParse("10m"))
	updated.MemoryRequest.Add(resource.MustParse("10Mi"))
	nt.MustMergePatch(rootSyncObj,
		fmt.Sprintf(`{"spec":{"override":{"resources":[{"containerName":%q,"memoryRequest":%q, "cpuRequest":%q}]}}}`,
			firstContainerName, &updated.MemoryRequest, &updated.CPURequest))

	// Update expectations
	expectedResources[firstContainerName] = updated
	if nt.IsGKEAutopilot {
		simulateAutopilotResourceAdjustment(nt, expectedResources, firstContainerName)
	}

	// Wait for overrides to be applied
	// Note: This depends on the Syncing condition reflecting the current RSync generation.
	if err := nt.WatchForAllSyncs(); err != nil {
		nt.T.Fatal(err)
	}

	nt.T.Log("Wait for the reconciler deployment to be updated once")
	generation++ // patched by reconciler-manager
	err = nt.Watcher.WatchObject(kinds.Deployment(), reconcilerNN.Name, reconcilerNN.Namespace, []testpredicates.Predicate{
		testpredicates.HasGenerationAtLeast(generation),
	})
	if err != nil {
		nt.T.Fatal(err)
	}

	nt.T.Log("Verify the reconciler-manager applied the override memory/CPU request change")
	reconcilerDeployment = &appsv1.Deployment{}
	err = nt.Validate(reconcilerNN.Name, reconcilerNN.Namespace, reconcilerDeployment,
		testpredicates.HasGenerationAtLeast(generation),
		testpredicates.DeploymentContainerResourcesAllEqual(nt.Scheme, nt.Logger, expectedResources),
		firstContainerNameEquals(firstContainerName),
	)
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = reconcilerDeployment.GetGeneration()

	nt.T.Log("Increasing CPU and Memory request on the reconciler Deployment spec.template.spec.containers")
	updated = expectedResources[firstContainerName]
	updated.CPURequest.Add(resource.MustParse("10m"))
	updated.MemoryRequest.Add(resource.MustParse("10Mi"))
	mustUpdateRootReconciler(nt, func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    updated.CPURequest,
			corev1.ResourceMemory: updated.MemoryRequest,
		}
	})

	// Don't update expectations.
	// We expect the manual changes to be reverted.

	nt.T.Log("Wait for the reconciler deployment to be updated twice")
	generation += 2 // manual update + reconciler-manager revert
	err = nt.Watcher.WatchObject(kinds.Deployment(), reconcilerNN.Name, reconcilerNN.Namespace, []testpredicates.Predicate{
		testpredicates.HasGenerationAtLeast(generation),
	})
	if err != nil {
		nt.T.Fatal(err)
	}

	nt.T.Log("Verify the reconciler-manager reverted the manual memory/CPU request change")
	reconcilerDeployment = &appsv1.Deployment{}
	err = nt.Validate(reconcilerNN.Name, reconcilerNN.Namespace, reconcilerDeployment,
		testpredicates.HasGenerationAtLeast(generation),
		testpredicates.DeploymentContainerResourcesAllEqual(nt.Scheme, nt.Logger, expectedResources),
		firstContainerNameEquals(firstContainerName),
	)
	if err != nil {
		nt.T.Fatal(err)
	}
	generation = reconcilerDeployment.GetGeneration()

	nt.T.Log("Decreasing CPU request on the RootSync spec.override")
	updated = expectedResources[firstContainerName]
	// Reduce CPU, but not by enough to change the value when rounded up.
	updated.CPURequest.Sub(resource.MustParse("10m"))
	expectedResources[firstContainerName] = updated
	if nt.IsGKEAutopilot {
		simulateAutopilotResourceAdjustment(nt, expectedResources, firstContainerName)
		// Expect reconciler-manager NOT to update the reconciler Deployment (no change after adjustment)
	} else {
		// Expect reconciler-manager to update the reconciler Deployment
		generation++
	}

	nt.MustMergePatch(rootSyncObj,
		fmt.Sprintf(`{"spec":{"override":{"resources":[{"containerName":%q,"memoryRequest":%q, "cpuRequest":%q}]}}}`,
			firstContainerName, &updated.MemoryRequest, &updated.CPURequest))

	// Wait for overrides to be applied
	// Note: This depends on the Syncing condition reflecting the current RSync generation.
	if err := nt.WatchForAllSyncs(); err != nil {
		nt.T.Fatal(err)
	}

	nt.T.Log("Wait for the reconciler-manager to update the reconciler deployment CPU request, only on non-autopilot cluster")
	err = nt.Watcher.WatchObject(kinds.Deployment(), reconcilerNN.Name, reconcilerNN.Namespace, []testpredicates.Predicate{
		testpredicates.HasGenerationAtLeast(generation),
	})
	if err != nil {
		nt.T.Fatal(err)
	}

	nt.T.Log("Verify the reconciler-manager changed the reconciler CPU request, only on non-autopilot cluster")
	reconcilerDeployment = &appsv1.Deployment{}
	err = nt.Validate(reconcilerNN.Name, reconcilerNN.Namespace, reconcilerDeployment,
		testpredicates.HasGenerationAtLeast(generation),
		testpredicates.DeploymentContainerResourcesAllEqual(nt.Scheme, nt.Logger, expectedResources),
		firstContainerNameEquals(firstContainerName),
	)
	if err != nil {
		nt.T.Fatal(err)
	}
}

func simulateAutopilotResourceAdjustment(nt *nomostest.NT, expectedResources map[string]v1beta1.ContainerResourcesSpec, firstContainerName string) {
	// Compute expected totals
	expectedTotalResources := totalExpectedContainerResources(expectedResources)
	nt.T.Logf("expectedTotalResources: %s", log.AsJSON(expectedTotalResources))

	// https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-resource-requests#autopilot-resource-management

	// Autopilot increases the CPU of the first container,
	// until the total CPU is a multiple of 250m.
	minimumTotalCPURequests := resource.MustParse("250m")
	remainder := expectedTotalResources.CPURequest.MilliValue() % minimumTotalCPURequests.MilliValue()
	if remainder > 0 {
		// Compute difference
		diff := minimumTotalCPURequests.DeepCopy()
		diff.Sub(*resource.NewMilliQuantity(remainder, minimumTotalCPURequests.Format))
		// Add difference to first container
		// Go doesn't allow modifying a struct field in a map directly,
		// so read, update, and write it back.
		updated := expectedResources[firstContainerName]
		updated.CPURequest.Add(diff)
		expectedResources[firstContainerName] = updated
	}

	// Re-compute expected totals
	expectedTotalResources = totalExpectedContainerResources(expectedResources)

	// Autopilot increases the Memory of the first container,
	// until the total Memory is at least 1CPU:1Gi ratio (1000m:1024Mi).
	// Note: This math assumes the values are too low to overflow.
	minimumTotalMemory := int64(float64(expectedTotalResources.CPURequest.MilliValue()) / float64(1000.0) * 1024)
	// TODO: Figure out how to build a Quantity in Mebibytes without parsing
	minimumTotalMemoryRequests := resource.MustParse(fmt.Sprintf("%dMi", minimumTotalMemory))
	if expectedTotalResources.MemoryRequest.Cmp(minimumTotalMemoryRequests) < 0 {
		// Compute difference
		diff := minimumTotalMemoryRequests.DeepCopy()
		diff.Sub(expectedTotalResources.MemoryRequest)
		// Add difference to first container
		// Go doesn't allow modifying a struct field in a map directly,
		// so read, update, and write it back.
		updated := expectedResources[firstContainerName]
		updated.MemoryRequest.Add(diff)
		expectedResources[firstContainerName] = updated
	}

	// Autopilot sets Limits to Requests
	setLimitsToRequests(expectedResources)
	nt.T.Logf("expectedResources (adjusted for autopilot): %s", log.AsJSON(expectedResources))
	expectedTotalResources = totalExpectedContainerResources(expectedResources)
	nt.T.Logf("expectedTotalResources (adjusted for autopilot): %s", log.AsJSON(expectedTotalResources))
}

func setLimitsToRequests(resourceMap map[string]v1beta1.ContainerResourcesSpec) {
	for containerName := range resourceMap {
		updated := resourceMap[containerName]
		updated.CPULimit = updated.CPURequest
		updated.MemoryLimit = updated.MemoryRequest
		resourceMap[containerName] = updated
	}
}

func getDeploymentGeneration(nt *nomostest.NT, name, namespace string) int64 {
	dep := &appsv1.Deployment{}
	if err := nt.KubeClient.Get(name, namespace, dep); err != nil {
		nt.T.Fatal(err)
	}
	return dep.GetGeneration()
}

func filterResourceMap(resourceMap map[string]v1beta1.ContainerResourcesSpec, containers ...string) map[string]v1beta1.ContainerResourcesSpec {
	filteredMap := make(map[string]v1beta1.ContainerResourcesSpec, len(containers))
	for _, name := range containers {
		if resources, found := resourceMap[name]; found {
			filteredMap[name] = resources
		}
	}
	return filteredMap
}
