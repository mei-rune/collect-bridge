package poller

import (
	"commons"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Checker interface {
	Run(value interface{}, res map[string]interface{}) (int, error)
}

type jsonFunc func(value interface{}, res map[string]interface{}) (bool, error)

func (f jsonFunc) Run(value interface{}, res map[string]interface{}) (int, error) {
	r, e := f(value, res)
	if nil != e {
		return 0, e
	}

	if r {
		return 1, nil
	} else {
		return 0, nil
	}
}

type jsonExpression struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
}

func makeJsonChecker(code string) (Checker, error) {
	var exp jsonExpression
	if e := json.Unmarshal([]byte(code), &exp); nil != e {
		return nil, errors.New("unmarshal expression failed, " + e.Error())
	}

	if "" == exp.Attribute {
		return nil, errors.New("'attribute' is required")
	}
	if "" == exp.Operator {
		return nil, errors.New("'operator' is required")
	}
	if "" == exp.Value {
		return nil, errors.New("'value' is required")
	}

	if strings.ContainsRune(exp.Value, '.') {
		if floatf, ok := float_ops[exp.Value]; ok {
			f64, e := strconv.ParseFloat(exp.Value, 64)
			if nil != e {
				return nil, errors.New("'value' is not a float, " + e.Error())
			}
			return floatf(exp.Attribute, f64), nil
		}
	} else if intf, ok := int_ops[exp.Value]; ok {
		i64, e := strconv.ParseInt(exp.Value, 10, 64)
		if nil != e {
			return nil, errors.New("'value' is not a int64, " + e.Error())
		}
		return intf(exp.Attribute, i64), nil
	}

	if stringf, ok := string_ops[exp.Value]; ok {
		return stringf(exp.Attribute, exp.Value), nil
	}

	return nil, errors.New("operator '" + exp.Operator + "' is unknown")
}

var (
	int_ops = map[string]func(string, int64) jsonFunc{
		">":  int_gt,
		">=": int_gte,
		"<":  int_lt,
		"<=": int_lte,
		"!=": int_not_equals,
		"=":  int_equals,
		"==": int_equals}

	float_ops = map[string]func(string, float64) jsonFunc{
		">":  float_gt,
		">=": float_gte,
		"<":  float_lt,
		"<=": float_lte,
		"!=": float_not_equals,
		"=":  float_equals,
		"==": float_equals}

	string_ops = map[string]func(string, string) jsonFunc{
		"not_contains": string_not_contains,
		"contains":     string_contains,
		"not_equals":   str_not_equals,
		"equals":       str_equals}
)

func get_int64(value interface{}, name string) (int64, error) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("value is not a map[string]interface{}, actual is %T", value)
	}
	return commons.TryGetInt64(m, name)
}

func int_gt(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		i64, e := get_int64(value, attribute)
		if nil != e {
			return false, e
		}
		return i64 > operand, nil
	}
}
func int_gte(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		i64, e := get_int64(value, attribute)
		if nil != e {
			return false, e
		}
		return i64 >= operand, nil
	}
}
func int_lt(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		i64, e := get_int64(value, attribute)
		if nil != e {
			return false, e
		}
		return i64 < operand, nil
	}
}
func int_lte(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		i64, e := get_int64(value, attribute)
		if nil != e {
			return false, e
		}
		return i64 <= operand, nil
	}
}
func int_not_equals(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		i64, e := get_int64(value, attribute)
		if nil != e {
			return false, e
		}
		return i64 != operand, nil
	}
}
func int_equals(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		i64, e := get_int64(value, attribute)
		if nil != e {
			return false, e
		}
		return i64 == operand, nil
	}
}

func get_float(value interface{}, name string) (float64, error) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("value is not a map[string]interface{}, actual is %T", value)
	}
	return commons.TryGetFloat(m, name)
}

func float_gt(attribute string, operand float64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		f64, e := get_float(value, attribute)
		if nil != e {
			return false, e
		}
		return f64 > operand, nil
	}
}
func float_gte(attribute string, operand float64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		f64, e := get_float(value, attribute)
		if nil != e {
			return false, e
		}
		return f64 >= operand, nil
	}
}
func float_lt(attribute string, operand float64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		f64, e := get_float(value, attribute)
		if nil != e {
			return false, e
		}
		return f64 < operand, nil
	}
}
func float_lte(attribute string, operand float64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		f64, e := get_float(value, attribute)
		if nil != e {
			return false, e
		}
		return f64 <= operand, nil
	}
}
func float_not_equals(attribute string, operand float64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		f64, e := get_float(value, attribute)
		if nil != e {
			return false, e
		}
		return f64 != operand, nil
	}
}
func float_equals(attribute string, operand float64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		f64, e := get_float(value, attribute)
		if nil != e {
			return false, e
		}
		return f64 == operand, nil
	}
}

func get_string(value interface{}, name string) (string, error) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("value is not a map[string]interface{}, actual is %T", value)
	}
	return commons.TryGetString(m, name)
}

func string_not_contains(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		v, e := get_string(value, attribute)
		if nil != e {
			return false, e
		}
		return !strings.Contains(v, operand), nil
	}
}
func string_contains(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		v, e := get_string(value, attribute)
		if nil != e {
			return false, e
		}
		return strings.Contains(v, operand), nil
	}
}
func str_not_equals(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		v, e := get_string(value, attribute)
		if nil != e {
			return false, e
		}
		return v != operand, nil
	}
}
func str_equals(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) (bool, error) {
		v, e := get_string(value, attribute)
		if nil != e {
			return false, e
		}
		return v == operand, nil
	}
}
