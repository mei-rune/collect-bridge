package commons

import (
	"encoding/json"
	"fmt"
	"time"
)

type DriverManager map[string]Driver

func NewDriverManager() *DriverManager {
	drv := make(DriverManager)
	return &drv
}

func (self *DriverManager) Stop(name string) {
	drv, ok := (*self)[name]
	if !ok {
		return
	}
	startable, ok := drv.(Startable)
	if ok && nil != startable {
		startable.Stop()
	}
}

func (self *DriverManager) Start(name string) error {
	drv, ok := (*self)[name]
	if !ok {
		return NotFound(name)
	}

	startable, ok := drv.(Startable)
	if ok && nil != startable {
		err := startable.Start()
		if nil != err {
			return err
		}
	}
	return nil
}

func (self *DriverManager) Reset(name string) error {
	drv, ok := (*self)[name]
	if !ok {
		return NotFound(name)
	}

	startable, ok := drv.(Startable)
	if ok && nil != startable {
		startable.Stop()
		err := startable.Start()
		if nil != err {
			return err
		}
	}
	return nil
}

func (self *DriverManager) Register(name string, driver Driver) {
	_, ok := (*self)[name]
	if ok {
		panic(fmt.Errorf("'%s' always registred.", name))
	}
	(*self)[name] = driver
}

func (self *DriverManager) Unregister(name string) {
	delete(*self, name)
}

func (self *DriverManager) Connect(name string) (Driver, bool) {
	driver, ok := (*self)[name]
	return driver, ok
}

func (self *DriverManager) Names() []string {
	names := make([]string, 0, 10)
	for k, _ := range *self {
		names = append(names, k)
	}
	return names
}

var (
	METRIC_DRVS          = map[string]func(ctx map[string]interface{}) (Driver, error){}
	NotImplementedResult = ReturnError(NotImplementedCode, "not implemented")
)

type Startable interface {
	Start() error
	Stop()
}

type Result interface {
	Return(value interface{}) Result
	ErrorCode() int
	ErrorMessage() string
	HasError() bool
	Error() RuntimeError
	Warnings() interface{}
	Value() Any
	InterfaceValue() interface{}
	Effected() int64
	LastInsertId() interface{}
	HasOptions() bool
	Options() Map
	RawOptions() map[string]interface{}
	CreatedAt() time.Time
	ToJson() string
}

type Any interface {
	AsInterface() interface{}

	AsBool() (bool, error)

	AsInt() (int, error)

	AsInt32() (int32, error)

	AsInt64() (int64, error)

	AsUint() (uint, error)

	AsUint32() (uint32, error)

	AsUint64() (uint64, error)

	AsString() (string, error)

	AsObject() (map[string]interface{}, error)

	AsArray() ([]interface{}, error)

	AsObjects() ([]map[string]interface{}, error)
}

type Map interface {
	Set(key string, value interface{})

	Contains(key string) bool

	GetWithDefault(key string, defaultValue interface{}) interface{}

	GetBoolWithDefault(key string, defaultValue bool) bool

	GetIntWithDefault(key string, defaultValue int) int

	GetInt32WithDefault(key string, defaultValue int32) int32

	GetInt64WithDefault(key string, defaultValue int64) int64

	GetUintWithDefault(key string, defaultValue uint) uint

	GetUint32WithDefault(key string, defaultValue uint32) uint32

	GetUint64WithDefault(key string, defaultValue uint64) uint64

	GetStringWithDefault(key, defaultValue string) string

	GetArrayWithDefault(key string, defaultValue []interface{}) []interface{}

	GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{}

	GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{}

	Get(key string) (interface{}, error)

	GetBool(key string) (bool, error)

	GetInt(key string) (int, error)

	GetInt32(key string) (int32, error)

	GetInt64(key string) (int64, error)

	GetUint(key string) (uint, error)

	GetUint32(key string) (uint32, error)

	GetUint64(key string) (uint64, error)

	GetString(key string) (string, error)

	GetArray(key string) ([]interface{}, error)

	GetObject(key string) (map[string]interface{}, error)

	GetObjects(key string) ([]map[string]interface{}, error)
}

type Driver interface {
	Get(map[string]string) Result
	Put(map[string]string) Result
	Create(map[string]string) Result
	Delete(map[string]string) Result
}

type DefaultDrv struct {
	GetValue, PutValue, CreateValue, DeleteValue interface{}
	GetErr, PutErr, CreateErr, DeleteErr         string
	GetCode, PutCode, CreateCode, DeleteCode     int
}

func (self *DefaultDrv) Get(params map[string]string) Result {
	return Return(self.GetValue).SetError(self.GetCode, self.GetErr)
}

func (self *DefaultDrv) Put(params map[string]string) Result {
	return Return(self.PutValue).SetError(self.PutCode, self.PutErr)
}

func (self *DefaultDrv) Create(params map[string]string) Result {
	return Return(self.CreateValue).SetError(self.CreateCode, self.CreateErr)
}

func (self *DefaultDrv) Delete(params map[string]string) Result {
	return Return(self.DeleteValue).SetError(self.DeleteCode, self.DeleteErr)
}

type SimpleResult struct {
	Verr          *ApplicationError      `json:"error,omitempty"`
	Vwarnings     interface{}            `json:"warnings,omitempty"`
	Vvalue        interface{}            `json:"value,omitempty"`
	Veffected     int64                  `json:"effected,omitempty"`
	VlastInsertId interface{}            `json:"lastInsertId,omitempty"`
	Voptions      map[string]interface{} `json:"options,omitempty"`
	Vcreated_at   time.Time              `json:"created_at,omitempty"`

	value AnyValue
}

func Return(value interface{}) *SimpleResult {
	return &SimpleResult{Vvalue: value, Vcreated_at: time.Now(), Veffected: -1, VlastInsertId: -1}
}

func ReturnError(code int, msg string) *SimpleResult {
	return Return(nil).SetError(code, msg)
}

func (self *SimpleResult) SetValue(value interface{}) *SimpleResult {
	self.Vvalue = value
	return self
}

func (self *SimpleResult) Return(value interface{}) Result {
	self.Vvalue = value
	return self
}

func (self *SimpleResult) SetOptions(options map[string]interface{}) *SimpleResult {
	self.Voptions = options
	return self
}

func (self *SimpleResult) SetOption(key string, value interface{}) *SimpleResult {
	if nil == self.Voptions {
		self.Voptions = make(map[string]interface{})
	}
	self.Voptions[key] = value
	return self
}

func (self *SimpleResult) SetError(code int, msg string) *SimpleResult {
	if 0 == code && 0 == len(msg) {
		return self
	}

	if nil == self.Verr {
		self.Verr = &ApplicationError{Vcode: code, Vmessage: msg}
	} else {
		self.Verr.Vcode = code
		self.Verr.Vmessage = msg
	}
	return self
}

func (self *SimpleResult) SetWarnings(value interface{}) *SimpleResult {
	self.Vwarnings = value
	return self
}

func (self *SimpleResult) SetEffected(effected int64) *SimpleResult {
	self.Veffected = effected
	return self
}

func (self *SimpleResult) SetLastInsertId(id interface{}) *SimpleResult {
	self.VlastInsertId = id
	return self
}

func (self *SimpleResult) ErrorCode() int {
	if nil != self.Verr {
		return self.Verr.Vcode
	}
	return -1
}

func (self *SimpleResult) ErrorMessage() string {
	if nil != self.Verr {
		return self.Verr.Vmessage
	}
	return ""
}

func (self *SimpleResult) HasError() bool {
	return nil != self.Verr && (0 != self.Verr.Vcode || 0 != len(self.Verr.Vmessage))
}

func (self *SimpleResult) Error() RuntimeError {
	if nil == self.Verr {
		return nil
	}
	return self.Verr
}

func (self *SimpleResult) Warnings() interface{} {
	return self.Vwarnings
}

func (self *SimpleResult) Value() Any {
	self.value.Value = self.Vvalue
	return &self.value
}

func (self *SimpleResult) InterfaceValue() interface{} {
	return self.Vvalue
}

func (self *SimpleResult) Effected() int64 {
	return self.Veffected
}

func (self *SimpleResult) LastInsertId() interface{} {
	return self.VlastInsertId
}

func (self *SimpleResult) HasOptions() bool {
	return nil != self.Voptions && 0 != len(self.Voptions)
}

func (self *SimpleResult) Options() Map {
	if nil == self.Voptions {
		self.Voptions = make(map[string]interface{})
	}
	return InterfaceMap(self.Voptions)
}

func (self *SimpleResult) RawOptions() map[string]interface{} {
	if nil == self.Voptions {
		self.Voptions = make(map[string]interface{})
	}
	return self.Voptions
}

func (self *SimpleResult) CreatedAt() time.Time {
	return self.Vcreated_at
}

func (self *SimpleResult) ToJson() string {
	bs, e := json.Marshal(self)
	if nil != e {
		panic(e.Error())
	}
	return string(bs)
}

type AnyValue struct {
	Value interface{}
}

func (self *AnyValue) IsNil() bool {
	return nil == self.Value
}

func (self *AnyValue) AsInterface() interface{} {
	return self.Value
}

func (self *AnyValue) AsBool() (bool, error) {
	return AsBool(self.Value)
}

func (self *AnyValue) AsInt() (int, error) {
	return AsInt(self.Value)
}

func (self *AnyValue) AsInt32() (int32, error) {
	return AsInt32(self.Value)
}

func (self *AnyValue) AsInt64() (int64, error) {
	return AsInt64(self.Value)
}

func (self *AnyValue) AsUint() (uint, error) {
	return AsUint(self.Value)
}

func (self *AnyValue) AsUint32() (uint32, error) {
	return AsUint32(self.Value)
}

func (self *AnyValue) AsUint64() (uint64, error) {
	return AsUint64(self.Value)
}

func (self *AnyValue) AsString() (string, error) {
	return AsString(self.Value)
}

func (self *AnyValue) AsArray() ([]interface{}, error) {
	if m, ok := self.Value.([]interface{}); ok {
		return m, nil
	}
	return nil, IsNotArray
}

func (self *AnyValue) AsObject() (map[string]interface{}, error) {
	if m, ok := self.Value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, IsNotMap
}

func (self *AnyValue) AsObjects() ([]map[string]interface{}, error) {
	if o, ok := self.Value.([]map[string]interface{}); ok {
		return o, nil
	}

	a, ok := self.Value.([]interface{})
	if !ok {
		return nil, IsNotArray
	}

	res := make([]map[string]interface{}, 0, len(a))
	for _, v := range a {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, IsNotMap
		}
		res = append(res, m)
	}
	return res, nil
}

func ReturnWithInternalError(message string) Result {
	return ReturnError(InternalErrorCode, message)
}

func ReturnWithBadRequest(message string) Result {
	return ReturnError(BadRequestCode, message)
}

func ReturnWithNotAcceptable(message string) Result {
	return ReturnError(NotAcceptableCode, message)
}

func ReturnWithIsRequired(name string) Result {
	return ReturnError(BadRequestCode, "'"+name+"' is required.")
}

func ReturnWithNotFound(id string) Result {
	return ReturnError(NotFoundCode, "'"+id+"' is not found.")
}

func ReturnWithRecordNotFound(id string) Result {
	return ReturnWithNotFound(id)
}

func ReturnWithRecordAlreadyExists(id string) Result {
	return ReturnError(NotAcceptableCode, "'"+id+"' is already exists.")
}

func ReturnWithNotImplemented() Result {
	return ReturnError(InternalErrorCode, "not implemented")
}
