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

package nomostest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"kpt.dev/configsync/e2e"
	"kpt.dev/configsync/e2e/nomostest/gitproviders"
	"kpt.dev/configsync/e2e/nomostest/metrics"
	"kpt.dev/configsync/e2e/nomostest/ntopts"
	"kpt.dev/configsync/e2e/nomostest/taskgroup"
	"kpt.dev/configsync/e2e/nomostest/testing"
	"kpt.dev/configsync/e2e/nomostest/testpredicates"
	"kpt.dev/configsync/e2e/nomostest/testwatcher"
	"kpt.dev/configsync/pkg/api/configmanagement"
	"kpt.dev/configsync/pkg/api/configsync"
	"kpt.dev/configsync/pkg/api/configsync/v1alpha1"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	"kpt.dev/configsync/pkg/applier"
	"kpt.dev/configsync/pkg/core"
	"kpt.dev/configsync/pkg/importer/filesystem"
	"kpt.dev/configsync/pkg/importer/filesystem/cmpath"
	"kpt.dev/configsync/pkg/importer/reader"
	"kpt.dev/configsync/pkg/kinds"
	"kpt.dev/configsync/pkg/metadata"
	ocmetrics "kpt.dev/configsync/pkg/metrics"
	"kpt.dev/configsync/pkg/reconcilermanager"
	"kpt.dev/configsync/pkg/reconcilermanager/controllers"
	"kpt.dev/configsync/pkg/testing/fake"
	webhookconfig "kpt.dev/configsync/pkg/webhook/configuration"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object/dependson"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// configSyncManifests is the folder of the test manifests
	configSyncManifests = "manifests"

	// e2e/raw-nomos/manifests/multi-repo-configmaps.yaml
	multiConfigMapsName = "multi-repo-configmaps.yaml"

	// clusterRoleName imitates a user-created (Cluster)Role for NS reconcilers.
	clusterRoleName = "cs-e2e"
)

var (
	// baseDir is the path to the Nomos repository root from test case files.
	//
	// All paths must be relative to the test file that is running. There is probably
	// a more elegant way to do this.
	baseDir            = filepath.FromSlash("../..")
	outputManifestsDir = filepath.Join(baseDir, ".output", "staging", "oss")
	configSyncManifest = filepath.Join(outputManifestsDir, "config-sync-manifest.yaml")
	multiConfigMaps    = filepath.Join(baseDir, "e2e", "raw-nomos", configSyncManifests, multiConfigMapsName)
)

var (
	// reconcilerPollingPeriod specifies reconciler polling period as time.Duration
	reconcilerPollingPeriod time.Duration
	// hydrationPollingPeriod specifies hydration-controller polling period as time.Duration
	hydrationPollingPeriod time.Duration
)

// IsReconcilerManagerConfigMap returns true if passed obj is the
// reconciler-manager ConfigMap reconciler-manager-cm in config-management namespace.
func IsReconcilerManagerConfigMap(obj client.Object) bool {
	return obj.GetName() == "reconciler-manager-cm" &&
		obj.GetNamespace() == "config-management-system" &&
		obj.GetObjectKind().GroupVersionKind() == kinds.ConfigMap()
}

// isOtelCollectorDeployment returns true if passed obj is the
// otel-collector Deployment in the config-management-monitoring namespace.
func isOtelCollectorDeployment(obj client.Object) bool {
	return obj.GetName() == ocmetrics.OtelCollectorName &&
		obj.GetNamespace() == ocmetrics.MonitoringNamespace &&
		obj.GetObjectKind().GroupVersionKind() == kinds.Deployment()
}

// ResetReconcilerManagerConfigMap resets the reconciler manager config map
// to what is defined in the manifest
func ResetReconcilerManagerConfigMap(nt *NT) error {
	objs, err := parseConfigSyncManifests(nt)
	if err != nil {
		return err
	}
	for _, obj := range objs {
		if !IsReconcilerManagerConfigMap(obj) {
			continue
		}
		nt.T.Logf("ResetReconcilerManagerConfigMap obj: %v", core.GKNN(obj))
		err := nt.KubeClient.Update(obj)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("failed to reset reconciler manager ConfigMap")
}

func parseConfigSyncManifests(nt *NT) ([]client.Object, error) {
	tmpManifestsDir := filepath.Join(nt.TmpDir, configSyncManifests)
	objs, err := installationManifests(tmpManifestsDir)
	if err != nil {
		return nil, err
	}
	objs, err = convertToTypedObjects(nt, objs)
	if err != nil {
		return nil, err
	}
	reconcilerPollingPeriod = nt.ReconcilerPollingPeriod
	hydrationPollingPeriod = nt.HydrationPollingPeriod
	objs, err = multiRepoObjects(objs,
		setReconcilerDebugMode,
		setPollingPeriods,
		setOtelCollectorPrometheusAnnotations)
	if err != nil {
		return nil, err
	}
	return objs, nil
}

// installConfigSync installs ConfigSync on the test cluster
func installConfigSync(nt *NT, nomos ntopts.Nomos) error {
	nt.T.Log("[SETUP] Installing Config Sync")
	objs, err := parseConfigSyncManifests(nt)
	if err != nil {
		return err
	}
	for _, o := range objs {
		nt.T.Logf("installConfigSync obj: %v", core.GKNN(o))
		if o.GetObjectKind().GroupVersionKind().GroupKind() == kinds.ConfigMap().GroupKind() && o.GetName() == reconcilermanager.SourceFormat {
			cm := o.(*corev1.ConfigMap)
			cm.Data[filesystem.SourceFormatKey] = string(nomos.SourceFormat)
		}
		if err := nt.KubeClient.Apply(o); err != nil {
			return err
		}
	}
	return nil
}

// uninstallConfigSync uninstalls ConfigSync on the test cluster
func uninstallConfigSync(nt *NT) error {
	nt.T.Log("[CLEANUP] Uninstalling Config Sync")
	objs, err := parseConfigSyncManifests(nt)
	if err != nil {
		return err
	}
	for _, o := range objs {
		if err := nt.KubeClient.Delete(o); err != nil {
			if !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
				return err
			}
		}
	}
	return nil
}

func isPSPCluster() bool {
	return strings.Contains(testing.GCPClusterFromEnv, "psp")
}

// convertToTypedObjects converts objects to their literal types. We can do this as
// we should have all required types in the Scheme anyway. This keeps us from
// having to do ugly Unstructured operations.
func convertToTypedObjects(nt *NT, objs []client.Object) ([]client.Object, error) {
	result := make([]client.Object, len(objs))
	for i, obj := range objs {
		uObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			// Already converted when read from the disk, or added manually.
			// We don't need to convert, so insert and go to the next element.
			result[i] = obj
			continue
		}
		rObj, err := kinds.ToTypedObject(uObj, nt.Scheme)
		if err != nil {
			return nil, err
		}
		cObj, ok := rObj.(client.Object)
		if !ok {
			return nil, errors.Errorf(
				"failed to cast object %T of kind %s to client.Object",
				rObj, uObj.GroupVersionKind().Kind)
		}
		result[i] = cObj
	}
	return result, nil
}

func copyFile(src, dst string) error {
	bytes, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, bytes, fileMode)
}

func copyDirContents(src, dest string) error {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, f := range files {
		// Explicitly not recursive
		if f.IsDir() {
			continue
		}

		from := filepath.Join(src, f.Name())
		to := filepath.Join(dest, f.Name())
		err := copyFile(from, to)
		if err != nil {
			return err
		}
	}
	return nil
}

// installationManifests generates the ConfigSync installation YAML and copies
// it to the test's temporary directory.
func installationManifests(tmpManifestsDir string) ([]client.Object, error) {
	err := os.MkdirAll(tmpManifestsDir, fileMode)
	if err != nil {
		return nil, err
	}

	// Copy all manifests to temporary dir
	if err := copyDirContents(outputManifestsDir, tmpManifestsDir); err != nil {
		return nil, err
	}
	if err := copyFile(multiConfigMaps, filepath.Join(tmpManifestsDir, multiConfigMapsName)); err != nil {
		return nil, err
	}

	return parseManifestDir(tmpManifestsDir)
}

// parseManifestDir reads YAML from all the files in a directory and parses them
// into unstructured objects.
func parseManifestDir(dirPath string) ([]client.Object, error) {
	// Create the list of paths for the File to read.
	readPath, err := cmpath.AbsoluteOS(dirPath)
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	paths := make([]cmpath.Absolute, len(files))
	for i, f := range files {
		paths[i] = readPath.Join(cmpath.RelativeSlash(f.Name()))
	}
	// Read the manifests cached in the tmpdir.
	r := reader.File{}
	filePaths := reader.FilePaths{
		RootDir: readPath,
		Files:   paths,
	}
	fos, err := r.Read(filePaths)
	if err != nil {
		return nil, err
	}

	var objs []client.Object
	for _, o := range fos {
		objs = append(objs, o.Unstructured)
	}
	return objs, nil
}

func multiRepoObjects(objects []client.Object, opts ...func(obj client.Object) error) ([]client.Object, error) {
	var filtered []client.Object
	found := false
	for _, obj := range objects {
		if !isPSPCluster() && obj.GetName() == "acm-psp" {
			continue
		}
		if IsReconcilerManagerConfigMap(obj) {
			// Mark that we've found the ReconcilerManager ConfigMap.
			// This way we know we've enabled debug mode.
			found = true
		}
		for _, opt := range opts {
			if err := opt(obj); err != nil {
				return nil, err
			}
		}
		filtered = append(filtered, obj)
	}
	if !found {
		return nil, fmt.Errorf("Reconciler Manager ConfigMap not found")
	}
	return filtered, nil
}

// ValidateMultiRepoDeployments validates if all Config Sync Components are available.
func ValidateMultiRepoDeployments(nt *NT) error {
	rs := RootSyncObjectV1Beta1FromRootRepo(nt, configsync.RootSyncName)
	if err := nt.KubeClient.Create(rs); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			nt.T.Fatal(err)
		}
	}

	tg := taskgroup.New()
	tg.Go(func() error {
		return nt.Watcher.WatchForCurrentStatus(kinds.Deployment(),
			configmanagement.RGControllerName, configmanagement.RGControllerNamespace)
	})
	tg.Go(func() error {
		return nt.Watcher.WatchForCurrentStatus(kinds.Deployment(),
			reconcilermanager.ManagerName, configmanagement.ControllerNamespace)
	})
	tg.Go(func() error {
		predicates := []testpredicates.Predicate{testpredicates.StatusEquals(nt.Scheme, kstatus.CurrentStatus)}
		// If testing on GKE, ensure that the otel-collector deployment has
		// picked up the `otel-collector-googlecloud` configmap.
		// This bumps the deployment generation.
		if *e2e.TestCluster == e2e.GKE {
			predicates = append(predicates, testpredicates.HasGenerationAtLeast(2))
		}
		return nt.Watcher.WatchObject(kinds.Deployment(),
			ocmetrics.OtelCollectorName, ocmetrics.MonitoringNamespace, predicates)
	})
	tg.Go(func() error {
		// The root-reconciler is created after the reconciler-manager is ready.
		// So wait longer for it to come up.
		return nt.Watcher.WatchForCurrentStatus(kinds.Deployment(),
			DefaultRootReconcilerName, configmanagement.ControllerNamespace,
			testwatcher.WatchTimeout(nt.DefaultWaitTimeout*2))
	})
	if !*nt.WebhookDisabled {
		// If disabled, we don't need to wait for the webhook to be ready.
		tg.Go(func() error {
			return nt.Watcher.WatchForCurrentStatus(kinds.ValidatingWebhookConfiguration(),
				"admission-webhook.configsync.gke.io", "")
		})
		tg.Go(func() error {
			// The webhook takes a long time to come up, because of SSL cert
			// injection. So wait longer for it to come up.
			return nt.Watcher.WatchForCurrentStatus(kinds.Deployment(),
				webhookconfig.ShortName, configmanagement.ControllerNamespace,
				testwatcher.WatchTimeout(nt.DefaultWaitTimeout*2))
		})
	}
	if err := tg.Wait(); err != nil {
		return err
	}

	return validateMultiRepoPods(nt)
}

// validateMultiRepoPods validates if all Config Sync Pods are healthy. It doesn't retry
// because it is supposed to be invoked after validateMultiRepoDeployments succeeds.
// It only checks the central root reconciler Pod.
func validateMultiRepoPods(nt *NT) error {
	for _, ns := range CSNamespaces {
		pods := corev1.PodList{}
		if err := nt.KubeClient.List(&pods, client.InNamespace(ns)); err != nil {
			return err
		}
		for _, pod := range pods.Items {
			if pod.Labels["app"] == "reconciler" {
				if pod.Labels[metadata.SyncKindLabel] != configsync.RootSyncKind || pod.Labels[metadata.SyncNameLabel] != configsync.RootSyncName {
					continue // Only validate the central root reconciler Pod but skip other reconcilers because other RSync objects might be managed by the central root reconciler and may not be reconciled yet.
				}
			}
			// Check if the Pod (running Pod only) has any former abnormally terminated container.
			// If the Pod is already in a terminating state (with a deletion timestamp) due to clean up, ignore the check.
			// It uses pod.DeletionTimestamp instead of pod.Status.Phase because the Pod may not be scheduled to evict yet.
			if pod.DeletionTimestamp == nil {
				if err := nt.Validate(pod.Name, pod.Namespace, &corev1.Pod{}, noOOMKilledContainer(nt)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// noOOMKilledContainer is a predicate to validate if the Pod has a container that was terminated due to OOMKilled
// It doesn't check other non-zero exit code because the admission-webhook sometimes exits with 1 and then recovers after restart.
// Example log:
//
//	kubectl logs admission-webhook-5c79b59f86-fzqgl -n config-management-system -p
//	I0826 07:54:47.435824       1 setup.go:31] Build Version: v1.13.0-rc.5-2-gef2d1ac-dirty
//	I0826 07:54:47.435992       1 logr.go:249] setup "msg"="starting manager"
//	E0826 07:55:17.436921       1 logr.go:265]  "msg"="Failed to get API Group-Resources" "error"="Get \"https://10.96.0.1:443/api?timeout=32s\": dial tcp 10.96.0.1:443: i/o timeout"
//	E0826 07:55:17.436945       1 logr.go:265] setup "msg"="starting manager" "error"="Get \"https://10.96.0.1:443/api?timeout=32s\": dial tcp 10.96.0.1:443: i/o timeout"
func noOOMKilledContainer(nt *NT) testpredicates.Predicate {
	return func(o client.Object) error {
		if o == nil {
			return testpredicates.ErrObjectNotFound
		}
		pod, ok := o.(*corev1.Pod)
		if !ok {
			return testpredicates.WrongTypeErr(o, pod)
		}

		for _, cs := range pod.Status.ContainerStatuses {
			terminated := cs.LastTerminationState.Terminated
			if terminated != nil && terminated.ExitCode != 0 && terminated.Reason == "OOMKilled" {
				// Display the full state of the malfunctioning Pod to aid in debugging.
				jsn, err := json.MarshalIndent(pod, "", "  ")
				if err != nil {
					return err
				}
				// Display the logs for the previous instance of the container.
				args := []string{"logs", pod.Name, "-n", pod.Namespace, "-p"}
				out, err := nt.Shell.Kubectl(args...)
				// Print a standardized header before each printed log to make ctrl+F-ing the
				// log you want easier.
				cmd := fmt.Sprintf("kubectl %s", strings.Join(args, " "))
				if err != nil {
					nt.T.Logf("failed to run %q: %v\n%s", cmd, err, out)
				}
				return fmt.Errorf("%w for pod/%s in namespace %q, container %q terminated with exit code %d and reason %q\n\n%s\n\n%s",
					testpredicates.ErrFailedPredicate, pod.Name, pod.Namespace, cs.Name, terminated.ExitCode, terminated.Reason, string(jsn), fmt.Sprintf("%s\n%s", cmd, out))
			}
		}
		return nil
	}
}

func setupRootSync(nt *NT, rsName string, reconcileTimeout *time.Duration) {
	// create RootSync to initialize the root reconciler.
	rs := RootSyncObjectV1Beta1FromRootRepo(nt, rsName)
	if reconcileTimeout != nil {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(*reconcileTimeout)
	}
	if err := nt.KubeClient.Apply(rs); err != nil {
		nt.T.Fatal(err)
	}
}

func setupRepoSync(nt *NT, nn types.NamespacedName, reconcileTimeout *time.Duration) {
	// create RepoSync to initialize the Namespace reconciler.
	rs := RepoSyncObjectV1Beta1FromNonRootRepo(nt, nn)
	if reconcileTimeout != nil {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(*reconcileTimeout)
	}
	if err := nt.KubeClient.Create(rs); err != nil {
		nt.T.Fatal(err)
	}
}

// RepoSyncRoleBinding returns rolebinding that grants service account
// permission to manage resources in the namespace.
func RepoSyncRoleBinding(nn types.NamespacedName) *rbacv1.RoleBinding {
	rb := fake.RoleBindingObject(core.Name(nn.Name), core.Namespace(nn.Namespace))
	sb := []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      core.NsReconcilerName(nn.Namespace, nn.Name),
			Namespace: configmanagement.ControllerNamespace,
		},
	}
	rf := rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     clusterRoleName,
	}
	rb.Subjects = sb
	rb.RoleRef = rf
	return rb
}

// repoSyncClusterRoleBinding returns clusterrolebinding that grants service account
// permission to manage resources in the namespace.
func repoSyncClusterRoleBinding(nn types.NamespacedName) *rbacv1.ClusterRoleBinding {
	rb := fake.ClusterRoleBindingObject(core.Name(nn.Name + "-" + nn.Namespace))
	sb := []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      core.NsReconcilerName(nn.Namespace, nn.Name),
			Namespace: configmanagement.ControllerNamespace,
		},
	}
	rf := rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     clusterRoleName,
	}
	rb.Subjects = sb
	rb.RoleRef = rf
	return rb
}

func setupRepoSyncRoleBinding(nt *NT, nn types.NamespacedName) error {
	if err := nt.KubeClient.Create(RepoSyncRoleBinding(nn)); err != nil {
		nt.T.Fatal(err)
	}

	// Validate rolebinding is present.
	return nt.Validate(nn.Name, nn.Namespace, &rbacv1.RoleBinding{})
}

// setReconcilerDebugMode ensures the Reconciler deployments are run in debug mode.
func setReconcilerDebugMode(obj client.Object) error {
	if obj == nil {
		return testpredicates.ErrObjectNotFound
	}
	if !IsReconcilerManagerConfigMap(obj) {
		return nil
	}

	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return fmt.Errorf("parsed Reconciler Manager ConfigMap was %T %v", obj, obj)
	}

	key := "deployment.yaml"
	deploymentYAML, found := cm.Data[key]
	if !found {
		return fmt.Errorf("Reconciler Manager ConfigMap has no deployment.yaml entry")
	}

	// The Deployment YAML for Reconciler deployments is a raw YAML string embedded
	// in the ConfigMap. Unmarshalling/marshalling is likely to lead to errors, so
	// this modifies the YAML string directly by finding the line we want to insert
	// the debug flag after, and then inserting the line we want to add.
	lines := strings.Split(deploymentYAML, "\n")
	found = false
	for i, line := range lines {
		// We want to set the debug flag immediately after setting the source-dir flag.
		if strings.Contains(line, "- \"--source-dir=/repo/source/rev\"") {
			// Standard Go "insert into slice" idiom.
			lines = append(lines, "")
			copy(lines[i+2:], lines[i+1:])
			// Prefix of 8 spaces as the run arguments are indented 8 spaces relative
			// to the embedded YAML string. The embedded YAML is indented 3 spaces,
			// so this is equivalent to indenting 11 spaces in the original file:
			// manifests/templates/reconciler-manager-configmap.yaml.
			lines[i+1] = "        - \"--debug\""
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Unable to set debug mode for reconciler")
	}

	cm.Data[key] = strings.Join(lines, "\n")
	return nil
}

// setPollingPeriods update Reconciler Manager configmap
// reconciler-manager-cm with reconciler and hydration-controller polling
// periods to override the default.
func setPollingPeriods(obj client.Object) error {
	if obj == nil {
		return testpredicates.ErrObjectNotFound
	}
	if !IsReconcilerManagerConfigMap(obj) {
		return nil
	}

	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return fmt.Errorf("parsed Reconciler Manager ConfigMap was not ConfigMap %T %v", obj, obj)
	}

	cm.Data[reconcilermanager.ReconcilerPollingPeriod] = reconcilerPollingPeriod.String()
	cm.Data[reconcilermanager.HydrationPollingPeriod] = hydrationPollingPeriod.String()
	return nil
}

// setOtelCollectorPrometheusAnnotations updates the otel-collector Deployment
// to add pod annotations to enable metrics scraping.
func setOtelCollectorPrometheusAnnotations(obj client.Object) error {
	if obj == nil {
		return testpredicates.ErrObjectNotFound
	}
	if !isOtelCollectorDeployment(obj) {
		return nil
	}
	dep, ok := obj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("failed to cast %T to *appsv1.Deployment", obj)
	}
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = make(map[string]string, 2)
	}
	dep.Spec.Template.Annotations["prometheus.io/scrape"] = "true"
	dep.Spec.Template.Annotations["prometheus.io/port"] = strconv.Itoa(metrics.OtelCollectorMetricsPort)
	return nil
}

func setupDelegatedControl(nt *NT, reconcileTimeout *time.Duration) {
	nt.T.Log("[SETUP] Delegated control")

	// Just create one RepoSync ClusterRole, even if there are no Namespace repos.
	if err := nt.KubeClient.Create(nt.RepoSyncClusterRole()); err != nil {
		nt.T.Fatal(err)
	}

	for rsName := range nt.RootRepos {
		// The default RootSync is created manually, no need to set up again.
		if rsName == configsync.RootSyncName {
			continue
		}
		setupRootSync(nt, rsName, reconcileTimeout)
	}

	for nn := range nt.NonRootRepos {
		// Add a ClusterRoleBinding so that the pods can be created
		// when the cluster has PodSecurityPolicy enabled.
		// Background: If a RoleBinding (not a ClusterRoleBinding) is used,
		// it will only grant usage for pods being run in the same namespace as the binding.
		// TODO: Remove the psp related change when Kubernetes 1.25 is
		// available on GKE.
		if isPSPCluster() {
			if err := nt.KubeClient.Create(repoSyncClusterRoleBinding(nn)); err != nil {
				nt.T.Fatal(err)
			}
		}

		// create namespace for namespace reconciler.
		err := nt.KubeClient.Create(fake.NamespaceObject(nn.Namespace))
		if err != nil {
			nt.T.Fatal(err)
		}

		// create secret for the namespace reconciler.
		CreateNamespaceSecret(nt, nn.Namespace)

		if err := setupRepoSyncRoleBinding(nt, nn); err != nil {
			nt.T.Fatal(err)
		}

		setupRepoSync(nt, nn, reconcileTimeout)
	}
}

// RootSyncObjectV1Alpha1 returns the default RootSync object.
func RootSyncObjectV1Alpha1(name, repoURL string, sourceFormat filesystem.SourceFormat) *v1alpha1.RootSync {
	rs := fake.RootSyncObjectV1Alpha1(name)
	rs.Spec.SourceFormat = string(sourceFormat)
	rs.Spec.SourceType = string(v1beta1.GitSource)
	rs.Spec.Git = &v1alpha1.Git{
		Repo:   repoURL,
		Branch: gitproviders.MainBranch,
		Dir:    gitproviders.DefaultSyncDir,
		Auth:   "ssh",
		SecretRef: &v1alpha1.SecretReference{
			Name: controllers.GitCredentialVolume,
		},
	}
	// Enable automatic deletion of managed objects by default.
	// This helps ensure that test artifacts are cleaned up.
	EnableDeletionPropagation(rs)
	return rs
}

// RootSyncObjectV1Alpha1FromRootRepo returns a v1alpha1 RootSync object which
// uses a repo from nt.RootRepos.
func RootSyncObjectV1Alpha1FromRootRepo(nt *NT, name string) *v1alpha1.RootSync {
	repo, found := nt.RootRepos[name]
	if !found {
		nt.T.Fatal("nonexistent root repo: %s", name)
	}
	repoURL := nt.GitProvider.SyncURL(repo.RemoteRepoName)
	sourceFormat := repo.Format
	rs := RootSyncObjectV1Alpha1(name, repoURL, sourceFormat)
	if nt.DefaultReconcileTimeout != 0 {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(nt.DefaultReconcileTimeout)
	}
	return rs
}

// RootSyncObjectV1Beta1 returns the default RootSync object with version v1beta1.
func RootSyncObjectV1Beta1(name, repoURL string, sourceFormat filesystem.SourceFormat) *v1beta1.RootSync {
	rs := fake.RootSyncObjectV1Beta1(name)
	rs.Spec.SourceFormat = string(sourceFormat)
	rs.Spec.SourceType = string(v1beta1.GitSource)
	rs.Spec.Git = &v1beta1.Git{
		Repo:   repoURL,
		Branch: gitproviders.MainBranch,
		Dir:    gitproviders.DefaultSyncDir,
		Auth:   "ssh",
		SecretRef: &v1beta1.SecretReference{
			Name: controllers.GitCredentialVolume,
		},
	}
	// Enable automatic deletion of managed objects by default.
	// This helps ensure that test artifacts are cleaned up.
	EnableDeletionPropagation(rs)
	return rs
}

// RootSyncObjectV1Beta1FromRootRepo returns a v1beta1 RootSync object which
// uses a repo from nt.RootRepos.
func RootSyncObjectV1Beta1FromRootRepo(nt *NT, name string) *v1beta1.RootSync {
	repo, found := nt.RootRepos[name]
	if !found {
		nt.T.Fatal("nonexistent root repo: %s", name)
	}
	repoURL := nt.GitProvider.SyncURL(repo.RemoteRepoName)
	sourceFormat := repo.Format
	rs := RootSyncObjectV1Beta1(name, repoURL, sourceFormat)
	if nt.DefaultReconcileTimeout != 0 {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(nt.DefaultReconcileTimeout)
	}
	return rs
}

// RootSyncObjectV1Beta1FromOtherRootRepo returns a v1beta1 RootSync object which
// uses a repo from a specific nt.RootRepo.
func RootSyncObjectV1Beta1FromOtherRootRepo(nt *NT, syncName, repoName string) *v1beta1.RootSync {
	repo, found := nt.RootRepos[repoName]
	if !found {
		nt.T.Fatal("nonexistent root repo: %s", repoName)
	}
	repoURL := nt.GitProvider.SyncURL(repo.RemoteRepoName)
	sourceFormat := repo.Format
	rs := RootSyncObjectV1Beta1(syncName, repoURL, sourceFormat)
	if nt.DefaultReconcileTimeout != 0 {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(nt.DefaultReconcileTimeout)
	}
	return rs
}

// StructuredNSPath returns structured path with namespace and resourcename in repo.
func StructuredNSPath(namespace, resourceName string) string {
	return fmt.Sprintf("acme/namespaces/%s/%s.yaml", namespace, resourceName)
}

// RepoSyncObjectV1Alpha1 returns the default RepoSync object in the given namespace.
// SourceFormat for RepoSync must be Unstructured (default), so it's left unspecified.
func RepoSyncObjectV1Alpha1(nn types.NamespacedName, repoURL string) *v1alpha1.RepoSync {
	rs := fake.RepoSyncObjectV1Alpha1(nn.Namespace, nn.Name)
	rs.Spec.SourceType = string(v1beta1.GitSource)
	rs.Spec.Git = &v1alpha1.Git{
		Repo:   repoURL,
		Branch: gitproviders.MainBranch,
		Dir:    gitproviders.DefaultSyncDir,
		Auth:   "ssh",
		SecretRef: &v1alpha1.SecretReference{
			Name: "ssh-key",
		},
	}
	// Enable automatic deletion of managed objects by default.
	// This helps ensure that test artifacts are cleaned up.
	EnableDeletionPropagation(rs)
	return rs
}

// RepoSyncObjectV1Alpha1FromNonRootRepo returns a v1alpha1 RepoSync object which
// uses a repo from nt.NonRootRepos.
func RepoSyncObjectV1Alpha1FromNonRootRepo(nt *NT, nn types.NamespacedName) *v1alpha1.RepoSync {
	repo, found := nt.NonRootRepos[nn]
	if !found {
		nt.T.Fatal("nonexistent non-root repo: %s", nn)
	}
	repoURL := nt.GitProvider.SyncURL(repo.RemoteRepoName)
	// RepoSync is always Unstructured. So ignore repo.Format.
	rs := RepoSyncObjectV1Alpha1(nn, repoURL)
	if nt.DefaultReconcileTimeout != 0 {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(nt.DefaultReconcileTimeout)
	}
	// Enable automatic deletion of managed objects by default.
	// This helps ensure that test artifacts are cleaned up.
	EnableDeletionPropagation(rs)
	// Add dependencies to ensure managed objects can be deleted.
	if err := SetRepoSyncDependencies(nt, rs); err != nil {
		nt.T.Fatal(err)
	}
	return rs
}

// RepoSyncObjectV1Beta1 returns the default RepoSync object
// with version v1beta1 in the given namespace.
func RepoSyncObjectV1Beta1(nn types.NamespacedName, repoURL string, sourceFormat filesystem.SourceFormat) *v1beta1.RepoSync {
	rs := fake.RepoSyncObjectV1Beta1(nn.Namespace, nn.Name)
	rs.Spec.SourceFormat = string(sourceFormat)
	rs.Spec.SourceType = string(v1beta1.GitSource)
	rs.Spec.Git = &v1beta1.Git{
		Repo:   repoURL,
		Branch: gitproviders.MainBranch,
		Dir:    gitproviders.DefaultSyncDir,
		Auth:   "ssh",
		SecretRef: &v1beta1.SecretReference{
			Name: "ssh-key",
		},
	}
	// Enable automatic deletion of managed objects by default.
	// This helps ensure that test artifacts are cleaned up.
	EnableDeletionPropagation(rs)
	return rs
}

// RepoSyncObjectV1Beta1FromNonRootRepo returns a v1beta1 RepoSync object which
// uses a repo from nt.NonRootRepos.
func RepoSyncObjectV1Beta1FromNonRootRepo(nt *NT, nn types.NamespacedName) *v1beta1.RepoSync {
	repo, found := nt.NonRootRepos[nn]
	if !found {
		nt.T.Fatal("nonexistent non-root repo: %s", nn)
	}
	repoURL := nt.GitProvider.SyncURL(repo.RemoteRepoName)
	sourceFormat := repo.Format
	rs := RepoSyncObjectV1Beta1(nn, repoURL, sourceFormat)
	if nt.DefaultReconcileTimeout != 0 {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(nt.DefaultReconcileTimeout)
	}
	// Add dependencies to ensure managed objects can be deleted.
	if err := SetRepoSyncDependencies(nt, rs); err != nil {
		nt.T.Fatal(err)
	}
	return rs
}

// RepoSyncObjectV1Beta1FromOtherRootRepo returns a v1beta1 RepoSync object which
// uses a repo from nt.RootRepos.
func RepoSyncObjectV1Beta1FromOtherRootRepo(nt *NT, nn types.NamespacedName, repoName string) *v1beta1.RepoSync {
	repo, found := nt.RootRepos[repoName]
	if !found {
		nt.T.Fatal("nonexistent root repo: %s", repoName)
	}
	repoURL := nt.GitProvider.SyncURL(repo.RemoteRepoName)
	sourceFormat := repo.Format
	rs := RepoSyncObjectV1Beta1(nn, repoURL, sourceFormat)
	if nt.DefaultReconcileTimeout != 0 {
		rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(nt.DefaultReconcileTimeout)
	}
	// Add dependencies to ensure managed objects can be deleted.
	if err := SetRepoSyncDependencies(nt, rs); err != nil {
		nt.T.Fatal(err)
	}
	return rs
}

// NotificationObjectV1Beta1 returns a v1beta1 Notification object
func NotificationObjectV1Beta1(nn types.NamespacedName) *v1beta1.Notification {
	notification := &v1beta1.Notification{}
	notification.Namespace = nn.Namespace
	notification.Name = nn.Name
	return notification
}

// setupCentralizedControl is a pure central-control mode.
// A default root repo (root-sync) manages all other root repos and namespace repos.
func setupCentralizedControl(nt *NT, reconcileTimeout *time.Duration) {
	nt.T.Log("[SETUP] Centralized control")

	rsCount := 0

	rootSyncNN := RootSyncNN(configsync.RootSyncName)

	// Add any RootSyncs specified by the test options
	for rsName := range nt.RootRepos {
		// The default RootSync is created manually, don't check it in the repo.
		if rsName == configsync.RootSyncName {
			continue
		}
		rs := RootSyncObjectV1Beta1FromRootRepo(nt, rsName)
		if reconcileTimeout != nil {
			rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(*reconcileTimeout)
		}
		nt.Must(nt.RootRepos[configsync.RootSyncName].Add(fmt.Sprintf("acme/namespaces/%s/%s.yaml", configsync.ControllerNamespace, rsName), rs))
		nt.MetricsExpectations.AddObjectApply(configsync.RootSyncKind, rootSyncNN, rs)
		nt.Must(nt.RootRepos[configsync.RootSyncName].CommitAndPush("Adding RootSync: " + rsName))
	}

	if len(nt.NonRootRepos) == 0 {
		return
	}

	// Add ClusterRole for all RepoSync reconcilers to use.
	// TODO: Test different permissions for different RepoSyncs.
	cr := nt.RepoSyncClusterRole()
	nt.Must(nt.RootRepos[configsync.RootSyncName].Add("acme/cluster/cr.yaml", cr))
	nt.MetricsExpectations.AddObjectApply(configsync.RootSyncKind, rootSyncNN, cr)

	// Use a map to record the number of RepoSync namespaces
	rsNamespaces := map[string]struct{}{}

	// Add any RepoSyncs specified by the test options
	for rsNN := range nt.NonRootRepos {
		ns := rsNN.Namespace
		rsCount++
		rsNamespaces[ns] = struct{}{}

		// Add Namespace for this RepoSync.
		// This may replace an existing Namespace, if there are multiple
		// RepoSyncs in the same namespace.
		nsObj := fake.NamespaceObject(ns)
		nt.Must(nt.RootRepos[configsync.RootSyncName].Add(StructuredNSPath(ns, ns), nsObj))
		nt.MetricsExpectations.AddObjectApply(configsync.RootSyncKind, rootSyncNN, nsObj)

		// Add RoleBinding to bind the ClusterRole to the RepoSync reconciler ServiceAccount
		rb := RepoSyncRoleBinding(rsNN)
		nt.Must(nt.RootRepos[configsync.RootSyncName].Add(StructuredNSPath(ns, fmt.Sprintf("rb-%s", rsNN.Name)), rb))
		nt.MetricsExpectations.AddObjectApply(configsync.RootSyncKind, rootSyncNN, rb)

		if isPSPCluster() {
			// Add a ClusterRoleBinding so that the pods can be created
			// when the cluster has PodSecurityPolicy enabled.
			// Background: If a RoleBinding (not a ClusterRoleBinding) is used,
			// it will only grant usage for pods being run in the same namespace as the binding.
			// TODO: Remove the psp related change when Kubernetes 1.25 is
			// available on GKE.
			crb := repoSyncClusterRoleBinding(rsNN)
			nt.Must(nt.RootRepos[configsync.RootSyncName].Add(fmt.Sprintf("acme/cluster/crb-%s-%s.yaml", ns, rsNN.Name), crb))
			nt.MetricsExpectations.AddObjectApply(configsync.RootSyncKind, rootSyncNN, crb)
		}

		// Add RepoSync pointing to the Git repo specified in nt.NonRootRepos[rsNN]
		rs := RepoSyncObjectV1Beta1FromNonRootRepo(nt, rsNN)
		if reconcileTimeout != nil {
			rs.Spec.SafeOverride().ReconcileTimeout = toMetav1Duration(*reconcileTimeout)
		}
		nt.Must(nt.RootRepos[configsync.RootSyncName].Add(StructuredNSPath(ns, rsNN.Name), rs))
		nt.MetricsExpectations.AddObjectApply(configsync.RootSyncKind, rootSyncNN, rs)

		nt.Must(nt.RootRepos[configsync.RootSyncName].CommitAndPush("Adding namespace, clusterrole, rolebinding, clusterrolebinding and RepoSync"))
	}

	// Convert namespace set to list
	namespaces := make([]string, 0, len(rsNamespaces))
	for namespace := range rsNamespaces {
		namespaces = append(namespaces, namespace)
	}

	// Wait for all Namespaces to be applied and reconciled.
	WaitForNamespaces(nt, nt.DefaultWaitTimeout, namespaces...)

	// Now that the Namespaces exist, create the Secrets,
	// which are required for the RepoSyncs to reconcile.
	for nn := range nt.NonRootRepos {
		CreateNamespaceSecret(nt, nn.Namespace)
	}
}

func validateRootSyncsExist(nt *NT) error {
	var err error
	for rsName := range nt.RootRepos {
		err = multierr.Append(err,
			nt.Validate(rsName, configsync.ControllerNamespace, &v1beta1.RootSync{}))
	}
	return err
}

func validateRepoSyncsExist(nt *NT) error {
	var err error
	for nn := range nt.NonRootRepos {
		err = multierr.Append(err,
			nt.Validate(nn.Name, nn.Namespace, &v1beta1.RepoSync{}))
	}
	return err
}

// SetRepoSyncDependencies sets the depends-on annotation on the RepoSync for
// other objects managed by CentralControl (`setupCentralizedControl`).
//
// RootSync dependencies aren't managed by another RootSync, and
// DelegatedControl doesn't apply with ConfigSync, so depends-on won't have
// any effect in those scenerios.
//
// Takes an Object because it supports both v1alpha1 and v1beta1 RepoSyncs.
func SetRepoSyncDependencies(nt *NT, rs client.Object) error {
	if nt.Control != ntopts.CentralControl {
		return nil
	}
	gvk, err := kinds.Lookup(rs, nt.Scheme)
	if err != nil {
		return err
	}
	if gvk.Kind != "RepoSync" {
		return fmt.Errorf("expected RepoSync, but got %s", gvk.Kind)
	}

	rsNN := client.ObjectKeyFromObject(rs)
	dependencies := []client.Object{
		nt.RepoSyncClusterRole(),
		RepoSyncRoleBinding(rsNN),
	}
	if isPSPCluster() {
		crb := repoSyncClusterRoleBinding(rsNN)
		dependencies = append(dependencies, crb)
	}
	return SetDependencies(rs, dependencies...)
}

// SetDependencies sets the specified objects as dependencies of the first object.
func SetDependencies(obj client.Object, dependencies ...client.Object) error {
	var deps dependson.DependencySet
	for _, dep := range dependencies {
		deps = append(deps, applier.ObjMetaFromObject(dep))
	}
	if err := setDependsOnAnnotation(obj, deps); err != nil {
		return errors.Wrapf(err, "failed to set dependencies on %T %s", obj, client.ObjectKeyFromObject(obj))
	}
	return nil
}

// setDependsOnAnnotation is a copy of cli-utils dependson.WriteAnnotation, but
// it accepts a core.Annotated instead of an Unstructured.
func setDependsOnAnnotation(obj core.Annotated, depSet dependson.DependencySet) error {
	if obj == nil {
		return errors.New("object is nil")
	}
	if depSet.Equal(dependson.DependencySet{}) {
		return errors.New("dependency set is empty")
	}
	depSetStr, err := dependson.FormatDependencySet(depSet)
	if err != nil {
		return fmt.Errorf("failed to format depends-on annotation: %w", err)
	}
	core.SetAnnotation(obj, dependson.Annotation, depSetStr)
	return nil
}

// podHasReadyCondition checks if a pod status has a READY condition.
func podHasReadyCondition(conditions []corev1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// NewPodReady checks if the new created pods are ready.
// It also checks if the new children pods that are managed by the pods are ready.
func NewPodReady(nt *NT, labelName, currentLabel, childLabel string, oldCurrentPods, oldChildPods []corev1.Pod) error {
	if len(oldCurrentPods) == 0 {
		return nil
	}
	newPods := &corev1.PodList{}
	if err := nt.KubeClient.List(newPods, client.InNamespace(configmanagement.ControllerNamespace), client.MatchingLabels{labelName: currentLabel}); err != nil {
		nt.T.Fatal(err)
	}
	for _, newPod := range newPods.Items {
		for _, oldPod := range oldCurrentPods {
			if newPod.Name == oldPod.Name {
				return fmt.Errorf("old pod %s is still alive", oldPod.Name)
			}
		}
		if !podHasReadyCondition(newPod.Status.Conditions) {
			return fmt.Errorf("new pod %s is not ready yet", currentLabel)
		}
	}
	return NewPodReady(nt, labelName, childLabel, "", oldChildPods, nil)
}

// DeletePodByLabel deletes pods that have the label and waits until new pods come up.
func DeletePodByLabel(nt *NT, label, value string, waitForChildren bool) {
	oldPods := &corev1.PodList{}
	if err := nt.KubeClient.List(oldPods, client.InNamespace(configmanagement.ControllerNamespace), client.MatchingLabels{label: value}); err != nil {
		nt.T.Fatal(err)
	}
	oldReconcilers := &corev1.PodList{}
	if value == reconcilermanager.ManagerName && waitForChildren {
		// Get the children pods managed by the reconciler-manager.
		if err := nt.KubeClient.List(oldReconcilers, client.InNamespace(configmanagement.ControllerNamespace), client.MatchingLabels{label: reconcilermanager.Reconciler}); err != nil {
			nt.T.Fatal(err)
		}
	}
	if err := nt.KubeClient.DeleteAllOf(&corev1.Pod{}, client.InNamespace(configmanagement.ControllerNamespace), client.MatchingLabels{label: value}); err != nil {
		nt.T.Fatalf("Pod delete failed: %v", err)
	}
	Wait(nt.T, "new pods come up", nt.DefaultWaitTimeout, func() error {
		if value == reconcilermanager.ManagerName && waitForChildren {
			// Wait for both the new reconciler-manager pod and the new reconcilers are ready.
			return NewPodReady(nt, label, value, reconcilermanager.Reconciler, oldPods.Items, oldReconcilers.Items)
		}
		return NewPodReady(nt, label, value, "", oldPods.Items, nil)
	}, WaitTimeout(nt.DefaultWaitTimeout))
}

// SetPolicyDir updates the root-sync object with the provided policyDir.
func SetPolicyDir(nt *NT, syncName, policyDir string) {
	nt.T.Logf("Set policyDir to %q", policyDir)
	rs := fake.RootSyncObjectV1Beta1(syncName)
	if err := nt.KubeClient.Get(syncName, configmanagement.ControllerNamespace, rs); err != nil {
		nt.T.Fatal(err)
	}
	if rs.Spec.Git.Dir != policyDir {
		nt.MustMergePatch(rs, fmt.Sprintf(`{"spec": {"git": {"dir": "%s"}}}`, policyDir))
	}
}

// SetGitBranch updates the root-sync object with the provided git branch
func SetGitBranch(nt *NT, syncName, branch string) {
	nt.T.Logf("Change git branch to %q", branch)
	rs := fake.RootSyncObjectV1Beta1(syncName)
	if err := nt.KubeClient.Get(syncName, configmanagement.ControllerNamespace, rs); err != nil {
		nt.T.Fatal(err)
	}
	if rs.Spec.Git.Branch != branch {
		nt.MustMergePatch(rs, fmt.Sprintf(`{"spec": {"git": {"branch": "%s"}}}`, branch))
	}
}

func toMetav1Duration(t time.Duration) *metav1.Duration {
	return &metav1.Duration{
		Duration: t,
	}
}
