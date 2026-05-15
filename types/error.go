package types

import "fmt"

type AppError struct {
	Code Code
	Msg  string
	Err  error
}

func NewError(code Code, msg string) *AppError {
	if msg == "" {
		msg = Message(code)
	}
	return &AppError{Code: code, Msg: msg}
}

func WrapError(code Code, msg string, err error) *AppError {
	if msg == "" {
		msg = Message(code)
	}
	return &AppError{Code: code, Msg: msg, Err: err}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Msg
	}
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
