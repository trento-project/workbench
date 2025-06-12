package operator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trento-project/workbench/internal/systemd"
)

const (
	EnablePacemakerOperatorName = "enablepacemaker"
	pacemakerServiceName        = "pacemaker.service"
)

type pacemakerEnablementDiffOutput struct {
	Enabled bool `json:"enabled"`
}

type EnablePacemakerOption Option[EnablePacemaker]

// EnablePacemaker operator enables Pacemaker systemd unit.
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

type EnablePacemaker struct {
	baseOperator
	systemdLoader    systemd.SystemdLoader
	systemdConnector systemd.Systemd
}

func WithCustomEnablePacemakerSystemdLoader(systemdLoader systemd.SystemdLoader) EnablePacemakerOption {
	return func(o *EnablePacemaker) {
		o.systemdLoader = systemdLoader
	}
}

func NewEnablePacemaker(
	arguments OperatorArguments,
	operationID string,
	options OperatorOptions[EnablePacemaker],
) *Executor {
	enablePacemaker := &EnablePacemaker{
		baseOperator:  newBaseOperator(operationID, arguments, options.BaseOperatorOptions...),
		systemdLoader: systemd.NewDefaultSystemdLoader(),
	}

	for _, opt := range options.OperatorOptions {
		opt(enablePacemaker)
	}

	return &Executor{
		phaser:      enablePacemaker,
		operationID: operationID,
	}
}

func (s *EnablePacemaker) plan(ctx context.Context) (bool, error) {
	if s.systemdConnector == nil {
		systemdConnector, err := s.systemdLoader.NewSystemd(ctx, s.logger)
		if err != nil {
			s.logger.Errorf("unable to initialize systemd connector: %s", err)
			return false, fmt.Errorf("unable to initialize systemd connector: %w", err)
		}
		s.systemdConnector = systemdConnector
	}

	pacemakerEnabled, err := s.systemdConnector.IsEnabled(ctx, pacemakerServiceName)
	if err != nil {
		s.logger.Error("failed to check if pacemaker service is enabled", "error", err)
		return false, fmt.Errorf("failed to check if pacemaker service is enabled: %w", err)
	}

	s.resources[beforeDiffField] = pacemakerEnabled

	if pacemakerEnabled {
		s.logger.Info("pacemaker service already enabled, skipping operation")
		s.resources[afterDiffField] = pacemakerEnabled
		return true, nil
	}

	return false, nil
}

func (s *EnablePacemaker) commit(ctx context.Context) error {
	if err := s.systemdConnector.Enable(ctx, pacemakerServiceName); err != nil {
		s.logger.Error("failed to start pacemaker service", "error", err)
		return fmt.Errorf("failed to start pacemaker service: %w", err)
	}
	s.logger.Info("Pacemaker service enabled successfully")
	return nil
}

func (s *EnablePacemaker) verify(ctx context.Context) error {
	pacemakerEnabled, err := s.systemdConnector.IsEnabled(ctx, pacemakerServiceName)
	if err != nil {
		s.logger.Error("failed to check if pacemaker service is enabled", "error", err)
		return fmt.Errorf("failed to check if pacemaker service is enabled: %w", err)
	}

	if !pacemakerEnabled {
		s.logger.Info("pacemaker service is not enabled, rolling back")
		return fmt.Errorf("pacemaker service is not enabled")
	}

	s.resources[afterDiffField] = pacemakerEnabled

	return nil
}

func (s *EnablePacemaker) rollback(ctx context.Context) error {
	return s.systemdConnector.Disable(ctx, pacemakerServiceName)
}

func (s *EnablePacemaker) operationDiff(ctx context.Context) map[string]any {
	return computeOperationDiff(s.resources)
}

func (s *EnablePacemaker) after(_ context.Context) {
	s.systemdConnector.Close()
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
