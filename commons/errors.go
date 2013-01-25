package commons

import (
	"errors"
	"fmt"
)

type TimeoutError struct {
	message string
}

func (err *TimeoutError) Error() string {
	return err.message
}

func IsTimeout(e error) bool {
	_, ok := e.(*TimeoutError)
	return ok
}

type PanicError struct {
	message string
	any     interface{}
}

func (err *PanicError) Error() string {
	return err.message
}

func NewPanicError(s string, any interface{}) error {
	return &PanicError{message: fmt.Sprintf("%s%v", s, any), any: any}
}

type TwinceError struct {
	message       string
	first, second error
}

func (err *TwinceError) Error() string {
	return err.message
}

func NewTwinceError(first, second error) error {
	msg := fmt.Sprintf("return two error, first is {%s}, second is {%s}",
		first.Error(), second.Error())
	return &TwinceError{message: msg, first: first, second: second}
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
	NotImplemented = fmt.Errorf("not implemented")
	TimeoutErr     = &TimeoutError{message: timeout_message}
	DieError       = errors.New("die.")
	IdNotFound     = errors.New("'id' is required.")
)
