package errs

import (
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/withstack"
)

// PublicError is an error that, when caught by error handler, should return a user-friendly error response to the user. Responses vary between each protocol (http, grpc, etc.).
type PublicError struct {
	err     error
	message string
}

func (p PublicError) Error() string {
	return p.err.Error()
}

func (p PublicError) Message() string {
	return p.message
}

func (p PublicError) Unwrap() error {
	return p.err
}

func NewPublicError(message string) error {
	return withstack.WithStackDepth(&PublicError{err: errors.New(message), message: message}, 1)
}

func WithPublicMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return withstack.WithStackDepth(&PublicError{err: err, message: message}, 1)
}
