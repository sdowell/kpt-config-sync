package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/argoproj/notifications-engine/pkg/api"
	"github.com/argoproj/notifications-engine/pkg/controller"
	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	csclientv1beta1 "kpt.dev/configsync/clientgen/apis/typed/configsync/v1beta1"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	"kpt.dev/configsync/pkg/core"
	"kpt.dev/configsync/pkg/kinds"
	"kpt.dev/configsync/pkg/util"
	"kpt.dev/configsync/pkg/util/log"
	ctrl "sigs.k8s.io/controller-runtime"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
)

const ObjectKey = "sync"

var (
	apiGroup = flag.String("api-group", util.EnvString(util.NotificationApiGroup, v1beta1.SchemeGroupVersion.Group),
		"Group of the Resource for notifications")
	apiVersion = flag.String("api-version", util.EnvString(util.NotificationApiVersion, v1beta1.SchemeGroupVersion.Version),
		"Version of the Resource for notifications")
	apiKind = flag.String("api-kind", os.Getenv(util.NotificationApiKind),
		"Resource information of the Resource for notifications")
	resourceName = flag.String("resource-name", os.Getenv(util.NotificationResourceName),
		"Name of the Resource to be notified")
	resourceNamespace = flag.String("resource-namespace", os.Getenv(util.NotificationResourceNamespace),
		"Namespace of the Resource to be notified")
	resyncCheckPeriod = flag.Duration("resync-period", pollingPeriod(util.NotificationResyncPeriod, time.Minute),
		"Period of time between checking if the apiResource listener needs a resync")
	cmName = flag.String("cm-name", os.Getenv(util.NotificationConfigMapName),
		"Name of the ConfigMap for the notification configs")
	secretName = flag.String("secret-name", os.Getenv(util.NotificationSecretName),
		"Name of the Secret for the notification configs")
)

// PollingPeriod parses the polling duration from the environment variable.
// If the variable is not present, it returns the default value.
func pollingPeriod(envName string, defaultValue time.Duration) time.Duration {
	val, present := os.LookupEnv(envName)
	if present {
		pollingFreq, err := time.ParseDuration(val)
		if err != nil {
			panic(errors.Wrapf(err, "failed to parse environment variable %q,"+
				"got value: %v, want err: nil", envName, pollingFreq))
		}
		return pollingFreq
	}
	return defaultValue
}

type statusHook struct {
	resourceName string
}

func (sh *statusHook) Fire(entry *logrus.Entry) error {
	resource, ok := entry.Data["resource"]
	if !ok {
		return nil // this is a log entry for something else
	}
	if resource != sh.resourceName {
		return nil
	}
	message, err := entry.String()
	if err != nil {
		klog.Infof("Failed to parse log entry:\n%v", err)
		return nil
	}
	// This is a log entry for our resource!
	klog.Errorf("+++ Found an error in the notification controller:\n%s", message)
	return nil
}

func (sh *statusHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
	}
}

func main() {
	log.Setup()
	ctrl.SetLogger(klogr.New())
	logrus.AddHook(&statusHook{
		resourceName: fmt.Sprintf("%s/%s", *resourceNamespace, *resourceName),
	})

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
				return map[string]interface{}{ObjectKey: obj}
			}, nil
		},
	}, *resourceNamespace, secrets, configMaps)

	// Create notifications controller that handles Kubernetes resources processing
	apiResource := fmt.Sprintf("%ss", strings.ToLower(*apiKind))
	gvr := schema.GroupVersionResource{
		Group: *apiGroup, Version: *apiVersion, Resource: apiResource,
	}
	notificationClient := dynamic.NewForConfigOrDie(restConfig).Resource(gvr)

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

	notificationController := controller.NewController(notificationClient, notificationInformer, notificationsFactory, controller.WithToUnstructured(func(obj metav1.Object) (*unstructured.Unstructured, error) {
		switch obj.(type) {
		case *unstructured.Unstructured:
			switch *apiKind {
			case kinds.RootSyncV1Beta1().Kind:
				rs := &v1beta1.RootSync{}
				unstructuredContent := obj.(*unstructured.Unstructured).UnstructuredContent()
				if err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredContent, rs); err != nil {
					return nil, fmt.Errorf("failed to convert the unstructured object into RootSync: %v", err)
				}
				return kinds.ToUnstructured(rs, core.Scheme)
			case kinds.RepoSyncV1Beta1().Kind:
				rs := &v1beta1.RepoSync{}
				unstructuredContent := obj.(*unstructured.Unstructured).UnstructuredContent()
				if err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredContent, rs); err != nil {
					return nil, fmt.Errorf("failed to convert the unstructured object into RepoSync: %v", err)
				}
				return kinds.ToUnstructured(rs, core.Scheme)
			}
		case *v1beta1.RootSync:
			rs, ok := obj.(*v1beta1.RootSync)
			if !ok {
				return nil, fmt.Errorf("object must be v1beta1.RootSync but is %T", rs)
			}
			return kinds.ToUnstructured(rs, core.Scheme)
		case *v1beta1.RepoSync:
			rs, ok := obj.(*v1beta1.RepoSync)
			if !ok {
				return nil, fmt.Errorf("object must be v1beta1.RepoSync but is %T", rs)
			}
			return kinds.ToUnstructured(rs, core.Scheme)
		}
		return nil, fmt.Errorf("unknown object type %T", obj)
	}))

	// Start informers and controller
	go informersFactory.Start(context.Background().Done())
	go notificationInformer.Run(context.Background().Done())
	if !cache.WaitForCacheSync(context.Background().Done(), secrets.HasSynced, configMaps.HasSynced, notificationInformer.HasSynced) {
		klog.Fatalf("Failed to synchronize informers")
	}

	notificationController.Run(10, context.Background().Done())
}
