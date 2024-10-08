package operator

import (
	"fmt"
)

type ExecutionError struct {
	ErrorPhase OPERATION_PHASES
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
	LastPhase OPERATION_PHASES
}

type ExecutionReport struct {
	OperationID string
	Success     *ExecutionSuccess
	Error       *ExecutionError
}

func executionReportWithError(error error, phase OPERATION_PHASES, operationID string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: operationID,
		Error: &ExecutionError{
			Message:    error.Error(),
			ErrorPhase: phase,
		},
	}
}

func executionReportWithSuccess(diff map[string]any, phase OPERATION_PHASES, operationID string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: operationID,
		Success: &ExecutionSuccess{
			Diff:      diff,
			LastPhase: phase,
		},
	}
}
