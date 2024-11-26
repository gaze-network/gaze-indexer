package errs

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/withstack"
)

// PublicError is an error that, when caught by error handler, should return a user-friendly error response to the user. Responses vary between each protocol (http, grpc, etc.).
type PublicError struct {
	err     error
	message string
	code    string // code is optional, it can be used to identify the error type
}

func (p PublicError) Error() string {
	return p.err.Error()
}

func (p PublicError) Message() string {
	return p.message
}

func (p PublicError) Code() string {
	return p.code
}

func (p PublicError) Unwrap() error {
	return p.err
}

func NewPublicError(message string) error {
	return withstack.WithStackDepth(&PublicError{err: errors.New(message), message: message}, 1)
}

func NewPublicErrorWithCode(message string, code string) error {
	return withstack.WithStackDepth(&PublicError{err: errors.New(message), message: message, code: code}, 1)
}

func WithPublicMessage(err error, prefix string) error {
	if err == nil {
		return nil
	}
	var message string
	if prefix != "" {
		message = fmt.Sprintf("%s: %s", prefix, err.Error())
	} else {
		message = err.Error()
	}
	return withstack.WithStackDepth(&PublicError{err: err, message: message}, 1)
}

func WithPublicMessageCode(err error, prefix string, code string) error {
	if err == nil {
		return nil
	}
	var message string
	if prefix != "" {
		message = fmt.Sprintf("%s: %s", prefix, err.Error())
	} else {
		message = err.Error()
	}
	return withstack.WithStackDepth(&PublicError{err: err, message: message, code: code}, 1)
}
