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
		return RecordNotFound(name)
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
		return RecordNotFound(name)
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
	ToMap() map[string]interface{}
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
	CopyTo(copy map[string]interface{})

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

	GetFloatWithDefault(key string, defaultValue float64) float64

	GetStringWithDefault(key, defaultValue string) string

	GetArrayWithDefault(key string, defaultValue []interface{}) []interface{}

	GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{}

	GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{}

	Get(key string) (interface{}, error)

	GetBool(key string) (bool, error)

	GetInt(key string) (int, error)

	GetInt32(key string) (int32, error)

	GetInt64(key string) (int64, error)

	GetFloat(key string) (float64, error)

	GetUint(key string) (uint, error)

	GetUint32(key string) (uint32, error)

	GetUint64(key string) (uint64, error)

	GetString(key string) (string, error)

	GetArray(key string) ([]interface{}, error)

	GetObject(key string) (map[string]interface{}, error)

	GetObjects(key string) ([]map[string]interface{}, error)
}

type Driver interface {
	Get(params map[string]string) Result
	Put(params map[string]string, body interface{}) Result
	Create(params map[string]string, body interface{}) Result
	Delete(params map[string]string) Result
}

type DefaultDrv struct {
	GetValue, PutValue, CreateValue, DeleteValue interface{}
	GetErr, PutErr, CreateErr, DeleteErr         string
	GetCode, PutCode, CreateCode, DeleteCode     int
}

func (self *DefaultDrv) Get(params map[string]string) Result {
	return Return(self.GetValue).SetError(self.GetCode, self.GetErr)
}

func (self *DefaultDrv) Put(params map[string]string, body interface{}) Result {
	return Return(self.PutValue).SetError(self.PutCode, self.PutErr)
}

func (self *DefaultDrv) Create(params map[string]string, body interface{}) Result {
	return Return(self.CreateValue).SetError(self.CreateCode, self.CreateErr)
}

func (self *DefaultDrv) Delete(params map[string]string) Result {
	return Return(self.DeleteValue).SetError(self.DeleteCode, self.DeleteErr)
}

type SimpleResult struct {
	Eid             interface{}            `json:"request_id,omitempty"`
	Eerror          *ApplicationError      `json:"error,omitempty"`
	Ewarnings       interface{}            `json:"warnings,omitempty"`
	Evalue          interface{}            `json:"value,omitempty"`
	Eeffected       *int64                 `json:"effected,omitempty"`
	ElastInsertId   interface{}            `json:"lastInsertId,omitempty"`
	Eoptions        map[string]interface{} `json:"options,omitempty"`
	Ecreated_at     time.Time              `json:"created_at,omitempty"`
	Erepresentation string                 `json:"representation,omitempty"`

	value    AnyData
	effected int64
}

func Return(value interface{}) *SimpleResult {
	return &SimpleResult{Evalue: value, Ecreated_at: time.Now(), effected: -1, ElastInsertId: nil}
}

func ReturnError(code int, msg string) *SimpleResult {
	return Return(nil).SetError(code, msg)
}

func (self *SimpleResult) SetValue(value interface{}) *SimpleResult {
	self.Evalue = value
	return self
}

func (self *SimpleResult) Return(value interface{}) Result {
	self.Evalue = value
	return self
}

func (self *SimpleResult) SetOptions(options map[string]interface{}) *SimpleResult {
	self.Eoptions = options
	return self
}

func (self *SimpleResult) SetOption(key string, value interface{}) *SimpleResult {
	if nil == self.Eoptions {
		self.Eoptions = make(map[string]interface{})
	}
	self.Eoptions[key] = value
	return self
}

func (self *SimpleResult) SetError(code int, msg string) *SimpleResult {
	if 0 == code && 0 == len(msg) {
		return self
	}

	if nil == self.Eerror {
		self.Eerror = &ApplicationError{Ecode: code, Emessage: msg}
	} else {
		self.Eerror.Ecode = code
		self.Eerror.Emessage = msg
	}
	return self
}

func (self *SimpleResult) SetWarnings(value interface{}) *SimpleResult {
	self.Ewarnings = value
	return self
}

func (self *SimpleResult) SetEffected(effected int64) *SimpleResult {
	self.effected = effected
	self.Eeffected = &self.effected
	return self
}

func (self *SimpleResult) SetLastInsertId(id interface{}) *SimpleResult {
	self.ElastInsertId = id
	return self
}

func (self *SimpleResult) ErrorCode() int {
	if nil != self.Eerror {
		return self.Eerror.Ecode
	}
	return -1
}

func (self *SimpleResult) ErrorMessage() string {
	if nil != self.Eerror {
		return self.Eerror.Emessage
	}
	return ""
}

func (self *SimpleResult) HasError() bool {
	return nil != self.Eerror && (0 != self.Eerror.Ecode || 0 != len(self.Eerror.Emessage))
}

func (self *SimpleResult) Error() RuntimeError {
	if nil == self.Eerror {
		return nil
	}
	return self.Eerror
}

func (self *SimpleResult) Warnings() interface{} {
	return self.Ewarnings
}

func (self *SimpleResult) Value() Any {
	self.value.Value = self.Evalue
	return &self.value
}

func (self *SimpleResult) InterfaceValue() interface{} {
	return self.Evalue
}

func (self *SimpleResult) Effected() int64 {
	if nil != self.Eeffected {
		return *self.Eeffected
	}
	return -1
}

func (self *SimpleResult) LastInsertId() interface{} {
	if nil == self.ElastInsertId {
		return -1
	}
	return self.ElastInsertId
}

func (self *SimpleResult) HasOptions() bool {
	return nil != self.Eoptions && 0 != len(self.Eoptions)
}

func (self *SimpleResult) Options() Map {
	if nil == self.Eoptions {
		self.Eoptions = make(map[string]interface{})
	}
	return InterfaceMap(self.Eoptions)
}

func (self *SimpleResult) RawOptions() map[string]interface{} {
	if nil == self.Eoptions {
		self.Eoptions = make(map[string]interface{})
	}
	return self.Eoptions
}

func (self *SimpleResult) CreatedAt() time.Time {
	return self.Ecreated_at
}

func (self *SimpleResult) ToJson() string {
	bs, e := json.Marshal(self)
	if nil != e {
		panic(e.Error())
	}
	return string(bs)
}

func (self *SimpleResult) ToMap() map[string]interface{} {
	res := map[string]interface{}{}

	res["created_at"] = self.Ecreated_at
	if 0 != len(self.Erepresentation) {
		res["representation"] = self.Erepresentation
	}

	if nil != self.Eerror {
		res["error"] = map[string]interface{}{"code": self.Eerror.Ecode, "message": self.Eerror.Emessage}
	}
	if nil != self.Ewarnings {
		res["warnings"] = self.Ewarnings
	}
	if nil != self.Evalue {
		res["value"] = self.Evalue
	}
	if nil != self.Eeffected && -1 != *self.Eeffected {
		res["effected"] = *self.Eeffected
	}
	if nil != self.ElastInsertId {
		res["lastInsertId"] = self.ElastInsertId
	}
	if nil != self.Eoptions {
		res["options"] = self.Eoptions
	}
	return res
}

type AnyData struct {
	Value interface{}
}

func (self *AnyData) IsNil() bool {
	return nil == self.Value
}

func (self *AnyData) AsInterface() interface{} {
	return self.Value
}

func (self *AnyData) AsBool() (bool, error) {
	return AsBool(self.Value)
}

func (self *AnyData) AsInt() (int, error) {
	return AsInt(self.Value)
}

func (self *AnyData) AsInt32() (int32, error) {
	return AsInt32(self.Value)
}

func (self *AnyData) AsInt64() (int64, error) {
	return AsInt64(self.Value)
}

func (self *AnyData) AsUint() (uint, error) {
	return AsUint(self.Value)
}

func (self *AnyData) AsUint32() (uint32, error) {
	return AsUint32(self.Value)
}

func (self *AnyData) AsUint64() (uint64, error) {
	return AsUint64(self.Value)
}

func (self *AnyData) AsString() (string, error) {
	return AsString(self.Value)
}

func (self *AnyData) AsArray() ([]interface{}, error) {
	if m, ok := self.Value.([]interface{}); ok {
		return m, nil
	}
	return nil, IsNotArray
}

func (self *AnyData) AsObject() (map[string]interface{}, error) {
	if m, ok := self.Value.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, IsNotMap
}

func (self *AnyData) AsObjects() ([]map[string]interface{}, error) {
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

func ReturnWithNotFoundWithMessage(id, msg string) Result {
	if 0 == len(id) {
		return ReturnError(NotFoundCode, msg)
	}
	return ReturnError(NotFoundCode, "'"+id+"' is not found - "+msg)
}
func ReturnWithNotFound(id string) Result {
	return ReturnError(NotFoundCode, "'"+id+"' is not found.")
}

func ReturnWithRecordNotFound(t, id string) Result {
	return ReturnError(NotFoundCode, t+" with id was '"+id+"' is not found.")
}

func ReturnWithRecordAlreadyExists(id string) Result {
	return ReturnError(NotAcceptableCode, "'"+id+"' is already exists.")
}

func ReturnWithNotImplemented() Result {
	return ReturnError(InternalErrorCode, "not implemented")
}

func ReturnWithServiceUnavailable(msg string) Result {
	return ReturnError(ServiceUnavailableCode, msg)
}
