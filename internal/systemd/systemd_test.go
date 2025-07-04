package systemd_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/coreos/go-systemd/v22/dbus"
	innerDbus "github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/support"
	"github.com/trento-project/workbench/internal/systemd"
	"github.com/trento-project/workbench/internal/systemd/mocks"
)

type SystemdTestSuite struct {
	suite.Suite
	dbusMock *mocks.MockDbusConnector
	logger   *slog.Logger
}

func TestSystemdClient(t *testing.T) {
	suite.Run(t, new(SystemdTestSuite))
}

func (suite *SystemdTestSuite) SetupTest() {
	suite.dbusMock = mocks.NewMockDbusConnector(suite.T())
	suite.logger = support.NewDefaultLogger(slog.LevelInfo).With("test", "systemd_test_suite")
}

func (suite *SystemdTestSuite) TestServiceIsEnabledFailure() {
	ctx := context.Background()

	suite.dbusMock.On(
		"GetUnitPropertyContext",
		ctx,
		"foo.service",
		"UnitFileState",
	).Return(
		nil,
		errors.New("exit status 1"),
	).Once()

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	enabled, err := systemdConnector.IsEnabled(ctx, "foo.service")

	suite.Error(err)
	suite.False(enabled)
	suite.ErrorContains(err, "failed to get unit file state for service foo.service: exit status 1")
}

func (suite *SystemdTestSuite) TestServiceIsEnabled() {
	ctx := context.Background()

	property := &dbus.Property{
		Name:  "UnitFileState",
		Value: innerDbus.MakeVariant("enabled"),
	}

	suite.dbusMock.On(
		"GetUnitPropertyContext",
		ctx,
		"foo.service",
		"UnitFileState",
	).Return(property, nil).
		Once()

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	enabled, err := systemdConnector.IsEnabled(ctx, "foo.service")

	suite.NoError(err)
	suite.True(enabled)
}

func (suite *SystemdTestSuite) TestServiceIsDisabled() {
	ctx := context.Background()

	property := &dbus.Property{
		Name:  "UnitFileState",
		Value: innerDbus.MakeVariant("disabled"),
	}

	suite.dbusMock.On(
		"GetUnitPropertyContext",
		ctx,
		"foo.service",
		"UnitFileState",
	).Return(property, nil).
		Once()

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	enabled, err := systemdConnector.IsEnabled(ctx, "foo.service")

	suite.NoError(err)
	suite.False(enabled)
}

func (suite *SystemdTestSuite) TestEnableServiceFailure() {
	ctx := context.Background()

	suite.dbusMock.On(
		"EnableUnitFilesContext",
		ctx,
		[]string{"foo.service"},
		false,
		true,
	).Return(
		false,
		[]dbus.EnableUnitFileChange{},
		errors.New("exit status 1"),
	).Once()

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	err := systemdConnector.Enable(ctx, "foo.service")

	suite.Error(err)
	suite.ErrorContains(err, "failed to enable service foo.service: exit status 1")
}

func (suite *SystemdTestSuite) TestEnableServiceFailureOnReload() {
	ctx := context.Background()

	enableCall := suite.dbusMock.On(
		"EnableUnitFilesContext",
		ctx,
		[]string{"foo.service"},
		false,
		true,
	).Return(
		true,
		[]dbus.EnableUnitFileChange{},
		nil,
	).Once()

	suite.dbusMock.On(
		"ReloadContext",
		ctx,
	).Return(errors.New("exit status 1")).
		Once().
		NotBefore(enableCall)

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	err := systemdConnector.Enable(ctx, "foo.service")

	suite.Error(err)
	suite.ErrorContains(err, "failed to reload service foo.service: exit status 1")
}

func (suite *SystemdTestSuite) TestSuccessfulEnableService() {
	ctx := context.Background()

	enableCall := suite.dbusMock.On(
		"EnableUnitFilesContext",
		ctx,
		[]string{"foo.service"},
		false,
		true,
	).Return(
		true,
		[]dbus.EnableUnitFileChange{},
		nil,
	).Once()

	suite.dbusMock.On(
		"ReloadContext",
		ctx,
	).Return(nil).
		Once().
		NotBefore(enableCall)

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	err := systemdConnector.Enable(ctx, "foo.service")

	suite.NoError(err)
}

func (suite *SystemdTestSuite) TestDisableServiceFailure() {
	ctx := context.Background()

	suite.dbusMock.On(
		"DisableUnitFilesContext",
		ctx,
		[]string{"foo.service"},
		false,
	).Return(
		[]dbus.DisableUnitFileChange{},
		errors.New("exit status 1"),
	).Once()

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	err := systemdConnector.Disable(ctx, "foo.service")

	suite.Error(err)
	suite.ErrorContains(err, "failed to disable service foo.service: exit status 1")
}

func (suite *SystemdTestSuite) TestDisableServiceFailureOnReload() {
	ctx := context.Background()

	disableCall := suite.dbusMock.On(
		"DisableUnitFilesContext",
		ctx,
		[]string{"foo.service"},
		false,
	).Return(
		[]dbus.DisableUnitFileChange{},
		nil,
	).Once()

	suite.dbusMock.On(
		"ReloadContext",
		ctx,
	).Return(errors.New("exit status 1")).
		Once().
		NotBefore(disableCall)

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	err := systemdConnector.Disable(ctx, "foo.service")

	suite.Error(err)
	suite.ErrorContains(err, "failed to reload service foo.service: exit status 1")
}

func (suite *SystemdTestSuite) TestSuccessfulDisableService() {
	ctx := context.Background()

	disableCall := suite.dbusMock.On(
		"DisableUnitFilesContext",
		ctx,
		[]string{"foo.service"},
		false,
	).Return(
		[]dbus.DisableUnitFileChange{},
		nil,
	).Once()

	suite.dbusMock.On(
		"ReloadContext",
		ctx,
	).Return(nil).
		Once().
		NotBefore(disableCall)

	systemdConnector, _ := systemd.NewSystemd(
		ctx,
		suite.logger,
		systemd.WithCustomDbusConnector(suite.dbusMock),
	)

	err := systemdConnector.Disable(ctx, "foo.service")

	suite.NoError(err)
}
