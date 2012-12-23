package commons

import (
	"fmt"
)

// const (
//	MGR_STATUS_STANDBY  = 0
//	MGR_STATUS_STARTING = 1
//	MGR_STATUS_RUNNING  = 2
//	MGR_STATUS_STOPPING = 3
// )

// type Error struct {
//	message string
// }

// func NewError(s string) error {
//	return &Error{s}
// }

// func (err *Error) Error() string {
//	return err.message
// }

// type StatusError struct {
//	status int32
//	Error
// }

// func (err *StatusError) Status() int {
//	return err.status
// }

// func NewStatusError(code int) *StatusError {
//	switch code {
//	case MGR_STATUS_STANDBY:
//		return &StatusError(code, "mgr.status is standby")
//	case MGR_STATUS_STARTING:
//		return &StatusError(code, "mgr.status is starting")
//	case MGR_STATUS_RUNNING:
//		return &StatusError(code, "mgr.status is running")
//	case MGR_STATUS_STOPPING:
//		return &StatusError(code, "mgr.status is stopping")
//	default:
//		return &StatusError(code, "mgr.status is unknown")
//	}
// }

func NewError(v interface{}) error {
	err, _ := v.(error)
	if nil != err {
		return err
	}
	return fmt.Errorf("%v", v)
}
