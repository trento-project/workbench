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
	return func(o *ServiceEnable) {
		o.systemdLoader = systemdLoader
	}
}

func WithService(service string) ServiceEnableOption {
	return func(o *ServiceEnable) {
		o.service = service
	}
}

func NewServiceEnable(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[ServiceEnable],
) *Executor {
	serviceEnable := &ServiceEnable{
		baseOperator:  newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		systemdLoader: systemd.NewDefaultSystemdLoader(),
	}

	for _, opt := range options.OperatorOptions {
		opt(serviceEnable)
	}

	return &Executor{
		phaser:      serviceEnable,
		operationID: operationID,
	}
}

func (p *ServiceEnable) plan(ctx context.Context) (bool, error) {
	systemdConnector, err := p.systemdLoader.NewSystemd(ctx, p.logger)
	if err != nil {
		p.logger.Errorf("unable to initialize systemd connector: %v", err)
		return false, fmt.Errorf("unable to initialize systemd connector: %w", err)
	}
	p.systemdConnector = systemdConnector

	serviceEnabled, err := p.systemdConnector.IsEnabled(ctx, p.service)
	if err != nil {
		p.logger.Errorf("failed to check if service %s is enabled: %v", p.service, err)
		return false, fmt.Errorf("failed to check if %s service is enabled: %w", p.service, err)
	}

	p.resources[beforeDiffField] = serviceEnabled

	if serviceEnabled {
		p.logger.Infof("service %s already enabled, skipping operation", p.service)
		p.resources[afterDiffField] = serviceEnabled
		return true, nil
	}

	return false, nil
}

func (p *ServiceEnable) commit(ctx context.Context) error {
	if err := p.systemdConnector.Enable(ctx, p.service); err != nil {
		p.logger.Errorf("failed to enable service %s: %v", p.service, err)
		return fmt.Errorf("failed to enable service %s: %w", p.service, err)
	}
	return nil
}

func (p *ServiceEnable) verify(ctx context.Context) error {
	serviceEnabled, err := p.systemdConnector.IsEnabled(ctx, p.service)
	if err != nil {
		p.logger.Errorf("failed to check if service %s is enabled: %v", p.service, err)
		return fmt.Errorf("failed to check if service %s is enabled: %w", p.service, err)
	}

	if !serviceEnabled {
		p.logger.Infof("service %s is not enabled, rolling back", p.service)
		return fmt.Errorf("service %s is not enabled", p.service)
	}

	p.resources[afterDiffField] = serviceEnabled

	return nil
}

func (p *ServiceEnable) rollback(ctx context.Context) error {
	return p.systemdConnector.Disable(ctx, p.service)
}

func (p *ServiceEnable) operationDiff(ctx context.Context) map[string]any {
	return computeOperationDiff(p.resources)
}

func (p *ServiceEnable) after(_ context.Context) {
	p.systemdConnector.Close()
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
