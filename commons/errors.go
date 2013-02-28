package commons

import (
	"bytes"
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

type MutiErrors struct {
	msg  string
	errs []error
}

func NewMutiErrors(msg string, errs []error) *MutiErrors {
	var buffer bytes.Buffer
	buffer.WriteString(msg)
	for _, e := range errs {
		buffer.WriteString("\n ")
		buffer.WriteString(e.Error())
	}
	return &MutiErrors{msg: buffer.String(), errs: errs}
}

func (self *MutiErrors) Error() string {
	return self.msg
}
func (self *MutiErrors) Errors() []error {
	return self.errs
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
	ContinueCode                     = 100 // 'Continue',
	SwitchingProtocolsCode           = 101
	ProcessingCode                   = 102
	OKCode                           = 200
	CreatedCode                      = 201
	AcceptedCode                     = 202
	NonAuthoritativeInformationCode  = 203
	NoContentCode                    = 204
	ResetContentCode                 = 205
	PartialContentCode               = 206
	MultiStatusCode                  = 207
	IMUsedCode                       = 226
	MultipleChoicesCode              = 300
	MovedPermanentlyCode             = 301
	MovedTemporarilyCode             = 302
	SeeOtherCode                     = 303
	NotModifiedCode                  = 304
	UseProxyCode                     = 305
	ReservedCode                     = 306
	TemporaryRedirectCode            = 307
	BadRequestCode                   = 400
	UnauthorizedCode                 = 401
	PaymentRequiredCode              = 402
	ForbiddenCode                    = 403
	NotFoundCode                     = 404
	MethodNotAllowedCode             = 405
	NotAcceptableCode                = 406
	ProxyAuthenticationRequiredCode  = 407
	RequestTimeoutCode               = 408
	ConflictCode                     = 409
	GoneCode                         = 410
	LengthRequiredCode               = 411
	PreconditionFailedCode           = 412
	RequestEntityTooLargeCode        = 413
	RequestURITooLongCode            = 414
	UnsupportedMediaTypeCode         = 415
	RequestedRangeNotSatisfiableCode = 416
	ExpectationFailedCode            = 417
	Im_a_TeapotCode                  = 418
	UnprocessableEntityCode          = 422
	LockedCode                       = 423
	FailedDependencyCode             = 424
	UpgradeRequiredCode              = 426
	InternalErrorCode                = 500
	NotImplementedCode               = 501
	BadGatewayCode                   = 502
	ServiceUnavailableCode           = 503
	GatewayTimeoutCode               = 504
	VersionNotSupportedCode          = 505
	VariantAlsoNegotiatesCode        = 506
	InsufficientStorageCode          = 507
	NotExtendedCode                  = 510
	NetworkErrorCode                 = 560

	ContinueError      = NewRuntimeError(ContinueCode, "continue")
	NotImplemented     = NewRuntimeError(NotImplementedCode, "not implemented")
	TimeoutErr         = NewRuntimeError(GatewayTimeoutCode, timeout_message)
	DieError           = NewRuntimeError(InternalErrorCode, "die.")
	IdNotExists        = NewRuntimeError(BadRequestCode, "'id' is required.")
	BodyNotExists      = NewRuntimeError(BadRequestCode, "'body' is required.")
	BodyIsEmpty        = NewRuntimeError(BadRequestCode, "'body' is empty.")
	ServiceUnavailable = NewRuntimeError(ServiceUnavailableCode, "service temporary unavailable, try again later")
	ValueIsNil         = NewRuntimeError(InternalErrorCode, "value is nil.")
)
