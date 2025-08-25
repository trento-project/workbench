package operator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trento-project/workbench/internal/systemd"
)

const (
	PacemakerEnableOperatorName = "pacemakerenable"
	pacemakerServiceName        = "pacemaker.service"
)

type serviceEnablementDiffOutput struct {
	Enabled bool `json:"enabled"`
}

type ServiceEnableOption Option[ServiceEnable]

// ServiceEnable operator enables a systemd unit.
//
// # Execution Phases
//
// - PLAN:
//   The operator connects to systemd and determines if the service is enabled.
//   The operation is skipped if the service is already enabled.
//
// - COMMIT:
//   It enables the systemd unit.
//
// - VERIFY:
//   The operator checks if the service is enabled after the commit phase.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the service is disabled back again.

type ServiceEnable struct {
	baseOperator
	systemdLoader    systemd.SystemdLoader
	systemdConnector systemd.Systemd
	service          string
}

func WithCustomServiceEnableSystemdLoader(systemdLoader systemd.SystemdLoader) ServiceEnableOption {
	return func(se *ServiceEnable) {
		se.systemdLoader = systemdLoader
	}
}

func WithServiceToEnable(service string) ServiceEnableOption {
	return func(se *ServiceEnable) {
		se.service = service
	}
}

func NewServiceEnable(
	name string,
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[ServiceEnable],
) *Executor {
	serviceEnable := &ServiceEnable{
		baseOperator: newBaseOperator(
			name, operationID, arguments, options.BaseOperatorOptions...,
		),
		systemdLoader: systemd.NewDefaultSystemdLoader(),
	}

	for _, opt := range options.OperatorOptions {
		opt(serviceEnable)
	}

	return &Executor{
		phaser:      serviceEnable,
		operationID: operationID,
		logger:      serviceEnable.logger,
	}
}

func (se *ServiceEnable) plan(ctx context.Context) (bool, error) {
	systemdConnector, err := se.systemdLoader.NewSystemd(ctx, se.logger)
	if err != nil {
		se.logger.Error("unable to initialize systemd connector", "error", err)
		return false, fmt.Errorf("unable to initialize systemd connector: %w", err)
	}
	se.systemdConnector = systemdConnector

	serviceEnabled, err := se.systemdConnector.IsEnabled(ctx, se.service)
	if err != nil {
		se.logger.Error("failed to check if service is enabled", "service", se.service, "error", err)
		return false, fmt.Errorf("failed to check if %s service is enabled: %w", se.service, err)
	}

	se.resources[beforeDiffField] = serviceEnabled

	if serviceEnabled {
		se.logger.Info("service already enabled, skipping operation", "service", se.service)
		se.resources[afterDiffField] = serviceEnabled
		return true, nil
	}

	return false, nil
}

func (se *ServiceEnable) commit(ctx context.Context) error {
	if err := se.systemdConnector.Enable(ctx, se.service); err != nil {
		se.logger.Error("failed to enable service", "service", se.service, "error", err)
		return fmt.Errorf("failed to enable service %s: %w", se.service, err)
	}
	return nil
}

func (se *ServiceEnable) verify(ctx context.Context) error {
	serviceEnabled, err := se.systemdConnector.IsEnabled(ctx, se.service)
	if err != nil {
		se.logger.Error("failed to check if service is enabled", "service", se.service, "error", err)
		return fmt.Errorf("failed to check if service %s is enabled: %w", se.service, err)
	}

	if !serviceEnabled {
		se.logger.Info("service %s is not enabled, rolling back", "service", se.service)
		return fmt.Errorf("service %s is not enabled", se.service)
	}

	se.resources[afterDiffField] = serviceEnabled

	return nil
}

func (se *ServiceEnable) rollback(ctx context.Context) error {
	return se.systemdConnector.Disable(ctx, se.service)
}

func (se *ServiceEnable) operationDiff(_ context.Context) map[string]any {
	return computeOperationDiff(se.resources)
}

func (se *ServiceEnable) after(_ context.Context) {
	se.systemdConnector.Close()
}

func computeOperationDiff(resources map[string]any) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := serviceEnablementDiffOutput{
		Enabled: resources[beforeDiffField].(bool),
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff[beforeDiffField] = string(before)

	afterDiffOutput := serviceEnablementDiffOutput{
		Enabled: resources[afterDiffField].(bool),
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff[afterDiffField] = string(after)

	return diff
}
