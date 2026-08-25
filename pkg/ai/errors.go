package ai

import "errors"

const (
	ErrorCodeTimeout            = "AI_TIMEOUT"
	ErrorCodeUnavailable        = "AI_UNAVAILABLE"
	ErrorCodeInvalidResponse    = "AI_INVALID_RESPONSE"
	ErrorCodeUnexpectedResponse = "AI_UNEXPECTED_RESPONSE"
)

type RequestError struct {
	Code string
	err  error
}

func (e *RequestError) Error() string {
	return "AI prediction request failed"
}

func (e *RequestError) Unwrap() error {
	return e.err
}

func ErrorCode(err error) string {
	var requestErr *RequestError
	if errors.As(err, &requestErr) {
		return requestErr.Code
	}
	return ErrorCodeUnavailable
}
