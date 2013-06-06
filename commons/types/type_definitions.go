package types

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	//"labix.org/v2/mgo/bson"
	"net"
	"regexp"
	"strconv"
	"time"
)

type Password string
type IPAddress string
type PhysicalAddress string

type TypeDefinition interface {
	Name() string
	MakeValue() interface{}
	CreateEnumerationValidator(values []string) (Validator, error)
	CreatePatternValidator(pattern string) (Validator, error)
	CreateRangeValidator(minValue, maxValue string) (Validator, error)
	CreateLengthValidator(minLength, maxLength string) (Validator, error)
	Parse(v string) (interface{}, error)
	ToInternal(v interface{}) (interface{}, error)
	ToExternal(v interface{}) interface{}
}

type integerType struct {
}

func (self *integerType) Name() string {
	return "integer"
}

func (self *integerType) MakeValue() interface{} {
	return &sql.NullInt64{Int64: 0, Valid: false}
}

func (self *integerType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]interface{}, 0, len(ss))
	for i, s := range ss {
		v, err := strconv.ParseInt(s, 10, 64)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, int64(v))
	}
	return &EnumerationValidator{Values: values}, nil
}

func (self *integerType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *integerType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max int64
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = strconv.ParseInt(minValue, 10, 64)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a integer", minValue)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = strconv.ParseInt(maxValue, 10, 64)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a integer", maxValue)
		}
	}
	return &IntegerValidator{HasMax: hasMax, MaxValue: max,
		HasMin: hasMin, MinValue: min}, nil
}

func (self *integerType) CreateLengthValidator(minLength,
	maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *integerType) ToInternal(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return int64(i64), nil
		}
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if 9223372036854775807 >= int64(v) {
			return int64(v), nil
		}
		return int64(0), errors.New("it is uint32, value is overflow.")
	case uint32:
		return int64(v), nil
	case uint64:
		if 9223372036854775807 >= v {
			return int64(v), nil
		}
		return int64(0), errors.New("it is uint64, value is overflow.")
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case []byte:
		i64, err := strconv.ParseInt(string(v), 10, 64)
		if nil == err {
			return int64(i64), nil
		}
	case *int64:
		return *v, nil
	case *sql.NullInt64:
		return v.Value()
	}
	return int64(0), errors.New("ToInternal to int64 failed")
}

func (self *integerType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *integerType) Parse(s string) (interface{}, error) {
	i64, e := strconv.ParseInt(s, 10, 64)
	return i64, e
}

type decimalType struct {
}

func (self *decimalType) Name() string {
	return "decimal"
}

func (self *decimalType) MakeValue() interface{} {
	return &sql.NullFloat64{Float64: 0, Valid: false}
}

func (self *decimalType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]interface{}, 0, len(ss))
	for i, s := range ss {
		v, err := strconv.ParseFloat(s, 64)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, float64(v))
	}
	return &EnumerationValidator{Values: values}, nil
}

func (self *decimalType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *decimalType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max float64
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = strconv.ParseFloat(minValue, 64)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a integer", minValue)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = strconv.ParseFloat(maxValue, 64)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a integer", maxValue)
		}
	}
	return &DecimalValidator{HasMax: hasMax, MaxValue: max, HasMin: hasMin, MinValue: min}, nil
}

func (self *decimalType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *decimalType) ToInternal(value interface{}) (interface{}, error) {

	switch v := value.(type) {
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
	case string:
		f64, err := strconv.ParseFloat(v, 64)
		if nil == err {
			return float64(f64), nil
		}
		return float64(0), err
	case []byte:
		i64, err := strconv.ParseFloat(string(v), 64)
		if nil == err {
			return float64(i64), nil
		}
	case *float64:
		return *v, nil
	case *sql.NullFloat64:
		return v.Value()
	}
	return float64(0), errors.New("ToInternal to float64 failed")
}

func (self *decimalType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *decimalType) Parse(s string) (interface{}, error) {
	f64, e := strconv.ParseFloat(s, 64)
	return f64, e
}

type stringType struct {
}

func (self *stringType) Name() string {
	return "string"
}

func (self *stringType) MakeValue() interface{} {
	return &sql.NullString{String: "", Valid: false}
}

func (self *stringType) CreateEnumerationValidator(values []string) (Validator, error) {
	if nil == values || 0 == len(values) {
		return nil, errors.New("values is null or empty")
	}

	new_values := make([]string, len(values))
	for i, s := range values {
		if "" == s {
			return nil, fmt.Errorf("value[%d] is empty", i)
		}
		new_values = append(new_values, string(s))
	}
	return &StringEnumerationValidator{Values: new_values}, nil
}

func (self *stringType) CreatePatternValidator(pattern string) (Validator, error) {
	if "" == pattern {
		return nil, errors.New("pattern is empty")
	}

	p, err := regexp.Compile(pattern)
	if nil != err {
		return nil, err
	}
	return &PatternValidator{Pattern: p}, nil
}

func (self *stringType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *stringType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	var err error
	var min int64 = -1
	var max int64 = -1

	if "" != minLength {
		min, err = strconv.ParseInt(minLength, 10, 32)
		if nil != err {
			return nil, fmt.Errorf("minLength '%s' is not a integer", minLength)
		}
	}

	if "" != maxLength {
		max, err = strconv.ParseInt(maxLength, 10, 32)
		if nil != err {
			return nil, fmt.Errorf("maxLength '%s' is not a integer", maxLength)
		}
	}
	return &StringLengthValidator{MaxLength: int(max), MinLength: int(min)}, nil
}

func (self *stringType) ToInternal(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case *string:
		return *v, nil
	case []byte:
		return string(v), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(v), 10), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(int64(v), 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case float64:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case *sql.NullString:
		return v.Value()
	}
	return "", errors.New("ToInternal to SqlString failed")
}

func (self *stringType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *stringType) Parse(s string) (interface{}, error) {
	return s, nil
}

type dateTimeType struct {
	Layout string //"2006-01-02 15:04:05"
	name   string //datetime
}

func (self *dateTimeType) Name() string {
	return self.name
}

// NullTime represents an time that may be null.
// NullTime implements the Scanner interface so
// it can be used as a scan destination, similar to NullTime.
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullTime) Scan(value interface{}) error {
	if value == nil {
		n.Time, n.Valid = time.Time{}, false
		return nil
	}

	n.Time, n.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

func (self *dateTimeType) MakeValue() interface{} {
	return &NullTime{Valid: false}
}

func (self *dateTimeType) CreateEnumerationValidator(ss []string) (Validator, error) {
	if nil == ss || 0 == len(ss) {
		return nil, errors.New("values is null or empty.")
	}

	values := make([]interface{}, 0, len(ss))
	for i, s := range ss {
		t, err := time.Parse(self.Layout, s)
		if nil != err {
			return nil, fmt.Errorf("value[%d] '%v' is syntex error, %s", i, s, err.Error())
		}
		values = append(values, t)
	}
	return &EnumerationValidator{Values: values}, nil
}

func (self *dateTimeType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *dateTimeType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	var min, max time.Time
	var err error
	hasMin := false
	hasMax := false

	if "" != minValue {
		hasMin = true
		min, err = time.Parse(self.Layout, minValue)
		if nil != err {
			return nil, fmt.Errorf("minValue '%s' is not a time(%s)", minValue, self.Layout)
		}
	}

	if "" != maxValue {
		hasMax = true
		max, err = time.Parse(self.Layout, maxValue)
		if nil != err {
			return nil, fmt.Errorf("maxValue '%s' is not a time(%s)", maxValue, self.Layout)
		}
	}
	return &DateValidator{HasMax: hasMax, MaxValue: max, HasMin: hasMin, MinValue: min}, nil
}

func (self *dateTimeType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *dateTimeType) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		t, err := time.Parse(self.Layout, value)
		if nil != err {
			return nil, err
		}
		return t, nil
	case *string:
		t, err := time.Parse(self.Layout, *value)
		if nil != err {
			return nil, err
		}
		return t, nil
	case time.Time:
		return value, nil
	case *time.Time:
		return *value, nil
	case *NullTime:
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a datetime")
}

func (self *dateTimeType) ToExternal(value interface{}) interface{} {
	return value
}

func (self *dateTimeType) Parse(s string) (interface{}, error) {
	t, e := time.Parse(self.Layout, s)
	return t, e
}

type ipAddressType struct {
}

func (self *ipAddressType) Name() string {
	return "ipAddress"
}

// NullIPAddress represents an ip address that may be null.
// NullIPAddress implements the Scanner interface so
// it can be used as a scan destination, similar to NullIPAddress.
type NullIPAddress struct {
	String string
	Valid  bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullIPAddress) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	var nullString sql.NullString
	e := nullString.Scan(value)
	if nil == e {
		n.Valid = nullString.Valid
		n.String = nullString.String
	}
	return e
}

// Value implements the driver Valuer interface.
func (n NullIPAddress) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	ip := net.ParseIP(n.String)
	if nil == ip {
		return nil, errors.New("syntex error, it is not IP.")
	}
	return IPAddress(ip.String()), nil
}

func (self *ipAddressType) MakeValue() interface{} {
	return &NullIPAddress{Valid: false}
}

func (self *ipAddressType) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}
func (self *ipAddressType) ToInternal(v interface{}) (interface{}, error) {
	if nil == v {
		return nil, nil
	}

	switch value := v.(type) {
	case string:
		ip := net.ParseIP(value)
		if nil == ip {
			return nil, errors.New("syntex error, it is not IP.")
		}

		addr := IPAddress(ip.String())
		return addr, nil
	case *string:
		ip := net.ParseIP(*value)
		if nil == ip {
			return nil, errors.New("syntex error, it is not IP.")
		}
		addr := IPAddress(ip.String())
		return addr, nil
	case []byte:
		ip := net.ParseIP(string(value))
		if nil == ip {
			return nil, errors.New("syntex error, it is not IP.")
		}
		addr := IPAddress(ip.String())
		return addr, nil
	case net.IP:
		addr := IPAddress(value.String())
		return addr, nil
	case *net.IP:
		addr := IPAddress(value.String())
		return addr, nil
	case IPAddress:
		return value, nil
	case *IPAddress:
		return *value, nil
	case *NullIPAddress:
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a ipAddress")
}

func (self *ipAddressType) ToExternal(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return v
	case IPAddress:
		return string(v)
	default:
		panic("syntex error, it is not a ipAddress")
	}
}

func (self *ipAddressType) Parse(s string) (interface{}, error) {
	ip := net.ParseIP(s)
	if nil == ip {
		return nil, errors.New("syntex error, it is not IP.")
	}
	return IPAddress(ip.String()), nil
}

type physicalAddressType struct {
}

// NullIPAddress represents an ip address that may be null.
// NullIPAddress implements the Scanner interface so
// it can be used as a scan destination, similar to NullIPAddress.
type NullPhysicalAddress struct {
	String string
	Valid  bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullPhysicalAddress) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	var nullString sql.NullString
	e := nullString.Scan(value)
	if nil == e {
		n.Valid = nullString.Valid
		n.String = nullString.String
	}
	return e
}

// Value implements the driver Valuer interface.
func (n NullPhysicalAddress) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	if 0 == len(n.String) {
		return nil, nil
	}
	mac, err := net.ParseMAC(n.String)
	if nil != err {
		return nil, err
	}
	return PhysicalAddress(mac.String()), nil
}

func (self *physicalAddressType) MakeValue() interface{} {
	return &NullPhysicalAddress{Valid: false}
}

func (self *physicalAddressType) Name() string {
	return "physicalAddress"
}

func (self *physicalAddressType) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *physicalAddressType) ToInternal(v interface{}) (interface{}, error) {
	if nil == v {
		return nil, nil
	}
	switch value := v.(type) {
	case string:
		if "" == value {
			return nil, nil
		}
		mac, err := net.ParseMAC(value)
		if nil != err {
			return nil, err
		}
		return PhysicalAddress(mac.String()), nil
	case *string:
		if "" == *value {
			return nil, nil
		}
		mac, err := net.ParseMAC(*value)
		if nil != err {
			return nil, err
		}
		return PhysicalAddress(mac.String()), nil
	case []byte:
		if nil == value || 0 == len(value) {
			return nil, nil
		}
		mac, err := net.ParseMAC(string(value))
		if nil != err {
			return nil, err
		}
		return PhysicalAddress(mac.String()), nil
	case net.HardwareAddr:
		return PhysicalAddress(value.String()), nil
	case *net.HardwareAddr:
		return PhysicalAddress(value.String()), nil
	case PhysicalAddress:
		return value, nil
	case *PhysicalAddress:
		return *value, nil
	case *NullPhysicalAddress:
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a physicalAddress")
}

func (self *physicalAddressType) ToExternal(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return v
	case PhysicalAddress:
		return string(v)
	default:
		panic("syntex error, it is not a physicalAddress")
	}
}

func (self *physicalAddressType) Parse(s string) (interface{}, error) {
	mac, e := net.ParseMAC(s)
	return mac, e
}

type booleanType struct {
}

func (self *booleanType) Name() string {
	return "boolean"
}

func (self *booleanType) MakeValue() interface{} {
	return &sql.NullBool{Valid: false}
}

func (self *booleanType) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *booleanType) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		return self.Parse(value)
	case *string:
		return self.Parse(*value)
	case []byte:
		return self.Parse(string(value))
	case bool:
		return value, nil
	case *bool:
		return *value, nil
	case *sql.NullBool:
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a boolean")
}

func (self *booleanType) ToExternal(v interface{}) interface{} {
	return v
}

func (self *booleanType) Parse(s string) (interface{}, error) {
	switch s {
	case "true", "True", "TRUE", "yes", "Yes", "YES", "1":
		return true, nil
	case "false", "False", "FALSE", "no", "No", "NO", "0":
		return false, nil
	default:
		return nil, errors.New("syntex error, it is not a boolean")
	}
}

type objectIdType struct {
	TypeDefinition
}

func (self *objectIdType) Name() string {
	return "objectId"
}

// type objectIdType struct {
// }

// func (self *objectIdType) Name() string {
// 	return "objectId"
// }

// func (self *objectIdType) CreateEnumerationValidator(values []string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) CreatePatternValidator(pattern string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
// 	panic("not supported")
// }

// func (self *objectIdType) ToInternal(v interface{}) (interface{}, error) {
// 	switch value := v.(type) {
// 	case string:
// 		return parseObjectIdHex(value)
// 	case *string:
// 		return parseObjectIdHex(*value)
// 	case bson.ObjectId:
// 		return value, nil
// 	case *bson.ObjectId:
// 		return *value, nil
// 	}

// 	return nil, errors.New("syntex error, it is not a boolean")
// }

type SqlIdTypeDefinition struct {
}

func (self *SqlIdTypeDefinition) Name() string {
	return "objectId"
}

func (self *SqlIdTypeDefinition) MakeValue() interface{} {
	return &sql.NullInt64{Valid: false}
}

func (self *SqlIdTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *SqlIdTypeDefinition) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		i64, e := strconv.ParseInt(value, 10, 0)
		return int(i64), e
	case *string:
		i64, e := strconv.ParseInt(*value, 10, 0)
		return int(i64), e
	case []byte:
		i64, e := strconv.ParseInt(string(value), 10, 0)
		return int(i64), e
	case int:
		return value, nil
	case *int:
		return *value, nil
	case int32:
		return value, nil
	case *int32:
		return *value, nil
	case int64:
		return value, nil
	case *int64:
		return *value, nil
	case *sql.NullInt64:
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a objectId")
}

func (self *SqlIdTypeDefinition) ToExternal(v interface{}) interface{} {
	return v
}

func (self *SqlIdTypeDefinition) Parse(s string) (interface{}, error) {
	i64, e := strconv.ParseInt(s, 10, 64)
	return i64, e
}

type passwordType struct {
	stringType
}

func (self *passwordType) Name() string {
	return "password"
}

// NullIPAddress represents an ip address that may be null.
// NullIPAddress implements the Scanner interface so
// it can be used as a scan destination, similar to NullIPAddress.
type NullPassword struct {
	String string
	Valid  bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullPassword) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	var nullString sql.NullString
	e := nullString.Scan(value)
	if nil == e {
		n.Valid = nullString.Valid
		n.String = nullString.String
	}
	return e
}

// Value implements the driver Valuer interface.
func (n NullPassword) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return Password(n.String), nil
}

func (self *passwordType) MakeValue() interface{} {
	return &NullPassword{Valid: false}
}

func (self *passwordType) ToInternal(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		return Password(value), nil
	case *string:
		return Password(*value), nil
	case []byte:
		return Password(string(value)), nil
	case Password:
		return value, nil
	case *Password:
		return *value, nil
	case *NullPassword:
		return value.Value()
	}

	return nil, errors.New("syntex error, it is not a password")
}

func (self *passwordType) ToExternal(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return v
	case Password:
		return string(v)
	default:
		panic("syntex error, it is not a password")
	}
}

func (self *passwordType) Parse(s string) (interface{}, error) {
	return Password(s), nil
}

var (
	BooleanType         booleanType
	IntegerType         integerType
	DecimalType         decimalType
	StringType          stringType
	DateTimeType        dateTimeType
	IPAddressType       ipAddressType
	PhysicalAddressType physicalAddressType
	PasswordType        passwordType
	ObjectIdType        objectIdType

	DATETIMELAYOUT string

	types = map[string]TypeDefinition{"boolean": &BooleanType,
		"integer":         &IntegerType,
		"decimal":         &DecimalType,
		"string":          &StringType,
		"datetime":        &DateTimeType,
		"ipAddress":       &IPAddressType,
		"physicalAddress": &PhysicalAddressType,
		"password":        &PasswordType,
		"objectId":        &ObjectIdType}
)

func init() {
	DATETIMELAYOUT = time.RFC3339
	DateTimeType.Layout = DATETIMELAYOUT // `"` + time.RFC3339 + `"`
	ObjectIdType.TypeDefinition = &SqlIdTypeDefinition{}
}

func RegisterTypeDefinition(t TypeDefinition) {
	types[t.Name()] = t
}

func GetTypeDefinition(t string) TypeDefinition {
	return types[t]
}
