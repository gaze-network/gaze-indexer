package requestcontext

// requestcontextError implements error interface
var _ error = requestcontextError{}

type requestcontextError struct {
	err     error
	status  int
	message string
}

func (r requestcontextError) Error() string {
	if r.err != nil {
		return r.err.Error()
	}
	return r.message
}

func (r requestcontextError) Unwrap() error {
	return r.err
}
