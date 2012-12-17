package mdb

import (
	"errors"
	"fmt"
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
		values = append(values, SqlInteger64(v))
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
		return SqlInteger64(v), nil
	case int8:
		return SqlInteger64(v), nil
	case int16:
		return SqlInteger64(v), nil
	case int32:
		return SqlInteger64(v), nil
	case int64:
		return SqlInteger64(v), nil
	case uint:
		if 9223372036854775807 >= int64(v) {
			return SqlInteger64(v), nil
		}
		return SqlInteger64(0), errors.New("it is uint32, value is overflow.")
	case uint8:
		return SqlInteger64(v), nil
	case uint16:
		return SqlInteger64(v), nil
	case uint32:
		return SqlInteger64(v), nil
	case uint64:
		if 9223372036854775807 >= v {
			return SqlInteger64(v), nil
		}
		return SqlInteger64(0), errors.New("it is uint64, value is overflow.")
	case float32:
		return SqlInteger64(v), nil
	case float64:
		return SqlInteger64(v), nil
	case string:
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return SqlInteger64(i64), nil
		}
	case SqlInteger64:
		return v, nil
	case *SqlInteger64:
		return *v, nil
	}
	return SqlInteger64(0), errors.New("convert to SqlInteger64 failed")
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
		values = append(values, SqlDecimal(v))
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
		return SqlDecimal(v), nil
	case uint8:
		return SqlDecimal(v), nil
	case uint16:
		return SqlDecimal(v), nil
	case uint32:
		return SqlDecimal(v), nil
	case uint64:
		return SqlDecimal(v), nil
	case int:
		return SqlDecimal(v), nil
	case int8:
		return SqlDecimal(v), nil
	case int16:
		return SqlDecimal(v), nil
	case int32:
		return SqlDecimal(v), nil
	case int64:
		return SqlDecimal(v), nil
	case float32:
		return SqlDecimal(v), nil
	case float64:
		return SqlDecimal(v), nil
	case string:
		f64, err := strconv.ParseFloat(v, 64)
		if nil == err {
			return SqlDecimal(f64), nil
		}
		return SqlDecimal(0), err
	case SqlDecimal:
		return v, nil
	case *SqlDecimal:
		return *v, nil
	}
	return SqlDecimal(0), errors.New("convert to SqlDecimal failed")
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

	new_values := make([]SqlString, len(values))
	for i, s := range values {
		if "" == s {
			return nil, fmt.Errorf("value[%d] is empty", i)
		}
		new_values = append(new_values, SqlString(s))
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
		return SqlString(v), nil
	case *string:
		return SqlString(*v), nil
	case uint:
		return SqlString(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return SqlString(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return SqlString(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return SqlString(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return SqlString(strconv.FormatUint(uint64(v), 10)), nil
	case int:
		return SqlString(strconv.FormatInt(int64(v), 10)), nil
	case int8:
		return SqlString(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return SqlString(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return SqlString(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return SqlString(strconv.FormatInt(int64(v), 10)), nil
	case float32:
		return SqlString(strconv.FormatFloat(float64(v), 'e', -1, 64)), nil
	case float64:
		return SqlString(strconv.FormatFloat(float64(v), 'e', -1, 64)), nil
	case SqlString:
		return v, nil
	case *SqlString:
		return *v, nil
	}
	return SqlString(""), errors.New("convert to SqlString failed")
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
		values = append(values, SqlDateTime(t))
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
		return SqlDateTime(t), nil
	case *string:
		t, err := time.Parse(self.Layout, *value)
		if nil != err {
			return nil, err
		}
		return SqlDateTime(t), nil
	case time.Time:
		return SqlDateTime(value), nil
	case *time.Time:
		return SqlDateTime(*value), nil
	case SqlDateTime:
		return value, nil
	case *SqlDateTime:
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
		return SqlIPAddress(ip), nil
	case *string:
		ip := net.ParseIP(*value)
		if nil == ip {
			return nil, errors.New("syntex error, it is not IP.")
		}
		return SqlIPAddress(ip), nil
	case net.IP:
		return SqlIPAddress(value), nil
	case *net.IP:
		return SqlIPAddress(*value), nil
	case SqlIPAddress:
		return value, nil
	case *SqlIPAddress:
		return *value, nil
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
		return SqlPhysicalAddress(mac), nil
	case *string:
		mac, err := net.ParseMAC(*value)
		if nil != err {
			return nil, err
		}
		return SqlPhysicalAddress(mac), nil
	case net.HardwareAddr:
		return SqlPhysicalAddress(value), nil
	case *net.HardwareAddr:
		return SqlPhysicalAddress(*value), nil
	case SqlPhysicalAddress:
		return value, nil
	case *SqlPhysicalAddress:
		return *value, nil
	}

	return nil, errors.New("syntex error, it is not a string")
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
	integerType         IntegerTypeDefinition
	decimalType         DecimalTypeDefinition
	stringType          StringTypeDefinition
	datetimeType        DateTimeTypeDefinition
	ipAddressType       IpAddressTypeDefinition
	physicalAddressType PhysicalAddressTypeDefinition
	passwordType        PasswordTypeDefinition
)

func init() {
	datetimeType.Layout = "2006-01-02 15:04:05"
}

func GetTypeDefinition(t string) TypeDefinition {
	switch t {
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
	}
	return nil
}
