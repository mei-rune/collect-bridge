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

func NotFound(id string) RuntimeError {
	return NewRuntimeError(404, "'"+id+"' is not found.")
}

func IsTimeout(e error) bool {
	a, ok := e.(RuntimeError)
	if ok {
		return a.Code() == TimeoutErr.Code()
	}
	return ok
}

func IsNotFound(e error) bool {
	a, ok := e.(RuntimeError)
	if ok {
		return a.Code() == 404
	}
	return ok
}

const (
	timeout_message = "time out"
)

var (
	NotFoundCode      = 404
	BadRequestCode    = 400
	InternalErrorCode = 500

	NotImplemented     = NewRuntimeError(501, "not implemented")
	TimeoutErr         = NewRuntimeError(504, timeout_message)
	DieError           = NewRuntimeError(500, "die.")
	IdNotExists        = NewRuntimeError(400, "'id' is required.")
	BodyNotExists      = NewRuntimeError(400, "'body' is required.")
	BodyIsEmpty        = NewRuntimeError(400, "'body' is empty.")
	ServiceUnavailable = NewRuntimeError(503, "service temporary unavailable, try again later")
)
