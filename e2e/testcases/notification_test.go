package e2e

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	"kpt.dev/configsync/e2e/nomostest"
	"kpt.dev/configsync/e2e/nomostest/ntopts"
	nomostesting "kpt.dev/configsync/e2e/nomostest/testing"
)

func TestNotification(t *testing.T) {
	nt := nomostest.New(t, nomostesting.SyncSource, ntopts.InstallNotificationServer)
	nt.WaitForRepoSyncs()

	_, err := nomostest.NotificationSecret(nt, "config-management-system")
	if err != nil {
		nt.T.Fatal(err)
	}
	_, err = nomostest.NotificationConfigMap(nt, "config-management-system")
	if err != nil {
		nt.T.Fatal(err)
	}
	rootSyncRef := types.NamespacedName{
		Name:      "root-sync",
		Namespace: "config-management-system",
	}
	err = nomostest.SubscribeRootSyncNotification(nt, rootSyncRef)
	if err != nil {
		nt.T.Fatal(err)
	}

	var records *nomostest.NotificationRecords
	took, err := nomostest.Retry(30*time.Second, func() error {
		records, err = nt.NotificationServer.DoGet()
		if err != nil {
			return err
		}
		if len(records.Records) == 0 {
			return fmt.Errorf("want x, got %d", len(records.Records))
		}
		return nil
	})
	nt.T.Logf("took %v to wait for notification", took)
	if err != nil {
		nt.T.Fatal(err)
	}
	assert.Equal(t, nomostest.NotificationRecords{
		Records: []nomostest.NotificationRecord{
			{
				Message: "{\n  \"content\": {\n    \"raw\": \"RootSync root-sync is synced!\"\n  }\n}",
				Auth:    fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("user:pass"))), // base64 encoded username/pass
			},
		},
	}, *records)

	// check that no duplicate notifications were delivered
	time.Sleep(30 * time.Second)
	records, err = nt.NotificationServer.DoGet()
	if err != nil {
		nt.T.Fatal(err)
	}
	assert.Equal(t, 1, len(records.Records))
}
