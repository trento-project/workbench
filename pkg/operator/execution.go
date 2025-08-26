package operator

import (
	"fmt"
)

type ExecutionError struct {
	ErrorPhase OperatorPhases
	Message    string
}

func (e ExecutionError) Error() string {
	return fmt.Sprintf(
		"error during operator exeuction in phase: %s, reason: %s",
		e.ErrorPhase,
		e.Message,
	)
}

type ExecutionSuccess struct {
	Diff      map[string]any
	LastPhase OperatorPhases
}

type ExecutionReport struct {
	OperationID string
	Success     *ExecutionSuccess
	Error       *ExecutionError
}

func executionReportWithError(error error, phase OperatorPhases, operationID string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: operationID,
		Error: &ExecutionError{
			Message:    error.Error(),
			ErrorPhase: phase,
		},
	}
}

func executionReportWithSuccess(diff map[string]any, phase OperatorPhases, operationID string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: operationID,
		Success: &ExecutionSuccess{
			Diff:      diff,
			LastPhase: phase,
		},
	}
}
