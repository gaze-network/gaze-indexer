package errs

import (
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/withstack"
)

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

func NewPublicError(message string) error {
	return withstack.WithStackDepth(&PublicError{err: errors.New(message), message: message}, 1)
}

func WithPublicMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return withstack.WithStackDepth(&PublicError{err: err, message: message}, 1)
}
