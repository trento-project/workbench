package dbus

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
)

// Connector acts as an abstract interface for the dbus functionalities exposed by the package "github.com/coreos/go-systemd/v22/dbus"
type Connector interface {
	GetUnitPropertyContext(ctx context.Context, unit string, propertyName string) (*dbus.Property, error)
	EnableUnitFilesContext(ctx context.Context, files []string, runtime bool, force bool) (bool, []dbus.EnableUnitFileChange, error)
	DisableUnitFilesContext(ctx context.Context, files []string, runtime bool) ([]dbus.DisableUnitFileChange, error)
	ReloadContext(ctx context.Context) error
	ListJobsContext(ctx context.Context) ([]dbus.JobStatus, error)
	ListUnitsContext(ctx context.Context) ([]dbus.UnitStatus, error)
	// NewWithContext establishes a connection to any available bus and authenticates.
	// Callers should call Close() when done with the connection.
	// see https://pkg.go.dev/github.com/coreos/go-systemd/v22@v22.5.0/dbus#NewWithContext
	Close()
}

func NewConnector(ctx context.Context) (Connector, error) {
	// the created connection does implement the DbusConnector interface, hence it can be returned as such
	dbusConnection, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return dbusConnection, nil
}
