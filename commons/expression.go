package commons

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var InvalidIndex = errors.New("invalid index")
var IndexOutOfRange = errors.New("index out of range")
var ValueIsNilInResult = errors.New("value is nil in the result")
var IsNull = ValueIsNil
var ValueNotFound = NotExists
var notImplemented = NotImplemented

func where(value interface{}, expression string) ([]interface{}, error) {
	return nil, notImplemented
}

func whereOne(value interface{}, expression string) (interface{}, error) {
	i, e := strconv.ParseInt(expression, 10, 0)
	if nil != e {
		return nil, e
	}
	if i < 0 {
		return nil, InvalidIndex
	}
	idx := int(i)

	if array, ok := value.([]interface{}); ok {
		if idx < len(array) {
			return array[idx], nil
		}
		return nil, IndexOutOfRange
	} else if array, ok := value.([]map[string]interface{}); ok {
		if idx < len(array) {
			return array[idx], nil
		}
		return nil, IndexOutOfRange
	} else if array, ok := value.([]map[string]string); ok {
		if idx < len(array) {
			return array[idx], nil
		}
		return nil, IndexOutOfRange
	}
	return nil, fmt.Errorf("it is not a slice, actual is &T", value)
}

func ToSimpleValue(value interface{}, attribute string) (interface{}, error) {
	if nil == value {
		return nil, IsNull
	}

	currentValue := value
	if current, ok := value.(Result); ok {
		currentValue = current.InterfaceValue()
		if nil == currentValue {
			return nil, ValueIsNilInResult
		}
	}

	if 0 == len(attribute) {
		return currentValue, nil
	}

	if '[' != attribute[0] {
		return getValueByField(currentValue, attribute)
	}

	idx := strings.IndexRune(attribute, ']')
	if -1 == idx {
		return nil, errors.New("sytex error: '" + attribute + "'")
	}

	var e error
	currentValue, e = whereOne(currentValue, attribute[1:idx])
	if nil != e {
		return nil, e
	}

	if (idx + 1) == len(attribute) {
		return currentValue, nil
	}

	if '.' != attribute[idx+1] || (idx+2) == len(attribute) {
		return nil, errors.New("sytex error: '" + attribute + "'")
	}

	return getValueByField(currentValue, attribute[idx+2:])
}

func getValueByField(currentValue interface{}, attribute string) (interface{}, error) {
	if m, ok := currentValue.(map[string]interface{}); ok {
		if currentValue, ok = m[attribute]; ok {
			return currentValue, nil
		} else {
			return nil, ValueNotFound
		}
	} else if m, ok := currentValue.(map[string]string); ok {
		if currentValue, ok = m[attribute]; ok {
			return currentValue, nil
		} else {
			return nil, ValueNotFound
		}
	}
	return nil, fmt.Errorf("value is not a map, actual is %T", currentValue)
}
