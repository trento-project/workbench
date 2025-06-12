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

type pacemakerEnablementDiffOutput struct {
	Enabled bool `json:"enabled"`
}

type PacemakerEnableOption Option[PacemakerEnable]

// PacemakerEnable operator enables Pacemaker systemd unit.
//
// # Execution Phases
//
// - PLAN:
//   The operator connects to systemd and determines if the Pacemaker service is enabled.
//   The operation is skipped if the Pacemaker service is already enabled.
//
// - COMMIT:
//   It enables the Pacemaker systemd unit.
//
// - VERIFY:
//   The operator checks if the Pacemaker service is enabled after the commit phase.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the Pacemaker service is disabled back again.

type PacemakerEnable struct {
	baseOperator
	systemdLoader    systemd.SystemdLoader
	systemdConnector systemd.Systemd
}

func WithCustomPacemakerEnableSystemdLoader(systemdLoader systemd.SystemdLoader) PacemakerEnableOption {
	return func(o *PacemakerEnable) {
		o.systemdLoader = systemdLoader
	}
}

func NewPacemakerEnable(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[PacemakerEnable],
) *Executor {
	pacemakerEnable := &PacemakerEnable{
		baseOperator:  newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		systemdLoader: systemd.NewDefaultSystemdLoader(),
	}

	for _, opt := range options.OperatorOptions {
		opt(pacemakerEnable)
	}

	return &Executor{
		phaser:      pacemakerEnable,
		operationID: operationID,
	}
}

func (p *PacemakerEnable) plan(ctx context.Context) (bool, error) {
	if p.systemdConnector == nil {
		systemdConnector, err := p.systemdLoader.NewSystemd(ctx, p.logger)
		if err != nil {
			p.logger.Errorf("unable to initialize systemd connector: %s", err)
			return false, fmt.Errorf("unable to initialize systemd connector: %w", err)
		}
		p.systemdConnector = systemdConnector
	}

	pacemakerEnabled, err := p.systemdConnector.IsEnabled(ctx, pacemakerServiceName)
	if err != nil {
		p.logger.Error("failed to check if pacemaker service is enabled", "error", err)
		return false, fmt.Errorf("failed to check if pacemaker service is enabled: %w", err)
	}

	p.resources[beforeDiffField] = pacemakerEnabled

	if pacemakerEnabled {
		p.logger.Info("pacemaker service already enabled, skipping operation")
		p.resources[afterDiffField] = pacemakerEnabled
		return true, nil
	}

	return false, nil
}

func (p *PacemakerEnable) commit(ctx context.Context) error {
	if err := p.systemdConnector.Enable(ctx, pacemakerServiceName); err != nil {
		p.logger.Error("failed to start pacemaker service", "error", err)
		return fmt.Errorf("failed to start pacemaker service: %w", err)
	}
	p.logger.Info("Pacemaker service enabled successfully")
	return nil
}

func (p *PacemakerEnable) verify(ctx context.Context) error {
	pacemakerEnabled, err := p.systemdConnector.IsEnabled(ctx, pacemakerServiceName)
	if err != nil {
		p.logger.Error("failed to check if pacemaker service is enabled", "error", err)
		return fmt.Errorf("failed to check if pacemaker service is enabled: %w", err)
	}

	if !pacemakerEnabled {
		p.logger.Info("pacemaker service is not enabled, rolling back")
		return fmt.Errorf("pacemaker service is not enabled")
	}

	p.resources[afterDiffField] = pacemakerEnabled

	return nil
}

func (p *PacemakerEnable) rollback(ctx context.Context) error {
	return p.systemdConnector.Disable(ctx, pacemakerServiceName)
}

func (p *PacemakerEnable) operationDiff(ctx context.Context) map[string]any {
	return computeOperationDiff(p.resources)
}

func (p *PacemakerEnable) after(_ context.Context) {
	p.systemdConnector.Close()
}

func computeOperationDiff(resources map[string]any) map[string]any {
	diff := make(map[string]any)

	beforeDiffOutput := pacemakerEnablementDiffOutput{
		Enabled: resources[beforeDiffField].(bool),
	}
	before, _ := json.Marshal(beforeDiffOutput)
	diff[beforeDiffField] = string(before)

	afterDiffOutput := pacemakerEnablementDiffOutput{
		Enabled: resources[afterDiffField].(bool),
	}
	after, _ := json.Marshal(afterDiffOutput)
	diff[afterDiffField] = string(after)

	return diff
}
