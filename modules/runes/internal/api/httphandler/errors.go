package httphandler

import "github.com/cockroachdb/errors"

// TODO: rewrite entire error handling logic
// may create new custom error type in errs package for using to return errors to client

type ValidationError struct {
	errs []error
}

func (v ValidationError) Error() string {
	return errors.Join(v.errs...).Error()
}

func NewValidationError(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	return &ValidationError{errs: errs}
}

type UserError struct {
	err error
}

func (u UserError) Error() string {
	return u.err.Error()
}

func NewUserError(err error) error {
	if err == nil {
		return nil
	}
	return &UserError{err: err}
}
