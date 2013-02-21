package mdb

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo/bson"
	"net"
	"regexp"
	"strconv"
	"time"
)

type TypeDefinition interface {
	Name() string
	CreateEnumerationValidator(values []string) (Validator, error)
	CreatePatternValidator(pattern string) (Validator, error)
	CreateRangeValidator(minValue, maxValue string) (Validator, error)
	CreateLengthValidator(minLength, maxLength string) (Validator, error)
	Convert(v interface{}) (interface{}, error)
}

type IntegerTypeDefinition struct {
}

func (self *IntegerTypeDefinition) Name() string {
	return "integer"
}

func (self *IntegerTypeDefinition) CreateEnumerationValidator(ss []string) (Validator, error) {
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

func (self *IntegerTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *IntegerTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
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

func (self *IntegerTypeDefinition) CreateLengthValidator(minLength,
	maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *IntegerTypeDefinition) Convert(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
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
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
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
	case string:
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return int64(i64), nil
		}
	case *int64:
		return *v, nil
	}
	return int64(0), errors.New("convert to int64 failed")
}

type DecimalTypeDefinition struct {
}

func (self *DecimalTypeDefinition) Name() string {
	return "decimal"
}

func (self *DecimalTypeDefinition) CreateEnumerationValidator(ss []string) (Validator, error) {
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

func (self *DecimalTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *DecimalTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
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

func (self *DecimalTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *DecimalTypeDefinition) Convert(value interface{}) (interface{}, error) {

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
	case *float64:
		return *v, nil
	}
	return float64(0), errors.New("convert to float64 failed")
}

type StringTypeDefinition struct {
}

func (self *StringTypeDefinition) Name() string {
	return "string"
}

func (self *StringTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
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

func (self *StringTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	if "" == pattern {
		return nil, errors.New("pattern is empty")
	}

	p, err := regexp.Compile(pattern)
	if nil != err {
		return nil, err
	}
	return &PatternValidator{Pattern: p}, nil
}

func (self *StringTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *StringTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
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

func (self *StringTypeDefinition) Convert(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case *string:
		return *v, nil
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
	}
	return "", errors.New("convert to SqlString failed")
}

type DateTimeTypeDefinition struct {
	Layout string //"2006-01-02 15:04:05"
	name   string //datetime
}

func (self *DateTimeTypeDefinition) Name() string {
	return self.name
}

func (self *DateTimeTypeDefinition) CreateEnumerationValidator(ss []string) (Validator, error) {
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

func (self *DateTimeTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *DateTimeTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
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

func (self *DateTimeTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *DateTimeTypeDefinition) Convert(v interface{}) (interface{}, error) {
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
	}

	return nil, errors.New("syntex error, it is not a string")
}

type IpAddressTypeDefinition struct {
}

func (self *IpAddressTypeDefinition) Name() string {
	return "ipAddress"
}

func (self *IpAddressTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}
func (self *IpAddressTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}
func (self *IpAddressTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}
func (self *IpAddressTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}
func (self *IpAddressTypeDefinition) Convert(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		ip := net.ParseIP(value)
		if nil == ip {
			return nil, errors.New("syntex error, it is not IP.")
		}

		addr := SqlIPAddress(ip)
		return &addr, nil
	case *string:
		ip := net.ParseIP(*value)
		if nil == ip {
			return nil, errors.New("syntex error, it is not IP.")
		}
		addr := SqlIPAddress(ip)
		return &addr, nil
	case net.IP:
		addr := SqlIPAddress(value)
		return &addr, nil
	case *net.IP:
		addr := SqlIPAddress(*value)
		return &addr, nil
	case SqlIPAddress:
		return &value, nil
	case *SqlIPAddress:
		return value, nil
	}

	return nil, errors.New("syntex error, it is not a string")
}

type PhysicalAddressTypeDefinition struct {
}

func (self *PhysicalAddressTypeDefinition) Name() string {
	return "physicalAddress"
}

func (self *PhysicalAddressTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *PhysicalAddressTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *PhysicalAddressTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *PhysicalAddressTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *PhysicalAddressTypeDefinition) Convert(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		mac, err := net.ParseMAC(value)
		if nil != err {
			return nil, err
		}
		addr := SqlPhysicalAddress(mac)
		return &addr, nil
	case *string:
		mac, err := net.ParseMAC(*value)
		if nil != err {
			return nil, err
		}
		addr := SqlPhysicalAddress(mac)
		return &addr, nil
	case net.HardwareAddr:
		addr := SqlPhysicalAddress(value)
		return &addr, nil
	case *net.HardwareAddr:
		addr := SqlPhysicalAddress(*value)
		return &addr, nil
	case SqlPhysicalAddress:
		return &value, nil
	case *SqlPhysicalAddress:
		return value, nil
	}

	return nil, errors.New("syntex error, it is not a string")
}

type BooleanTypeDefinition struct {
}

func (self *BooleanTypeDefinition) Name() string {
	return "boolean"
}

func (self *BooleanTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *BooleanTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *BooleanTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *BooleanTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *BooleanTypeDefinition) Convert(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		switch value {
		case "true", "True", "TRUE", "yes", "Yes", "YES":
			return true, nil
		case "false", "False", "FALSE", "no", "No", "NO":
			return false, nil
		}
	case *string:
		switch *value {
		case "true", "True", "TRUE", "yes", "Yes", "YES":
			return true, nil
		case "false", "False", "FALSE", "no", "No", "NO":
			return false, nil
		}
	case bool:
		return value, nil
	case *bool:
		return *value, nil
	}

	return nil, errors.New("syntex error, it is not a boolean")
}

type ObjectIdTypeDefinition struct {
}

func (self *ObjectIdTypeDefinition) Name() string {
	return "objectId"
}

func (self *ObjectIdTypeDefinition) CreateEnumerationValidator(values []string) (Validator, error) {
	panic("not supported")
}

func (self *ObjectIdTypeDefinition) CreatePatternValidator(pattern string) (Validator, error) {
	panic("not supported")
}

func (self *ObjectIdTypeDefinition) CreateRangeValidator(minValue, maxValue string) (Validator, error) {
	panic("not supported")
}

func (self *ObjectIdTypeDefinition) CreateLengthValidator(minLength, maxLength string) (Validator, error) {
	panic("not supported")
}

func (self *ObjectIdTypeDefinition) Convert(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		return parseObjectIdHex(value)
	case *string:
		return parseObjectIdHex(*value)
	case bson.ObjectId:
		return value, nil
	case *bson.ObjectId:
		return *value, nil
	}

	return nil, errors.New("syntex error, it is not a boolean")
}

type PasswordTypeDefinition struct {
	StringTypeDefinition
}

func (self *PasswordTypeDefinition) Name() string {
	return "password"
}

func (self *PasswordTypeDefinition) Convert(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case string:
		return SqlPassword(value), nil
	case SqlPassword:
		return value, nil
	case *SqlPassword:
		return *value, nil
	}

	return nil, errors.New("syntex error, it is not a string")
}

var (
	booleanType         BooleanTypeDefinition
	integerType         IntegerTypeDefinition
	decimalType         DecimalTypeDefinition
	stringType          StringTypeDefinition
	datetimeType        DateTimeTypeDefinition
	ipAddressType       IpAddressTypeDefinition
	physicalAddressType PhysicalAddressTypeDefinition
	passwordType        PasswordTypeDefinition
	objectIdType        ObjectIdTypeDefinition
)

func init() {
	datetimeType.Layout = time.RFC3339 // `"` + time.RFC3339 + `"`
}

func GetTypeDefinition(t string) TypeDefinition {
	switch t {
	case "boolean":
		return &booleanType
	case "integer":
		return &integerType
	case "decimal":
		return &decimalType
	case "string":
		return &stringType
	case "datetime":
		return &datetimeType
	case "ipAddress":
		return &ipAddressType
	case "physicalAddress":
		return &physicalAddressType
	case "password":
		return &passwordType
	case "objectId":
		return &objectIdType
	}
	return nil
}
