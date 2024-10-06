package errs

import (
	"github.com/cockroachdb/errors"
)

// set depth to 10 to skip runtime stacks and current file.
const depth = 10

// Common Application Errors
var (
	// NotFound is returned when a resource is not found
	NotFound = errors.NewWithDepth(depth, "not found")

	// InternalError is returned when internal logic got error
	InternalError = errors.NewWithDepth(depth, "internal error")

	// SomethingWentWrong is returned when got some bug or unexpected case
	//
	// inherited error from InternalError,
	// so errors.Is(err, InternalError) == true
	SomethingWentWrong = errors.WrapWithDepth(depth, InternalError, "something went wrong")

	// Skippable is returned when got an error but it can be skipped or ignored and continue
	Skippable = errors.NewWithDepth(depth, "skippable")

	// Retryable is returned when got an error but it can be retried
	Retryable = errors.NewWithDepth(depth, "retryable")

	// Unsupported is returned when a feature or result is not supported
	Unsupported = errors.NewWithDepth(depth, "unsupported")

	// NotSupported is returned when a feature or result is not supported
	// alias of Unsupported
	NotSupported = Unsupported

	// Unauthorized is returned when a request is unauthorized
	Unauthorized = errors.NewWithDepth(depth, "unauthorized")

	// Timeout is returned when a connection to a resource timed out
	Timeout = errors.NewWithDepth(depth, "timeout")

	// BadRequest is returned when a request is invalid
	BadRequest = errors.NewWithDepth(depth, "bad request")

	// InvalidArgument is returned when an argument is invalid
	//
	// inherited error from BadRequest,
	// so errors.Is(err, BadRequest) == true
	InvalidArgument = errors.WrapWithDepth(depth, BadRequest, "invalid argument")

	// ArgumentRequired is returned when an argument is required
	//
	// inherited error from BadRequest,
	// so errors.Is(err, BadRequest) == true
	ArgumentRequired = errors.WrapWithDepth(depth, BadRequest, "argument required")

	// Duplicate is returned when a resource already exists
	Duplicate = errors.NewWithDepth(depth, "duplicate")

	// Unimplemented is returned when a feature or method is not implemented
	//
	// inherited error from Unsupported,
	// so errors.Is(err, Unsupported) == true
	Unimplemented = errors.WrapWithDepth(depth, Unsupported, "unimplemented")
)

// Business Logic errors
var (
	// Overflow is returned when an overflow error occurs
	//
	// inherited error from InternalError,
	// so errors.Is(err, InternalError) == true
	Overflow = errors.WrapWithDepth(depth, InternalError, "overflow")

	// OverflowUint64 is returned when an uint64 overflow error occurs
	//
	// inherited error from Overflow,
	// so errors.Is(err, Overflow) == true
	OverflowUint32 = errors.WrapWithDepth(depth, Overflow, "overflow uint32")

	// OverflowUint64 is returned when an uint64 overflow error occurs
	//
	// inherited error from Overflow,
	// so errors.Is(err, Overflow) == true
	OverflowUint64 = errors.WrapWithDepth(depth, Overflow, "overflow uint64")

	// OverflowUint128 is returned when an uint128 overflow error occurs
	//
	// inherited error from Overflow,
	// so errors.Is(err, Overflow) == true
	OverflowUint128 = errors.WrapWithDepth(depth, Overflow, "overflow uint128")

	// InvalidState is returned when a state is invalid
	InvalidState = errors.NewWithDepth(depth, "invalid state")

	// ConflictSetting is returned when an indexer setting is conflicted
	ConflictSetting = errors.NewWithDepth(depth, "conflict setting")

	// Closed is returned when a resource is closed
	Closed = errors.NewWithDepth(depth, "closed")
)
