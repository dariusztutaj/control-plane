package deprovisioning

import (
	"testing"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/reconciler"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeregisterClusterStep_Run(t *testing.T) {
	// given
	cli := reconciler.NewFakeClient()
	memoryStorage := storage.NewMemoryStorage()
	step := NewDeregisterClusterStep(memoryStorage.Operations(), cli)
	op := fixDeprovisioningOperation()
	op.ClusterConfigurationVersion = 1
	memoryStorage.Operations().InsertDeprovisioningOperation(op)
	op.RuntimeID = "runtime-id"
	cli.ApplyClusterConfig(reconciler.Cluster{
		Cluster: op.RuntimeID,
	})

	// when
	_, d, err := step.Run(op, logrus.New())

	// then
	require.NoError(t, err)
	assert.Zero(t, d)
	assert.False(t, cli.IsClusterExists(op.RuntimeID))
}

func TestDeregisterClusterStep_RunForNotExistingCluster(t *testing.T) {
	// given
	cli := reconciler.NewFakeClient()
	memoryStorage := storage.NewMemoryStorage()
	step := NewDeregisterClusterStep(memoryStorage.Operations(), cli)
	op := fixDeprovisioningOperation()
	op.ClusterConfigurationVersion = 1
	op.ClusterConfigurationDeleted = true
	memoryStorage.Operations().InsertDeprovisioningOperation(op)
	op.RuntimeID = "runtime-id"

	// when
	_, d, err := step.Run(op, logrus.New())

	// then
	require.NoError(t, err)
	assert.Zero(t, d)
	assert.False(t, cli.IsClusterExists(op.RuntimeID))
}
