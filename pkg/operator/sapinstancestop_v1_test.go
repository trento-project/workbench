package operator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/workbench/internal/sapcontrol"
	"github.com/trento-project/workbench/internal/sapcontrol/mocks"
	"github.com/trento-project/workbench/pkg/operator"
)

type SAPInstanceStopOperatorTestSuite struct {
	suite.Suite
	mockSapcontrol *mocks.MockSAPControlConnector
}

func TestSAPInstanceStopOperator(t *testing.T) {
	suite.Run(t, new(SAPInstanceStopOperatorTestSuite))
}

func (suite *SAPInstanceStopOperatorTestSuite) SetupTest() {
	suite.mockSapcontrol = mocks.NewMockSAPControlConnector(suite.T())
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopInstanceNumberMissing() {
	ctx := context.Background()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument instance_number not provided, could not use the operator", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopInstanceNumberInvalid() {
	ctx := context.Background()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": 0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_number argument as string, argument provided: 0", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopTimeoutInvalid() {
	ctx := context.Background()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
			"timeout":         "value",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse timeout argument as a number, argument provided: value", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopPlanError() {
	ctx := context.Background()

	suite.mockSapcontrol.On(
		"GetProcessListContext",
		ctx,
		mock.Anything,
	).Return(nil, errors.New("error getting processes")).
		Once()

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
			"timeout":         300.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("error checking processes state: error getting instance process list: error getting processes", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopCommitAlreadyStopped() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	suite.mockSapcontrol.On(
		"GetProcessListContext",
		ctx,
		mock.Anything,
	).Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Dispstatus: &gray,
			},
			{
				Dispstatus: &gray,
			},
		},
	}, nil)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"stopped":true}`,
		"after":  `{"stopped":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopCommitStoppingError() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.On(
		"GetProcessListContext",
		ctx,
		mock.Anything,
	).Return(
		&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
			},
		}, nil,
	).Once().On(
		"GetProcessListContext",
		mock.Anything,
		mock.Anything,
	).Return(
		&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
			},
		}, nil,
	).On(
		"StopContext",
		ctx,
		mock.Anything,
	).Return(
		nil, errors.New("error stopping"),
	).On(
		"StartContext",
		ctx,
		mock.Anything,
	).Return(nil, nil)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("error stopping instance: error stopping", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopCommitStoppingTimeout() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.On(
		"GetProcessListContext",
		mock.Anything,
		mock.Anything,
	).Return(
		&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
			},
		}, nil,
	).On(
		"StopContext",
		ctx,
		mock.Anything,
	).Return(
		nil, nil,
	).On(
		"StartContext",
		ctx,
		mock.Anything,
	).Return(nil, nil)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
			"timeout":         0.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopInterval(0 * time.Second)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("error waiting until instance is in desired state\n"+
		"error waiting until instance is in desired state", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopVerifyError() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY
	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.On(
		"GetProcessListContext",
		ctx,
		mock.Anything,
	).Return(
		&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
			},
		}, nil,
	).Once().On(
		"StopContext",
		ctx,
		mock.Anything,
	).Return(
		nil, nil,
	).On(
		"GetProcessListContext",
		mock.Anything,
		mock.Anything,
	).Return(
		&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &gray,
				},
			},
		}, nil,
	).Once().On(
		"GetProcessListContext",
		ctx,
		mock.Anything,
	).Return(
		nil, errors.New("error getting processes in verify"),
	).Once().On(
		"StartContext",
		ctx,
		mock.Anything,
	).Return(nil, nil).On(
		"GetProcessListContext",
		mock.Anything,
		mock.Anything,
	).Return(
		&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
			},
		}, nil,
	)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("error checking processes state: error getting instance process list: error getting processes in verify", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopRollbackStartingError() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.On(
		"GetProcessListContext",
		ctx,
		mock.Anything,
	).Return(
		&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
			},
		}, nil,
	).On(
		"StopContext",
		ctx,
		mock.Anything,
	).Return(
		nil, errors.New("error starting"),
	).On(
		"StartContext",
		ctx,
		mock.Anything,
	).Return(nil, errors.New("error starting"))

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("error starting instance: error starting\nerror stopping instance: error starting", report.Error.Message)
}

func (suite *SAPInstanceStopOperatorTestSuite) TestSAPInstanceStopSuccess() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN
	gray := sapcontrol.STATECOLORSAPControlGRAY

	suite.mockSapcontrol.On(
		"GetProcessListContext",
		ctx,
		mock.Anything,
	).Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Dispstatus: &green,
			},
		},
	}, nil).Once().On(
		"GetProcessListContext",
		mock.Anything,
		mock.Anything,
	).Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Dispstatus: &gray,
			},
		},
	}, nil).On(
		"StopContext",
		ctx,
		mock.Anything,
	).Return(
		nil, nil,
	)

	sapInstanceStopOperator := operator.NewSAPInstanceStop(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStop]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStop]{
				operator.Option[operator.SAPInstanceStop](operator.WithCustomStopSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStopOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"stopped":false}`,
		"after":  `{"stopped":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
