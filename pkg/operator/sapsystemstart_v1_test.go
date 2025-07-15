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

type SAPSystemStartOperatorTestSuite struct {
	suite.Suite
	mockSapcontrol *mocks.MockSAPControlConnector
}

func TestSAPSystemStartOperator(t *testing.T) {
	suite.Run(t, new(SAPSystemStartOperatorTestSuite))
}

func (suite *SAPSystemStartOperatorTestSuite) SetupTest() {
	suite.mockSapcontrol = mocks.NewMockSAPControlConnector(suite.T())
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceNumberMissing() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument instance_number not provided, could not use the operator", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceNumberInvalid() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": 0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_number argument as string, argument provided: 0", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartTimeoutInvalid() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
			"timeout":         "value",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse timeout argument as a number, argument provided: value", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceTypeInvalid() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
			"instance_type":   0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_type argument as a string, argument provided: 0", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartInstanceTypeUnknown() {
	ctx := context.Background()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
			"instance_type":   "unknown",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("invalid instance_type value: unknown", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartPlanError() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(nil, errors.New("error getting instances")).
		Once()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
			"timeout":         300.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("error checking system state: error getting instance list: error getting instances", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartCommitAlreadyStarted() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(&sapcontrol.GetSystemInstanceListResponse{
			Instances: []*sapcontrol.SAPInstance{
				{
					Dispstatus: &green,
				},
				{
					Dispstatus: &green,
				},
			},
		}, nil).
		Once()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"started":true}`,
		"after":  `{"started":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartCommitAlreadyStartedFiltered() {
	cases := []struct {
		instanceType string
		features     string
	}{
		{
			instanceType: "abap",
			features:     "ABAP|GATEWAY|ICMAN|IGS",
		},
		{
			instanceType: "j2ee",
			features:     "J2EE|IGS",
		},
		{
			instanceType: "scs",
			features:     "MESSAGESERVER|ENQUE",
		},
		{
			instanceType: "enqrep",
			features:     "ENQREP",
		},
	}

	for _, tt := range cases {
		ctx := context.Background()
		suite.mockSapcontrol = mocks.NewMockSAPControlConnector(suite.T())

		green := sapcontrol.STATECOLORSAPControlGREEN
		gray := sapcontrol.STATECOLORSAPControlGRAY

		suite.mockSapcontrol.
			On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
			Return(&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Features:   "Other",
						Dispstatus: &gray,
					},
					{
						Features:   tt.features,
						Dispstatus: &green,
					},
				},
			}, nil).
			Once()

		sapSystemStartOperator := operator.NewSAPSystemStart(
			operator.OperatorArguments{
				"instance_number": "00",
				"instance_type":   tt.instanceType,
			},
			"test-op",
			operator.OperatorOptions[operator.SAPSystemStart]{
				OperatorOptions: []operator.Option[operator.SAPSystemStart]{
					operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				},
			},
		)

		report := sapSystemStartOperator.Run(ctx)

		expectedDiff := map[string]any{
			"before": `{"started":true}`,
			"after":  `{"started":true}`,
		}

		suite.Nil(report.Error)
		suite.Equal(operator.PLAN, report.Success.LastPhase)
		suite.EqualValues(expectedDiff, report.Success.Diff)
	}
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartCommitStartingError() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once()

	suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting")).
		NotBefore(planGetInstances).
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetInstances).
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once().
		NotBefore(planGetInstances)

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("error starting system: error starting", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartVerifyError() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once()

	verifyGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(nil, errors.New("error getting instances in verify")).
		Once().
		NotBefore(planGetInstances)

	suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetInstances).
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(verifyGetInstances).
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		NotBefore(verifyGetInstances)

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("verify system started failed: error getting instance list: error getting instances in verify", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartVerifyTimeout() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Times(3).
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil)

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
			"timeout":         0.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemInterval(0 * time.Second)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues(
		"rollback to stopped failed: error waiting until system is in desired state\n"+
			"verify system started failed: "+
			"error waiting until system is in desired state", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartRollbackStoppingError() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
						Features:   "ABAP|GATEWAY|ICMAN|IGS",
					},
				},
			}, nil,
		).
		Once().
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, errors.New("error starting")).
		On(
			"StopSystemContext",
			ctx,
			mock.MatchedBy(func(req *sapcontrol.StopSystem) bool {
				return *req.Options == sapcontrol.StartStopOptionSAPControlABAPINSTANCES
			}),
		).
		Return(nil, errors.New("error stopping"))

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
			"instance_type":   "abap",
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("error stopping system: error stopping\nerror starting system: error starting", report.Error.Message)
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartSuccess() {
	cases := []struct {
		instanceType string
		features     string
		options      sapcontrol.StartStopOption
	}{
		{
			instanceType: "all",
			features:     "OTHER",
			options:      sapcontrol.StartStopOptionSAPControlALLINSTANCES,
		},
		{
			instanceType: "abap",
			features:     "ABAP|GATEWAY|ICMAN|IGS",
			options:      sapcontrol.StartStopOptionSAPControlABAPINSTANCES,
		},
		{
			instanceType: "j2ee",
			features:     "J2EE|IGS",
			options:      sapcontrol.StartStopOptionSAPControlJ2EEINSTANCES,
		},
		{
			instanceType: "scs",
			features:     "MESSAGESERVER|ENQUE",
			options:      sapcontrol.StartStopOptionSAPControlSCSINSTANCES,
		},
		{
			instanceType: "enqrep",
			features:     "ENQREP",
			options:      sapcontrol.StartStopOptionSAPControlENQREPINSTANCES,
		},
	}

	for _, tt := range cases {
		ctx := context.Background()
		suite.mockSapcontrol = mocks.NewMockSAPControlConnector(suite.T())

		green := sapcontrol.STATECOLORSAPControlGREEN
		gray := sapcontrol.STATECOLORSAPControlGRAY

		planGetInstances := suite.mockSapcontrol.
			On("GetSystemInstanceListContext", ctx, mock.Anything).
			Return(&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
						Features:   tt.features,
					},
				},
			}, nil).
			Once()

		suite.mockSapcontrol.
			On(
				"StartSystemContext",
				ctx,
				mock.MatchedBy(func(req *sapcontrol.StartSystem) bool {
					return *req.Options == tt.options
				}),
			).
			Return(nil, nil).
			NotBefore(planGetInstances).
			On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
			Return(&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &green,
						Features:   tt.features,
					},
				},
			}, nil).
			Once().
			NotBefore(planGetInstances)

		sapSystemStartOperator := operator.NewSAPSystemStart(
			operator.OperatorArguments{
				"instance_number": "00",
				"instance_type":   tt.instanceType,
			},
			"test-op",
			operator.OperatorOptions[operator.SAPSystemStart]{
				OperatorOptions: []operator.Option[operator.SAPSystemStart]{
					operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				},
			},
		)

		report := sapSystemStartOperator.Run(ctx)

		expectedDiff := map[string]any{
			"before": `{"started":false}`,
			"after":  `{"started":true}`,
		}

		suite.Nil(report.Error)
		suite.Equal(operator.VERIFY, report.Success.LastPhase)
		suite.EqualValues(expectedDiff, report.Success.Diff)
	}
}

func (suite *SAPSystemStartOperatorTestSuite) TestSAPSystemStartSuccessMultipleQueries() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY
	green := sapcontrol.STATECOLORSAPControlGREEN

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Once()

	suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetInstances)

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
					},
				},
			}, nil,
		).
		Times(3).
		NotBefore(planGetInstances).
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &green,
					},
				},
			}, nil,
		).
		Once()

	sapSystemStartOperator := operator.NewSAPSystemStart(
		operator.OperatorArguments{
			"instance_number": "00",
			"timeout":         5.0,
		},
		"test-op",
		operator.OperatorOptions[operator.SAPSystemStart]{
			OperatorOptions: []operator.Option[operator.SAPSystemStart]{
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPSystemStart](operator.WithCustomStartSystemInterval(0 * time.Second)),
			},
		},
	)

	report := sapSystemStartOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"started":false}`,
		"after":  `{"started":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
