package plugin

import (
	"errors"
	"fmt"
)

type ErrorKind string

const (
	ErrorKindValidation ErrorKind = "validation"
	ErrorKindHTTP       ErrorKind = "http"
	ErrorKindRuntime    ErrorKind = "runtime"
)

var (
	ErrInvalidManifest = errors.New("invalid plugin manifest")
	ErrUnexpectedReply = errors.New("unexpected plugin registry response")
)

type Error struct {
	Kind ErrorKind
	Op   string
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Op, e.Msg)
	}
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Msg, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func validationError(msg string, err error) error {
	return &Error{Kind: ErrorKindValidation, Op: "validate manifest", Msg: msg, Err: err}
}

func httpError(op string, msg string, err error) error {
	return &Error{Kind: ErrorKindHTTP, Op: op, Msg: msg, Err: err}
}
