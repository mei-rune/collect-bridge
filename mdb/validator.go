package mdb

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"time"
)

type Validator interface {
	Validate(value interface{}) (bool, error)
}

type StringValidator struct {
	MinLength, MaxLength int
	Pattern              *regexp.Regexp
}

func (self *StringValidator) Validate(obj interface{}) (bool, error) {
	value, ok := obj.(string)
	if !ok {
		return false, errors.New("syntex error")
	}

	if 0 <= self.MinLength && self.MinLength > len(value) {
		return false, errors.New("length of '" + value + "' is less " + strconv.Itoa(self.MinLength))
	}

	if 0 <= self.MaxLength && self.MaxLength < len(value) {
		return false, errors.New("length of '" + value + "' is greate " + strconv.Itoa(self.MaxLength))
	}

	if nil != self.Pattern {
		if !self.Pattern.MatchString(value) {
			return false, errors.New("'" + value + "' is not match '" + self.Pattern.String() + "'")
		}
	}
	return true, nil
}

type PhysicalAddressValidator struct{}

func (self *PhysicalAddressValidator) Validate(obj interface{}) (bool, error) {
	value, ok := obj.(string)
	if !ok {
		return false, errors.New("syntex error")
	}

	if _, err := net.ParseMAC(value); nil != err {
		return false, errors.New("syntex error, " + err.Error())
	}
	return true, nil
}

type IPAddressValidator struct{}

func (self *IPAddressValidator) Validate(obj interface{}) (bool, error) {
	value, ok := obj.(string)
	if !ok {
		return false, errors.New("syntex error")
	}
	if nil == net.ParseIP(value) {
		return false, errors.New("syntex error")
	}
	return true, nil
}

type IntegerValidator struct {
	HasMin, HasMax     bool
	MinValue, MaxValue int64
}

func (self *IntegerValidator) Validate(obj interface{}) (bool, error) {
	var value int64
	switch v := obj.(type) {
	case int:
		value = int64(v)
	case int8:
		value = int64(v)
	case int16:
		value = int64(v)
	case int32:
		value = int64(v)
	case int64:
		value = int64(v)
	default:
		return false, errors.New("syntex error")
	}

	if self.HasMin && self.MinValue > value {
		return false, fmt.Errorf("'%d' is less minValue '%d'", value, self.MinValue)
	}

	if self.HasMax && self.MaxValue < value {
		return false, fmt.Errorf("'%d' is greate maxValue '%d'", value, self.MaxValue)
	}

	return true, nil
}

type EnumerationValidator struct {
	Values []interface{}
}

func (self *EnumerationValidator) Validate(obj interface{}) (bool, error) {
	var found bool = false
	for v := range self.Values {
		if v == obj {
			found = true
			break
		}
	}
	if !found {
		return false, fmt.Errorf("enum is not contains %v", obj)
	}
	return true, nil
}

type DecimalValidator struct {
	HasMin, HasMax     bool
	MinValue, MaxValue float64
}

func (self *DecimalValidator) Validate(obj interface{}) (bool, error) {
	var value float64
	switch v := obj.(type) {
	case uint:
		value = float64(v)
	case uint8:
		value = float64(v)
	case uint16:
		value = float64(v)
	case uint32:
		value = float64(v)
	case uint64:
		value = float64(v)
	case int:
		value = float64(v)
	case int8:
		value = float64(v)
	case int16:
		value = float64(v)
	case int32:
		value = float64(v)
	case int64:
		value = float64(v)
	case float32:
		value = float64(v)
	case float64:
		value = float64(v)
	default:
		return false, errors.New("syntex error")
	}

	if self.HasMin && self.MinValue > value {
		return false, fmt.Errorf("'%f' is less minValue '%f'", value, self.MinValue)
	}

	if self.HasMax && self.MaxValue < value {
		return false, fmt.Errorf("'%f' is greate maxValue '%f'", value, self.MaxValue)
	}
	return true, nil
}

type DateValidator struct {
	HasMin, HasMax     bool
	MinValue, MaxValue time.Time
}

func (self *DateValidator) Validate(obj interface{}) (bool, error) {
	value, ok := obj.(time.Time)
	if !ok {
		return false, errors.New("syntex error")
	}

	if self.HasMin && self.MinValue.After(value) {
		return false, fmt.Errorf("'%s' is less minValue '%s'", value.String(), self.MinValue.String())
	}

	if self.HasMax && self.MaxValue.Before(value) {
		return false, fmt.Errorf("'%s' is greate maxValue '%s'", value.String(), self.MaxValue.String())
	}
	return true, nil
}
