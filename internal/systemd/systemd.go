package systemd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/trento-project/workbench/internal/dbus"
)

type Systemd interface {
	Enable(ctx context.Context, service string) error
	Disable(ctx context.Context, service string) error
	IsEnabled(ctx context.Context, service string) (bool, error)
	Close()
}

type Connector struct {
	dbusConnection dbus.Connector
	logger         *slog.Logger
}

type ConnectorOption func(*Connector)

type Loader interface {
	NewSystemd(ctx context.Context, logger *slog.Logger, options ...ConnectorOption) (Systemd, error)
}

type defaultSystemdLoader struct{}

func (d *defaultSystemdLoader) NewSystemd(
	ctx context.Context,
	logger *slog.Logger,
	options ...ConnectorOption,
) (Systemd, error) {
	return NewSystemd(ctx, logger, options...)
}

func NewDefaultSystemdLoader() Loader {
	return &defaultSystemdLoader{}
}

func WithCustomDbusConnector(dbusConnection dbus.Connector) ConnectorOption {
	return func(s *Connector) {
		s.dbusConnection = dbusConnection
	}
}

func NewSystemd(ctx context.Context, logger *slog.Logger, options ...ConnectorOption) (Systemd, error) {
	systemdInstance := &Connector{
		logger: logger,
	}

	for _, opt := range options {
		opt(systemdInstance)
	}

	if systemdInstance.dbusConnection != nil {
		return systemdInstance, nil
	}

	dbusConnection, err := dbus.NewConnector(ctx)
	if err != nil {
		logger.Error("failed to create dbus connection", "error", err)
		return nil, err
	}
	systemdInstance.dbusConnection = dbusConnection

	return systemdInstance, nil
}

func (s *Connector) Enable(ctx context.Context, service string) error {
	_, _, err := s.dbusConnection.EnableUnitFilesContext(ctx, []string{service}, false, true)
	if err != nil {
		s.logger.Error("failed to enable service", "service", service, "error", err)
		return fmt.Errorf("failed to enable service %s: %w", service, err)
	}

	return s.reload(ctx, service)
}

func (s *Connector) Disable(ctx context.Context, service string) error {
	_, err := s.dbusConnection.DisableUnitFilesContext(ctx, []string{service}, false)
	if err != nil {
		s.logger.Error("failed to disable service", "service", service, "error", err)
		return fmt.Errorf("failed to disable service %s: %w", service, err)
	}

	return s.reload(ctx, service)
}

func (s *Connector) IsEnabled(ctx context.Context, service string) (bool, error) {
	unitFileState, err := s.dbusConnection.GetUnitPropertyContext(ctx, service, "UnitFileState")
	if err != nil {
		s.logger.Error("failed to get unit file state for service", "service", service, "error", err)
		return false, fmt.Errorf("failed to get unit file state for service %s: %w", service, err)
	}

	return unitFileState.Value.Value().(string) == "enabled", nil
}

func (s *Connector) Close() {
	s.dbusConnection.Close()
}

func (s *Connector) reload(ctx context.Context, service string) error {
	err := s.dbusConnection.ReloadContext(ctx)
	if err != nil {
		s.logger.Error("failed to reload service", "service", service, "error", err)
		return fmt.Errorf("failed to reload service %s: %w", service, err)
	}
	return nil
}
