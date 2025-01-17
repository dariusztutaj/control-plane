package provisioning

import (
	"fmt"
	"time"

	kebError "github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/error"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/reconciler"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"
)

// CheckClusterConfigurationStep checks if the SKR configuration is applied (by reconciler)
type CheckClusterConfigurationStep struct {
	reconcilerClient    reconciler.Client
	operationManager    *process.ProvisionOperationManager
	provisioningTimeout time.Duration
}

func NewCheckClusterConfigurationStep(os storage.Operations,
	reconcilerClient reconciler.Client,
	provisioningTimeout time.Duration) *CheckClusterConfigurationStep {
	return &CheckClusterConfigurationStep{
		reconcilerClient:    reconcilerClient,
		operationManager:    process.NewProvisionOperationManager(os),
		provisioningTimeout: provisioningTimeout,
	}
}

var _ Step = (*CheckClusterConfigurationStep)(nil)

func (s *CheckClusterConfigurationStep) Name() string {
	return "Check_Cluster_Configuration"
}

func (s *CheckClusterConfigurationStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > s.provisioningTimeout {
		log.Infof("operation has reached the time limit: updated operation time: %s", operation.UpdatedAt)
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", s.provisioningTimeout), log)
	}

	state, err := s.reconcilerClient.GetCluster(operation.RuntimeID, operation.ClusterConfigurationVersion)
	if kebError.IsTemporaryError(err) {
		log.Errorf("Reconciler GetCluster method failed (temporary error, retrying): %s", err.Error())
		return operation, 1 * time.Minute, nil
	}
	if err != nil {
		log.Errorf("Reconciler GetCluster method failed: %s", err.Error())
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("unable to get cluster state: %s", err.Error()), log)
	}
	log.Debugf("Cluster configuration status %s", state.Status)

	switch state.Status {
	case reconciler.ClusterStatusReconciling, reconciler.ClusterStatusPending:
		return operation, 30 * time.Second, nil
	case reconciler.ClusterStatusReady:
		return operation, 0, nil
	case reconciler.ClusterStatusError:
		log.Warnf("Reconciler failed")
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("Reconciler failed"), log)
	default:
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("unknown cluster status: %s", state.Status), log)
	}
}
