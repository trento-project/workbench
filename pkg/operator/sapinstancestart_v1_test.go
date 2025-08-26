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

type SAPInstanceStartOperatorTestSuite struct {
	suite.Suite
	mockSapcontrol *mocks.MockSAPControlConnector
}

func TestSAPInstanceStartOperator(t *testing.T) {
	suite.Run(t, new(SAPInstanceStartOperatorTestSuite))
}

func (suite *SAPInstanceStartOperatorTestSuite) SetupTest() {
	suite.mockSapcontrol = mocks.NewMockSAPControlConnector(suite.T())
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartInstanceNumberMissing() {
	ctx := context.Background()

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument instance_number not provided, could not use the operator", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartInstanceNumberInvalid() {
	ctx := context.Background()

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": 0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_number argument as string, argument provided: 0", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartTimeoutInvalid() {
	ctx := context.Background()

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         "value",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse timeout argument as a number, argument provided: value", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartPlanError() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(nil, errors.New("error getting processes")).
		Once()

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         300.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("error checking processes state: error getting instance process list: error getting processes", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartCommitAlreadyStarted() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
				{
					Dispstatus: &green,
				},
			},
		}, nil).
		Once()

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"started":true}`,
		"after":  `{"started":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartCommitStartingError() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once()

	suite.mockSapcontrol.
		On("StartContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting")).
		NotBefore(planGetProcesses).
		On("StopContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once().
		NotBefore(planGetProcesses)

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("error starting instance: error starting", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartVerifyError() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once()

	verifyGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(nil, errors.New("error getting processes in verify")).
		Once().
		NotBefore(planGetProcesses)

	suite.mockSapcontrol.
		On("StartContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses).
		On("StopContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(verifyGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		NotBefore(verifyGetProcesses)

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("verify instance started failed: error getting instance process list: error getting processes in verify", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartVerifyTimeout() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Times(3).
		On("StartContext", ctx, mock.Anything).
		Return(nil, nil).
		On("StopContext", ctx, mock.Anything).
		Return(nil, nil)

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         0.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartInterval(0 * time.Second)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues(
		"rollback to stopped failed: error waiting until instance is in desired state\n"+
			"verify instance started failed: "+
			"error waiting until instance is in desired state", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartRollbackStoppingError() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once().
		On("StartContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting")).
		On("StopContext", ctx, mock.Anything).
		Return(nil, errors.New("error stopping"))

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("error stopping instance: error stopping\nerror starting instance: error starting", report.Error.Message)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartSuccess() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN
	gray := sapcontrol.STATECOLORSAPControlGRAY

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &gray,
				},
			},
		}, nil).
		Once()

	suite.mockSapcontrol.
		On("StartContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Dispstatus: &green,
				},
			},
		}, nil).
		Once().
		NotBefore(planGetProcesses)

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"started":false}`,
		"after":  `{"started":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPInstanceStartOperatorTestSuite) TestSAPInstanceStartSuccessMultipleQueries() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY
	green := sapcontrol.STATECOLORSAPControlGREEN

	planGetProcesses := suite.mockSapcontrol.
		On("GetProcessListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once()

	suite.mockSapcontrol.
		On("StartContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetProcesses)

	suite.mockSapcontrol.
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{},
		}, nil).
		Twice().
		NotBefore(planGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Times(3).
		NotBefore(planGetProcesses).
		On("GetProcessListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetProcessListResponse{
				Processes: []*sapcontrol.OSProcess{
					{
						Dispstatus: &green,
					},
				},
			}, nil,
		).
		Once()

	sapInstanceStartOperator := operator.NewSAPInstanceStart(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         5.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPInstanceStart]{
			OperatorOptions: []operator.Option[operator.SAPInstanceStart]{
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPInstanceStart](operator.WithCustomStartInterval(0 * time.Second)),
			},
		},
	)

	report := sapInstanceStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"started":false}`,
		"after":  `{"started":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
