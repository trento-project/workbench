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

type SAPSystemStopOperatorTestSuite struct {
	suite.Suite
	mockSapcontrol *mocks.MockSAPControlConnector
}

func TestSAPSystemStopOperator(t *testing.T) {
	suite.Run(t, new(SAPSystemStopOperatorTestSuite))
}

func (suite *SAPSystemStopOperatorTestSuite) SetupTest() {
	suite.mockSapcontrol = mocks.NewMockSAPControlConnector(suite.T())
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopInstanceNumberMissing() {
	ctx := context.Background()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{},
		"test-op",
		operator.Options[operator.SAPSystemStop]{},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("argument instance_number not provided, could not use the operator", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopInstanceNumberInvalid() {
	ctx := context.Background()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": 0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_number argument as string, argument provided: 0", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopTimeoutInvalid() {
	ctx := context.Background()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         "value",
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse timeout argument as a number, argument provided: value", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopInstanceTypeInvalid() {
	ctx := context.Background()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
			"instance_type":   0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("could not parse instance_type argument as a string, argument provided: 0", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopInstanceTypeUnknown() {
	ctx := context.Background()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
			"instance_type":   "unknown",
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("invalid instance_type value: unknown", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopPlanError() {
	ctx := context.Background()

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(nil, errors.New("error getting instances")).
		Once()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         300.0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{
			OperatorOptions: []operator.Option[operator.SAPSystemStop]{
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.PLAN, report.Error.ErrorPhase)
	suite.EqualValues("error checking system state: error getting instance list: error getting instances", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopCommitAlreadyStopped() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(&sapcontrol.GetSystemInstanceListResponse{
			Instances: []*sapcontrol.SAPInstance{
				{
					Dispstatus: &gray,
				},
				{
					Dispstatus: &gray,
				},
			},
		}, nil).
		Once()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{
			OperatorOptions: []operator.Option[operator.SAPSystemStop]{
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStopOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"stopped":true}`,
		"after":  `{"stopped":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.PLAN, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopCommitAlreadyStoppedFiltered() {
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
						Dispstatus: &green,
					},
					{
						Features:   tt.features,
						Dispstatus: &gray,
					},
				},
			}, nil).
			Once()

		sapSystemStopOperator := operator.NewSAPSystemStop(
			operator.Arguments{
				"instance_number": "00",
				"instance_type":   tt.instanceType,
			},
			"test-op",
			operator.Options[operator.SAPSystemStop]{
				OperatorOptions: []operator.Option[operator.SAPSystemStop]{
					operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
				},
			},
		)

		report := sapSystemStopOperator.Run(ctx)

		expectedDiff := map[string]any{
			"before": `{"stopped":true}`,
			"after":  `{"stopped":true}`,
		}

		suite.Nil(report.Error)
		suite.Equal(operator.PLAN, report.Success.LastPhase)
		suite.EqualValues(expectedDiff, report.Success.Diff)
	}
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopCommitStoppingError() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
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

	stopSystem := suite.mockSapcontrol.
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, errors.New("error stopping")).
		NotBefore(planGetInstances)

	rollbackStartSystem := suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(stopSystem)

	suite.mockSapcontrol.
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
		Once().
		NotBefore(rollbackStartSystem)

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{
			OperatorOptions: []operator.Option[operator.SAPSystemStop]{
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.COMMIT, report.Error.ErrorPhase)
	suite.EqualValues("error stopping system: error stopping", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopVerifyError() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
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

	stopSystem := suite.mockSapcontrol.
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetInstances)

	verifyGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
		Return(nil, errors.New("error getting instances in verify")).
		Once().
		NotBefore(stopSystem)

	rollbackStartSystem := suite.mockSapcontrol.
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(verifyGetInstances)

	suite.mockSapcontrol.
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
		Once().
		NotBefore(rollbackStartSystem)

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{
			OperatorOptions: []operator.Option[operator.SAPSystemStop]{
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.VERIFY, report.Error.ErrorPhase)
	suite.EqualValues("verify system stopped failed: error getting instance list: error getting instances in verify", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopVerifyTimeout() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.
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
		Times(3).
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		On("StartSystemContext", ctx, mock.Anything).
		Return(nil, nil)

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         0.0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{
			OperatorOptions: []operator.Option[operator.SAPSystemStop]{
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemInterval(0 * time.Second)),
			},
		},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues(
		"rollback to started failed: error waiting until system is in desired state\n"+
			"verify system stopped failed: "+
			"error waiting until system is in desired state", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopRollbackStoppingError() {
	ctx := context.Background()

	green := sapcontrol.STATECOLORSAPControlGREEN

	suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
		Return(
			&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &green,
						Features:   "ABAP|GATEWAY|ICMAN|IGS",
					},
				},
			}, nil,
		).
		Once().
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, errors.New("error stopping")).
		On(
			"StartSystemContext",
			ctx,
			mock.MatchedBy(func(req *sapcontrol.StartSystem) bool {
				return *req.Options == sapcontrol.StartStopOptionSAPControlABAPINSTANCES
			}),
		).
		Return(nil, errors.New("error starting"))

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
			"instance_type":   "abap",
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{
			OperatorOptions: []operator.Option[operator.SAPSystemStop]{
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
			},
		},
	)

	report := sapSystemStopOperator.Run(ctx)

	suite.Nil(report.Success)
	suite.Equal(operator.ROLLBACK, report.Error.ErrorPhase)
	suite.EqualValues("error starting system: error starting\nerror stopping system: error stopping", report.Error.Message)
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopSuccess() {
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
						Dispstatus: &green,
						Features:   tt.features,
					},
				},
			}, nil).
			Once()

		stopSystem := suite.mockSapcontrol.
			On(
				"StopSystemContext",
				ctx,
				mock.MatchedBy(func(req *sapcontrol.StopSystem) bool {
					return *req.Options == tt.options
				}),
			).
			Return(nil, nil).
			NotBefore(planGetInstances)

		suite.mockSapcontrol.
			On("GetSystemInstanceListContext", mock.Anything, mock.Anything).
			Return(&sapcontrol.GetSystemInstanceListResponse{
				Instances: []*sapcontrol.SAPInstance{
					{
						Dispstatus: &gray,
						Features:   tt.features,
					},
				},
			}, nil).
			Once().
			NotBefore(stopSystem)

		sapSystemStopOperator := operator.NewSAPSystemStop(
			operator.Arguments{
				"instance_number": "00",
				"instance_type":   tt.instanceType,
			},
			"test-op",
			operator.Options[operator.SAPSystemStop]{
				OperatorOptions: []operator.Option[operator.SAPSystemStop]{
					operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
				},
			},
		)

		report := sapSystemStopOperator.Run(ctx)

		expectedDiff := map[string]any{
			"before": `{"stopped":false}`,
			"after":  `{"stopped":true}`,
		}

		suite.Nil(report.Error)
		suite.Equal(operator.VERIFY, report.Success.LastPhase)
		suite.EqualValues(expectedDiff, report.Success.Diff)
	}
}

func (suite *SAPSystemStopOperatorTestSuite) TestSAPSystemStopSuccessMultipleQueries() {
	ctx := context.Background()

	gray := sapcontrol.STATECOLORSAPControlGRAY
	green := sapcontrol.STATECOLORSAPControlGREEN

	planGetInstances := suite.mockSapcontrol.
		On("GetSystemInstanceListContext", ctx, mock.Anything).
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

	stopSystem := suite.mockSapcontrol.
		On("StopSystemContext", ctx, mock.Anything).
		Return(nil, nil).
		NotBefore(planGetInstances)

	suite.mockSapcontrol.
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
		Times(3).
		NotBefore(stopSystem).
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
		Once()

	sapSystemStopOperator := operator.NewSAPSystemStop(
		operator.Arguments{
			"instance_number": "00",
			"timeout":         5.0,
		},
		"test-op",
		operator.Options[operator.SAPSystemStop]{
			OperatorOptions: []operator.Option[operator.SAPSystemStop]{
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemSapcontrol(suite.mockSapcontrol)),
				operator.Option[operator.SAPSystemStop](operator.WithCustomStopSystemInterval(0 * time.Second)),
			},
		},
	)

	report := sapSystemStopOperator.Run(ctx)

	expectedDiff := map[string]any{
		"before": `{"stopped":false}`,
		"after":  `{"stopped":true}`,
	}

	suite.Nil(report.Error)
	suite.Equal(operator.VERIFY, report.Success.LastPhase)
	suite.EqualValues(expectedDiff, report.Success.Diff)
}
