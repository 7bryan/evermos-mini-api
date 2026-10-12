package apperrors

import "errors"

type Kind int

const (
	KindInternal Kind = iota
	KindValidation
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
	KindUpstream
)

type Error struct {
	Kind    Kind
	Message string
	Details []string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func Validation(message string, details ...string) *Error {
	return &Error{Kind: KindValidation, Message: message, Details: details}
}

func Unauthorized(message string) *Error { return &Error{Kind: KindUnauthorized, Message: message} }
func Forbidden(message string) *Error    { return &Error{Kind: KindForbidden, Message: message} }
func NotFound(message string) *Error     { return &Error{Kind: KindNotFound, Message: message} }
func Conflict(message string) *Error     { return &Error{Kind: KindConflict, Message: message} }

func Upstream(message string, cause error) *Error {
	return &Error{Kind: KindUpstream, Message: message, Err: cause}
}

func Internal(cause error) *Error {
	return &Error{Kind: KindInternal, Message: "internal server error", Err: cause}
}

func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
