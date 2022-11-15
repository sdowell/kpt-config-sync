package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/argoproj/notifications-engine/pkg/api"
	"github.com/argoproj/notifications-engine/pkg/controller"
	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"kpt.dev/configsync/pkg/api/configsync"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	"kpt.dev/configsync/pkg/util"
	"kpt.dev/configsync/pkg/util/log"
	ctrl "sigs.k8s.io/controller-runtime"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
)

const (
	ApiGroup      = "API_GROUP"
	ApiVersion    = "API_VERSION"
	ApiResource   = "API_RESOURCE"
	ResyncPeriod  = "RESYNC_PERIOD"
	ConfigMapName = "CONFIGMAP_NAME"
	SecretName    = "SECRET_NAME"
	ObjectKey     = "sync"
)

var (
	apiGroup = flag.String("api-group", util.EnvString(ApiGroup, v1beta1.SchemeGroupVersion.Group),
		"Group of the Resource for notifications")
	apiVersion = flag.String("api-version", util.EnvString(ApiVersion, v1beta1.SchemeGroupVersion.Version),
		"Version of the Resource for notifications")
	apiResource = flag.String("api-resource", os.Getenv(ApiResource),
		"Resource information of the Resource for notifications")
	resyncCheckPeriod = flag.Duration("resync-period", pollingPeriod(ResyncPeriod, time.Minute),
		"Period of time between checking if the apiResource listener needs a resync")
	cmName = flag.String("cm-name", os.Getenv(ConfigMapName),
		"Name of the ConfigMap for the notification configs")
	secretName = flag.String("secret-name", os.Getenv(SecretName),
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

func main() {
	log.Setup()
	ctrl.SetLogger(klogr.New())

	// Get Kubernetes REST Config and current Namespace so we can talk to Kubernetes
	restConfig := ctrl.GetConfigOrDie()

	// Create ConfigMap and Secret informer to access notifications configuration
	informersFactory := informers.NewSharedInformerFactoryWithOptions(
		kubernetes.NewForConfigOrDie(restConfig),
		*resyncCheckPeriod,
		informers.WithNamespace(configsync.ControllerNamespace))
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
	}, configsync.ControllerNamespace, secrets, configMaps)

	// Create notifications controller that handles Kubernetes resources processing
	csClient := dynamic.NewForConfigOrDie(restConfig).Resource(schema.GroupVersionResource{
		Group: *apiGroup, Version: *apiVersion, Resource: *apiResource,
	})
	csInformer := cache.NewSharedIndexInformer(&cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return csClient.List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return csClient.Watch(context.Background(), metav1.ListOptions{})
		},
	}, &unstructured.Unstructured{}, *resyncCheckPeriod, cache.Indexers{})
	notificationController := controller.NewController(csClient, csInformer, notificationsFactory)

	// Start informers and controller
	go informersFactory.Start(context.Background().Done())
	go csInformer.Run(context.Background().Done())
	if !cache.WaitForCacheSync(context.Background().Done(), secrets.HasSynced, configMaps.HasSynced, csInformer.HasSynced) {
		klog.Fatalf("Failed to synchronize informers")
	}

	notificationController.Run(10, context.Background().Done())
}
