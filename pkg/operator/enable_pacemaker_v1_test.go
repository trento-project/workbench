package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/systemd/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

type EnablePacemakerOperatorTestSuite struct {
	suite.Suite
	logger            *logrus.Logger
	loggerEntry       *logrus.Entry
	mockSystemd       *mocks.MockSystemd
	mockSystemdLoader *mocks.MockSystemdLoader
}

func buildEnablePacemakerOperator(suite *EnablePacemakerOperatorTestSuite) operator.Operator {
	return operator.NewEnablePacemaker(
		operator.OperatorArguments{},
		"test-op",
		operator.OperatorOptions[operator.EnablePacemaker]{
			BaseOperatorOptions: []operator.BaseOperatorOption{
				operator.WithCustomLogger(suite.logger),
			},
			OperatorOptions: []operator.Option[operator.EnablePacemaker]{
				operator.Option[operator.EnablePacemaker](operator.WithCustomEnablePacemakerSystemdLoader(suite.mockSystemdLoader)),
			},
		},
	)
}

func TestEnablePacemakerOperator(t *testing.T) {
	suite.Run(t, new(EnablePacemakerOperatorTestSuite))
}

func (suite *EnablePacemakerOperatorTestSuite) SetupTest() {
	suite.logger = logrus.StandardLogger()
	suite.loggerEntry = suite.logger.WithField("operation_id", "test-op")
	suite.mockSystemd = mocks.NewMockSystemd(suite.T())
	suite.mockSystemdLoader = mocks.NewMockSystemdLoader(suite.T())
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorPlanErrorDbusConnection() {
	ctx := context.Background()

	suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(nil, errors.New("dbus connection error")).
		Once()

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.EqualValues("unable to initialize systemd connector: dbus connection error", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorPlanErrorIsEnabled() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("systemd error")).
		Once().
		NotBefore(systemdLoaderCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.EqualValues("failed to check if pacemaker service is enabled: systemd error", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorPlanAlreadyEnabled() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(systemdLoaderCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(isEnabledCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"enabled":true}`,
		"after":  `{"enabled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorCommitErrorEnableFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(errors.New("systemd enable error")).
		Once().
		NotBefore(isEnabledCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(enableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd disable error\nfailed to start pacemaker service: systemd enable error", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorCommitErrorEnableSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(errors.New("systemd enable error")).
		Once().
		NotBefore(isEnabledCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(enableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("failed to start pacemaker service: systemd enable error", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorVerifyErrorIsEnabledFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("error verifying is enabled")).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd disable error\nfailed to check if pacemaker service is enabled: error verifying is enabled", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorVerifyErrorIsEnabledSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, errors.New("error verifying is enabled")).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("failed to check if pacemaker service is enabled: error verifying is enabled", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorVerifyNotEnabledFailedRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(errors.New("systemd disable error")).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("systemd disable error\npacemaker service is not enabled", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorVerifyNotEnabledSuccessfulRollback() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(enableCall)

	disableCall := suite.mockSystemd.On("Disable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(verifyIsEnabledCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(disableCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("pacemaker service is not enabled", report.Error.Message)
}

func (suite *EnablePacemakerOperatorTestSuite) TestEnablePacemakerOperatorSuccess() {
	ctx := context.Background()

	systemdLoaderCall := suite.mockSystemdLoader.On("NewSystemd", ctx, suite.loggerEntry).
		Return(suite.mockSystemd, nil).
		Once()

	isEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(false, nil).
		Once().
		NotBefore(systemdLoaderCall)

	enableCall := suite.mockSystemd.On("Enable", ctx, "pacemaker.service").
		Return(nil).
		Once().
		NotBefore(isEnabledCall)

	verifyIsEnabledCall := suite.mockSystemd.On("IsEnabled", ctx, "pacemaker.service").
		Return(true, nil).
		Once().
		NotBefore(enableCall)

	suite.mockSystemd.On("Close").
		Return().
		Once().
		NotBefore(verifyIsEnabledCall)

	report := buildEnablePacemakerOperator(suite).Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"enabled":false}`,
		"after":  `{"enabled":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
