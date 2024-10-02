package operator

import "fmt"

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
	Diff map[string]string
}

type ExecutionReport struct {
	OperationID string
	Success     *ExecutionSuccess
	Error       *ExecutionError
}
