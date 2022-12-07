package util

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj/notifications-engine/pkg/subscriptions"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"kpt.dev/configsync/pkg/api/configsync/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NotificationApiGroup          = "NOTIFICATION_API_GROUP"
	NotificationApiVersion        = "NOTIFICATION_API_VERSION"
	NotificationApiKind           = "NOTIFICATION_API_KIND"
	NotificationResourceName      = "NOTIFICATION_RESOURCE_NAME"
	NotificationResourceNamespace = "NOTIFICATION_RESOURCE_NAMESPACE"
	NotificationResyncPeriod      = "NOTIFICATION_RESYNC_PERIOD"
	NotificationConfigMapName     = "NOTIFICATION_CONFIGMAP_NAME"
	NotificationSecretName        = "NOTIFICATION_SECRET_NAME"

	// SubscribeAnnotationPrefix is the annotation prefix of the notification subscription.
	SubscribeAnnotationPrefix = subscriptions.AnnotationPrefix + "/subscribe."

	// MultiSubscriptionsField is the field name for subscriptions configured globally.
	MultiSubscriptionsField = "subscriptions"

	// NotifiedAnnotationPath is the annotation path that indicates it has been notified.
	NotifiedAnnotationPath = ".metadata.annotations.notified.notifications.argoproj.io"
)

// NotificationEnabled returns whether the notification is enabled for the RSync object.
func NotificationEnabled(ctx context.Context, client client.Client, rsNamespace string, annotations map[string]string, notificationConfig *v1beta1.NotificationConfig) (bool, error) {
	for key := range annotations {
		if strings.HasPrefix(key, SubscribeAnnotationPrefix) {
			return true, nil
		}
	}

	if notificationConfig != nil && notificationConfig.ConfigMapRef != nil && notificationConfig.ConfigMapRef.Name != "" {
		cm := &corev1.ConfigMap{}
		cmObjectKey := types.NamespacedName{
			Namespace: rsNamespace,
			Name:      notificationConfig.ConfigMapRef.Name,
		}
		if err := client.Get(ctx, cmObjectKey, cm); err != nil {
			if apierrors.IsNotFound(err) {
				return false, fmt.Errorf("notification ConfigMap %s not found in the %s namespace", cmObjectKey.Name, cmObjectKey.Namespace)
			}
			return false, err
		}
		_, found := cm.Data[MultiSubscriptionsField]
		return found, nil
	}
	return false, nil
}
