package systemd

import (
	"context"
	"fmt"
	"log/slog"
)

type Systemd interface {
	Enable(ctx context.Context, service string) error
	Disable(ctx context.Context, service string) error
	IsEnabled(ctx context.Context, service string) (bool, error)
	Close()
}

type SystemdConnector struct {
	dbusConnection DbusConnector
	logger         *slog.Logger
}

type SystemdConnectorOption func(*SystemdConnector)

type SystemdLoader interface {
	NewSystemd(ctx context.Context, logger *slog.Logger, options ...SystemdConnectorOption) (Systemd, error)
}

type defaultSystemdLoader struct{}

func (d *defaultSystemdLoader) NewSystemd(ctx context.Context, logger *slog.Logger, options ...SystemdConnectorOption) (Systemd, error) {
	return NewSystemd(ctx, logger, options...)
}

func NewDefaultSystemdLoader() SystemdLoader {
	return &defaultSystemdLoader{}
}

func WithCustomDbusConnector(dbusConnection DbusConnector) SystemdConnectorOption {
	return func(s *SystemdConnector) {
		s.dbusConnection = dbusConnection
	}
}

func NewSystemd(ctx context.Context, logger *slog.Logger, options ...SystemdConnectorOption) (Systemd, error) {
	systemdInstance := &SystemdConnector{
		logger: logger,
	}

	for _, opt := range options {
		opt(systemdInstance)
	}

	if systemdInstance.dbusConnection != nil {
		return systemdInstance, nil
	}

	dbusConnection, err := NewDbusConnector(ctx)
	if err != nil {
		logger.Error("failed to create dbus connection", "error", err)
		return nil, err
	}
	systemdInstance.dbusConnection = dbusConnection

	return systemdInstance, nil
}

func (s *SystemdConnector) Enable(ctx context.Context, service string) error {
	_, _, err := s.dbusConnection.EnableUnitFilesContext(ctx, []string{service}, false, true)
	if err != nil {
		s.logger.Error("failed to enable service", "service", service, "error", err)
		return fmt.Errorf("failed to enable service %s: %w", service, err)
	}

	return s.reload(ctx, service)
}

func (s *SystemdConnector) Disable(ctx context.Context, service string) error {
	_, err := s.dbusConnection.DisableUnitFilesContext(ctx, []string{service}, false)
	if err != nil {
		s.logger.Error("failed to disable service", "service", service, "error", err)
		return fmt.Errorf("failed to disable service %s: %w", service, err)
	}

	return s.reload(ctx, service)
}

func (s *SystemdConnector) IsEnabled(ctx context.Context, service string) (bool, error) {
	unitFileState, err := s.dbusConnection.GetUnitPropertyContext(ctx, service, "UnitFileState")
	if err != nil {
		s.logger.Error("failed to get unit file state for service", "service", service, "error", err)
		return false, fmt.Errorf("failed to get unit file state for service %s: %w", service, err)
	}

	return unitFileState.Value.Value().(string) == "enabled", nil
}

func (s *SystemdConnector) Close() {
	s.dbusConnection.Close()
}

func (s *SystemdConnector) reload(ctx context.Context, service string) error {
	err := s.dbusConnection.ReloadContext(ctx)
	if err != nil {
		s.logger.Error("failed to reload service", "service", service, "error", err)
		return fmt.Errorf("failed to reload service %s: %w", service, err)
	}
	return nil
}
