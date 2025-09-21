package agent

import "errors"

// Common errors
var (
	ErrNoHandler        = errors.New("no handler provided")
	ErrNoStreamHandler  = errors.New("no stream handler provided")
	ErrInvalidInput     = errors.New("invalid input")
	ErrExecutionFailed  = errors.New("execution failed")
	ErrCapabilityExists = errors.New("capability already exists")
	ErrCapabilityNotFound = errors.New("capability not found")
)