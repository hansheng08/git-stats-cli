// Package apperr classifies errors so cli can map them to exit codes.
// It imports only the standard library.
package apperr

import "errors"

type kind int

const (
	kindUsage kind = iota
	kindRuntime
)

type kindError struct {
	kind kind
	err  error
}

func (e *kindError) Error() string { return e.err.Error() }

func (e *kindError) Unwrap() error { return e.err }

// Usage marks err as a usage error (exit code 2). Usage(nil) returns nil.
func Usage(err error) error {
	if err == nil {
		return nil
	}
	return &kindError{kind: kindUsage, err: err}
}

// Runtime marks err as a runtime error (exit code 1). Runtime(nil) returns nil.
func Runtime(err error) error {
	if err == nil {
		return nil
	}
	return &kindError{kind: kindRuntime, err: err}
}

// IsUsage reports whether err or any error it wraps was marked by Usage.
func IsUsage(err error) bool {
	for err != nil {
		var ke *kindError
		if !errors.As(err, &ke) {
			return false
		}
		if ke.kind == kindUsage {
			return true
		}
		err = ke.err
	}
	return false
}
