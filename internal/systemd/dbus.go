package systemd

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
)

// DbusConnector acts as an abstract interface for the dbus functionalities exposed by the package "github.com/coreos/go-systemd/v22/dbus"
type DbusConnector interface {
	GetUnitPropertyContext(ctx context.Context, unit string, propertyName string) (*dbus.Property, error)
	EnableUnitFilesContext(ctx context.Context, files []string, runtime bool, force bool) (bool, []dbus.EnableUnitFileChange, error)
	DisableUnitFilesContext(ctx context.Context, files []string, runtime bool) ([]dbus.DisableUnitFileChange, error)
	ReloadContext(ctx context.Context) error
	// NewWithContext establishes a connection to any available bus and authenticates.
	// Callers should call Close() when done with the connection.
	// see https://pkg.go.dev/github.com/coreos/go-systemd/v22@v22.5.0/dbus#NewWithContext
	Close()
}

func NewDbusConnector(ctx context.Context) (DbusConnector, error) {
	// the created connection does implement the DbusConnector interface, hence it can be returned as such
	dbusConnection, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return dbusConnection, nil
}
