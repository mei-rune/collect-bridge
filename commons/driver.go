package commons

import (
	"commons/as"
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
	METRIC_DRVS          = map[string]func(ctx map[string]interface{}) (Driver, RuntimeError){}
	NotImplementedResult = ReturnError(NotImplementedCode, "not implemented")
)

type Startable interface {
	Start() error
	Stop()
}

type Result interface {
	ErrorCode() int
	ErrorMessage() string
	HasError() bool
	Warnings() interface{}
	Value() Any
	InterfaceValue() interface{}
	Effected() int
	HasOptions() bool
	Options() Map
	CreatedAt() time.Time
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

	AsObjects() ([]map[string]interface{}, error)
}

type Map interface {
	TryGetBool(key string) (bool, error)

	TryGetInt(key string) (int, error)

	TryGetInt32(key string) (int32, error)

	TryGetInt64(key string) (int64, error)

	TryGetUint(key string) (uint, error)

	TryGetUint32(key string) (uint32, error)

	TryGetUint64(key string) (uint64, error)

	TryGetString(key string) (string, error)

	TryGetObject(key string) (map[string]interface{}, error)

	TryGetObjects(key string) ([]map[string]interface{}, error)

	ToMap() map[string]interface{}
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

type simpleResult struct {
	err        *applicationError      `json:"error"`
	warnings   interface{}            `json:"warnings"`
	value      AnyValue               `json:"value"`
	effected   int                    `json:"effected"`
	options    map[string]interface{} `json:"options"`
	created_at time.Time              `json:"created_at"`
}

func Return(value interface{}) *simpleResult {
	return &simpleResult{value: AnyValue{Value: value}, created_at: time.Now()}
}

func ReturnError(code int, msg string) *simpleResult {
	return &simpleResult{err: &applicationError{code: code, message: msg}, created_at: time.Now()}
}

func ReturnWithError(e RuntimeError) *simpleResult {
	return ReturnError(e.Code(), e.Error())
}

func (self *simpleResult) SetValue(value interface{}) *simpleResult {
	self.value.Value = value
	return self
}

func (self *simpleResult) SetOption(key string, value interface{}) *simpleResult {
	if nil == self.options {
		self.options = make(map[string]interface{})
	}
	self.options[key] = value
	return self
}

func (self *simpleResult) SetErrorMessage(msg string) *simpleResult {
	if nil == self.err {
		self.err = &applicationError{code: 500, message: msg}
	} else {
		self.err.message = msg
	}
	return self
}

func (self *simpleResult) SetErrorCode(code int) *simpleResult {
	if nil == self.err {
		self.err = &applicationError{code: code}
	} else {
		self.err.code = code
	}
	return self
}

func (self *simpleResult) SetError(code int, msg string) *simpleResult {
	if nil == self.err {
		self.err = &applicationError{code: code, message: msg}
	} else {
		self.err.code = code
		self.err.message = msg
	}
	return self
}

func (self *simpleResult) SetWarnings(value interface{}) *simpleResult {
	self.warnings = value
	return self
}

func (self *simpleResult) SetEffected(effected int) *simpleResult {
	self.effected = effected
	return self
}

func (self *simpleResult) ErrorCode() int {
	if nil != self.err {
		return self.err.code
	}
	return -1
}
func (self *simpleResult) ErrorMessage() string {
	if nil != self.err {
		return self.err.message
	}
	return ""
}
func (self *simpleResult) HasError() bool {
	return nil != self.err
}
func (self *simpleResult) Warnings() interface{} {
	return self.warnings
}
func (self *simpleResult) Value() Any {
	return &self.value
}
func (self *simpleResult) InterfaceValue() interface{} {
	return self.value.Value
}
func (self *simpleResult) Effected() int {
	return self.effected
}
func (self *simpleResult) HasOptions() bool {
	return nil != self.options && 0 != len(self.options)
}
func (self *simpleResult) Options() Map {
	if nil == self.options {
		self.options = make(map[string]interface{})
	}
	return StringMap(self.options)
}
func (self *simpleResult) CreatedAt() time.Time {
	return self.created_at
}

func (self *simpleResult) ToJson() string {
	bs, e := json.Marshal(self)
	if nil != e {
		panic(e.Error())
	}
	return string(bs)
}

type AnyValue struct {
	Value interface{}
}

func (self *AnyValue) AsInterface() interface{} {
	return self.Value
}

func (self *AnyValue) AsBool() (bool, error) {
	return as.AsBool(self.Value)
}

func (self *AnyValue) AsInt() (int, error) {
	return as.AsInt(self.Value)
}

func (self *AnyValue) AsInt32() (int32, error) {
	return as.AsInt32(self.Value)
}

func (self *AnyValue) AsInt64() (int64, error) {
	return as.AsInt64(self.Value)
}

func (self *AnyValue) AsUint() (uint, error) {
	return as.AsUint(self.Value)
}

func (self *AnyValue) AsUint32() (uint32, error) {
	return as.AsUint32(self.Value)
}

func (self *AnyValue) AsUint64() (uint64, error) {
	return as.AsUint64(self.Value)
}

func (self *AnyValue) AsString() (string, error) {
	return as.AsString(self.Value)
}

func (self *AnyValue) AsObject() (map[string]interface{}, error) {
	if m, ok := self.Value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, as.IsNotMap
}

func (self *AnyValue) AsObjects() ([]map[string]interface{}, error) {
	if o, ok := self.Value.([]map[string]interface{}); ok {
		return o, nil
	}

	a, ok := self.Value.([]interface{})
	if !ok {
		return nil, as.IsNotArray
	}

	res := make([]map[string]interface{}, 0, len(a))
	for _, v := range a {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, as.IsNotMap
		}
		res = append(res, m)
	}
	return res, nil
}

type StringMap map[string]interface{}

func (self StringMap) ToMap() map[string]interface{} {
	return map[string]interface{}(self)
}

func (self StringMap) TryGetBool(key string) (bool, error) {
	return TryGetBool(self, key)
}

func (self StringMap) TryGetInt(key string) (int, error) {
	return TryGetInt(self, key)
}

func (self StringMap) TryGetInt32(key string) (int32, error) {
	return TryGetInt32(self, key)
}

func (self StringMap) TryGetInt64(key string) (int64, error) {
	return TryGetInt64(self, key)
}

func (self StringMap) TryGetUint(key string) (uint, error) {
	return TryGetUint(self, key)
}

func (self StringMap) TryGetUint32(key string) (uint32, error) {
	return TryGetUint32(self, key)
}

func (self StringMap) TryGetUint64(key string) (uint64, error) {
	return TryGetUint64(self, key)
}

func (self StringMap) TryGetString(key string) (string, error) {
	return TryGetString(self, key)
}

func (self StringMap) TryGetObject(key string) (map[string]interface{}, error) {
	return TryGetObject(self, key)
}

func (self StringMap) TryGetObjects(key string) ([]map[string]interface{}, error) {
	return TryGetObjects(self, key)
}
