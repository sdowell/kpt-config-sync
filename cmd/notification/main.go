// Copyright 2023 Google LLC
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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/argoproj/notifications-engine/pkg/api"
	"github.com/argoproj/notifications-engine/pkg/controller"
	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/argoproj/notifications-engine/pkg/subscriptions"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"kpt.dev/configsync/pkg/notifications"
	stringexpressions "kpt.dev/configsync/pkg/notifications/expressions/strings"
	utilexpressions "kpt.dev/configsync/pkg/notifications/expressions/utils"
	"kpt.dev/configsync/pkg/reconcilermanager/controllers"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	csclientv1beta1 "kpt.dev/configsync/clientgen/apis/typed/configsync/v1beta1"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	"kpt.dev/configsync/pkg/core"
	"kpt.dev/configsync/pkg/kinds"
	"kpt.dev/configsync/pkg/util"
	"kpt.dev/configsync/pkg/util/log"
	ctrl "sigs.k8s.io/controller-runtime"
)

const objectKey = "sync"
const fieldManager = "notification-controller"

var (
	apiGroup = flag.String("api-group", util.EnvString(util.NotificationAPIGroup, v1beta1.SchemeGroupVersion.Group),
		"Group of the Resource for notifications")
	apiVersion = flag.String("api-version", util.EnvString(util.NotificationAPIVersion, v1beta1.SchemeGroupVersion.Version),
		"Version of the Resource for notifications")
	apiKind = flag.String("api-kind", os.Getenv(util.NotificationAPIKind),
		"Resource information of the Resource for notifications")
	resourceName = flag.String("resource-name", os.Getenv(util.NotificationResourceName),
		"Name of the Resource to be notified")
	resourceNamespace = flag.String("resource-namespace", os.Getenv(util.NotificationResourceNamespace),
		"Namespace of the Resource to be notified")
	resyncCheckPeriod = flag.Duration("resync-period", controllers.PollingPeriod(util.NotificationResyncPeriod, time.Minute),
		"Period of time between checking if the apiResource listener needs a resync")
	cmName = flag.String("cm-name", os.Getenv(util.NotificationConfigMapName),
		"Name of the ConfigMap for the notification configs")
	secretName = flag.String("secret-name", os.Getenv(util.NotificationSecretName),
		"Name of the Secret for the notification configs")
)

// toUnstructured converts the object to unstructured with the provided API kind
func toUnstructured(obj metav1.Object) (*unstructured.Unstructured, error) {
	switch obj := obj.(type) {
	case *unstructured.Unstructured:
		switch *apiKind {
		case kinds.RootSyncV1Beta1().Kind:
			rs := &v1beta1.RootSync{}
			unstructuredContent := obj.UnstructuredContent()
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredContent, rs); err != nil {
				return nil, fmt.Errorf("failed to convert the unstructured object into RootSync: %v", err)
			}
			return kinds.ToUnstructured(rs, core.Scheme)
		case kinds.RepoSyncV1Beta1().Kind:
			rs := &v1beta1.RepoSync{}
			unstructuredContent := obj.UnstructuredContent()
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredContent, rs); err != nil {
				return nil, fmt.Errorf("failed to convert the unstructured object into RepoSync: %v", err)
			}
			return kinds.ToUnstructured(rs, core.Scheme)
		}
	case *v1beta1.RootSync:
		return kinds.ToUnstructured(obj, core.Scheme)
	case *v1beta1.RepoSync:
		return kinds.ToUnstructured(obj, core.Scheme)
	}
	return nil, fmt.Errorf("unknown object type %T", obj)
}

func initGetVars(obj map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		objectKey: obj,
		// strings contains string-related helper functions
		"strings": stringexpressions.New(),
		// utils contains helper functions for RootSync/RepoSync
		"utils": utilexpressions.New(&unstructured.Unstructured{Object: obj}, *apiKind),
	}
}

func main() {
	log.Setup()
	ctrl.SetLogger(klogr.New())

	subscriptions.SetAnnotationPrefix(util.AnnotationsPrefix)

	// Get Kubernetes REST Config and current Namespace so we can talk to Kubernetes
	restConfig := ctrl.GetConfigOrDie()

	// Create ConfigMap and Secret informer to access notifications configuration
	informersFactory := informers.NewSharedInformerFactoryWithOptions(
		kubernetes.NewForConfigOrDie(restConfig),
		*resyncCheckPeriod,
		informers.WithNamespace(*resourceNamespace))
	secrets := informersFactory.Core().V1().Secrets().Informer()
	configMaps := informersFactory.Core().V1().ConfigMaps().Informer()

	// Create "Notifications" API factory that handles notifications processing
	notificationsFactory := api.NewFactory(api.Settings{
		ConfigMapName: *cmName,
		SecretName:    *secretName,
		InitGetVars: func(cfg *api.Config, configMap *v1.ConfigMap, secret *v1.Secret) (api.GetVars, error) {
			return func(obj map[string]interface{}, dest services.Destination) map[string]interface{} {
				return initGetVars(obj)
			}, nil
		},
	}, *resourceNamespace, secrets, configMaps)

	// Create notifications controller that handles Kubernetes resources processing
	apiResource := fmt.Sprintf("%ss", strings.ToLower(*apiKind))
	gvr := schema.GroupVersionResource{
		Group: *apiGroup, Version: *apiVersion, Resource: apiResource,
	}
	notificationClient := dynamic.NewForConfigOrDie(restConfig).Resource(gvr)
	notificationStatusClient := dynamic.NewForConfigOrDie(restConfig).Resource(v1beta1.GroupVersionResource("notifications")).Namespace(*resourceNamespace)

	fieldsSelector := fields.SelectorFromSet(map[string]string{"metadata.name": *resourceName})
	csClient, err := csclientv1beta1.NewForConfig(restConfig)
	if err != nil {
		klog.Fatal(err)
	}

	var exampleObject runtime.Object
	if *apiKind == kinds.RootSyncV1Beta1().Kind {
		exampleObject = &v1beta1.RootSync{}
	} else {
		exampleObject = &v1beta1.RepoSync{}
	}

	notificationInformer := cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(csClient.RESTClient(), apiResource, *resourceNamespace, fieldsSelector),
		exampleObject, *resyncCheckPeriod, cache.Indexers{})

	statusUpdater := notificationStatusUpdater(notificationStatusClient)

	notificationController := controller.NewController(
		notificationClient,
		notificationInformer,
		notificationsFactory,
		controller.WithToUnstructured(toUnstructured),
		controller.WithEventCallback(func(eventSequence controller.NotificationEventSequence) {
			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("Recovered in eventCallback: %v", r)
				}
			}()
			err := statusUpdater(eventSequence)
			if err != nil {
				klog.Errorf("Failed to update notification status for %s/%s. Err: %v",
					eventSequence.Resource.GetNamespace(), eventSequence.Resource.GetName(), err)
			}
		}),
	)

	// Start informers and controller
	go informersFactory.Start(context.Background().Done())
	go notificationInformer.Run(context.Background().Done())
	if !cache.WaitForCacheSync(context.Background().Done(), secrets.HasSynced, configMaps.HasSynced, notificationInformer.HasSynced) {
		klog.Fatalf("Failed to synchronize informers")
	}

	notificationController.Run(1, context.Background().Done())
}

func notificationStatusUpdater(notificationStatusClient dynamic.ResourceInterface) func(eventSequence controller.NotificationEventSequence) error {
	var mux sync.Mutex
	return func(eventSequence controller.NotificationEventSequence) error {
		mux.Lock()
		defer mux.Unlock()
		ctx := context.Background()
		unstructuredRSync, err := toUnstructured(eventSequence.Resource)
		if err != nil {
			return err
		}
		curNotificationUn, err := notificationStatusClient.Get(ctx, unstructuredRSync.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}
		curNotification := &v1beta1.Notification{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(curNotificationUn.UnstructuredContent(), curNotification); err != nil {
			return err
		}
		commit, err := util.LastCommit(unstructuredRSync, *apiKind)
		if err != nil {
			return err
		}
		newNotificationStatus := notifications.CreateNotificationStatus(unstructuredRSync.GetGeneration(), commit, eventSequence)
		if notifications.IsNotificationStatusSame(curNotification.Status, newNotificationStatus) {
			return nil
		}
		curNotification.Status = newNotificationStatus
		un, err := kinds.ToUnstructured(curNotification, core.Scheme)
		if err != nil {
			return err
		}
		up, err := notificationStatusClient.UpdateStatus(
			ctx,
			un,
			metav1.UpdateOptions{
				FieldManager: fieldManager,
			},
		)
		if err != nil {
			return err
		}
		klog.Infof("updated notification: %s/%s - %s", up.GetNamespace(), up.GetName(), up.GetResourceVersion())
		return nil
	}
}