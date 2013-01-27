package commons

import (
	"errors"
	"fmt"
)

type RuntimeError interface {
	Error() string
	Code() int
}

type ApplicationError struct {
	code    int
	message string
}

func (err *ApplicationError) Code() int {
	return err.code
}

func (err *ApplicationError) Error() string {
	return err.message
}

func NewRuntimeError(code int, message string) RuntimeError {
	return &ApplicationError{code: code, message: message}
}

func GetCode(e error, default_code int) int {
	a, ok := e.(RuntimeError)
	if !ok {
		return default_code
	}
	return a.Code()
}

func IsTimeout(e error) bool {
	a, ok := e.(RuntimeError)
	if ok {
		return a.Code() == TimeoutErr.Code()
	}
	return ok
}

func NewPanicError(s string, any interface{}) error {
	return errors.New(fmt.Sprintf("%s%v", s, any))
}

func NewTwinceError(first, second error) error {
	msg := fmt.Sprintf("return two error, first is {%s}, second is {%s}",
		first.Error(), second.Error())
	return errors.New(msg)
}

func NewError(v interface{}) error {
	err, _ := v.(error)
	if nil != err {
		return err
	}
	return fmt.Errorf("%v", v)
}

const (
	timeout_message = "time out"
)

var (
	NotImplemented = NewRuntimeError(501, "not implemented")
	TimeoutErr     = NewRuntimeError(504, timeout_message)
	DieError       = NewRuntimeError(500, "die.")
	IdNotExists    = NewRuntimeError(400, "'id' is required.")
	BodyNotExists  = NewRuntimeError(400, "'body' is required.")
)
