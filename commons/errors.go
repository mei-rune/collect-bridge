package commons

import (
	"bytes"
	"errors"
	"fmt"
)

type RuntimeError interface {
	Code() int
	Error() string
}
type ApplicationError struct {
	Ecode    int    `json:"code,omitempty"`
	Emessage string `json:"message,omitempty"`
}

func (err *ApplicationError) Code() int {
	return err.Ecode
}

func (err *ApplicationError) Error() string {
	return err.Emessage
}

func NewApplicationError(code int, msg string) RuntimeError {
	return &ApplicationError{Ecode: code, Emessage: msg}
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

func IsTimeout(e error) bool {
	return timeout_message == e.Error()
}

// func IsNotFound(e error) bool {
// 	a, ok := e.(error)
// 	if ok {
// 		return a.Code() == 404
// 	}
// 	return ok
// }

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
	TypeErrorCode                    = 460
	NilValueCode                     = 461
	ParameterIsEmptyCode             = 462
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
	InterruptErrorCode               = 561
	IsRequiredCode                   = BadRequestCode
	TableIsNotExists                 = BadRequestCode + 80

	AlreadyStartedError = errors.New("already started.")
	ContinueError       = errors.New("continue")
	NotImplemented      = NewApplicationError(NotImplementedCode, "not implemented")
	TimeoutErr          = errors.New(timeout_message)
	DieError            = errors.New("die.")
	NotExists           = NewApplicationError(NotFoundCode, "not found.")
	IdNotExists         = IsRequired("id")
	BodyNotExists       = IsRequired("body")
	BodyIsEmpty         = errors.New("'body' is empty.")
	ServiceUnavailable  = errors.New("service temporary unavailable, try again later")
	ValueIsNil          = errors.New("value is nil.")
	NotIntValue         = errors.New("it is not a int.")
	InterruptError      = errors.New("interrupt error")

	IsNotMapOrArray = TypeError("it is not a map[string]interface{} or a []interface{}.")
	IsNotMap        = TypeError("it is not a map[string]interface{}.")
	IsNotArray      = TypeError("it is not a []interface{}.")
	IsNotBool       = TypeError("it is not a bool.")
	IsNotInt8       = TypeError("it is not a int8.")
	IsNotInt16      = TypeError("it is not a int16.")
	IsNotInt32      = TypeError("it is not a int32.")
	IsNotInt64      = TypeError("it is not a int64.")
	Int8OutRange    = TypeError("it is not a int8, out range")
	Int16OutRange   = TypeError("it is not a int16, out range")
	Int32OutRange   = TypeError("it is not a int32, out range")
	Int64OutRange   = TypeError("it is not a int64, out range")
	IsNotUint8      = TypeError("it is not a uint8.")
	IsNotUint16     = TypeError("it is not a uint16.")
	IsNotUint32     = TypeError("it is not a uint32.")
	IsNotUint64     = TypeError("it is not a uint64.")
	Uint8OutRange   = TypeError("it is not a uint8, out range")
	Uint16OutRange  = TypeError("it is not a uint16, out range")
	Uint32OutRange  = TypeError("it is not a uint32, out range")
	Uint64OutRange  = TypeError("it is not a uint64, out range")
	IsNotFloat32    = TypeError("it is not a float32.")
	IsNotFloat64    = TypeError("it is not a float64.")
	IsNotString     = TypeError("it is not a string.")

	ErrNotString     = TypeError("it is not a string")
	ErrNotTimeString = TypeError("it is not a time string")

	ParameterIsNil   = errors.New("parameter is nil.")
	ParameterIsEmpty = errors.New("parameter is empty.")

	MultipleValuesError = errors.New("Multiple values meet the conditions")
)

func TypeError(message string) RuntimeError {
	return NewApplicationError(TypeErrorCode, message)
}

func IsRequired(name string) error {
	return NewApplicationError(NotFoundCode, "'"+name+"' is required.")
}

func NotFound(msg string) error {
	return NewApplicationError(NotFoundCode, msg)
}

func RecordNotFound(id string) error {
	return NewApplicationError(NotFoundCode, "'"+id+"' is not found.")
}

func RecordNotFoundWithType(t, id string) error {
	if 0 == len(t) {
		return RecordNotFound(id)
	}
	return NewApplicationError(NotFoundCode, t+" with id was '"+id+"' is not found.")
}

func RecordAlreadyExists(id string) error {
	return errors.New("'" + id + "' is already exists.")
}
