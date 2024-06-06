package common

type HttpResponse[T any] struct {
	Error  *string `json:"error"`
	Result *T      `json:"result,omitempty"`
}
