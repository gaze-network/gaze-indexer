package errs

// ErrorKind identifies a kind of internal error.
// fully support for errors.Is and errors.As.
type ErrorKind string

const (
	// NotFound is returned when a requested item is not found.
	NotFound = ErrorKind("Not Found")
)

// Error satisfies the error interface and prints human-readable errors.
func (e ErrorKind) Error() string {
	return string(e)
}
