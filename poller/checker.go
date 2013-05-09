package poller

import (
	"commons"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type Checker interface {
	Run(value interface{}, res map[string]interface{}) bool
}

type jsonFunc func(value interface{}, res map[string]interface{}) bool

func (self jsonFunc) Run(value interface{}, res map[string]interface{}) bool {
	return self(value, res)
}

type jsonExpression struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
}

func makeJsonChecker(code stirng) (Checker, error) {
	var exp jsonExpression
	if e := json.Unmarshal(code, &exp); nil != e {
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
			return floatf(exp.Attribute, f64)
		}
	} else if intf, ok := int_ops[exp.Value]; ok {
		i64, e := strconv.ParseInt(exp.Value, 64)
		if nil != e {
			return nil, errors.New("'value' is not a int64, " + e.Error())
		}
		return intf(exp.Attribute, i64)
	}

	if stringf, ok := string_ops[exp.Value]; ok {
		return stringf(exp.Attribute, exp.Value)
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

func get_int64(value interface{}, name string) int64 {
	m := value.(map[string]interface{})
	return commons.GetInt64(m, name)
}

func int_gt(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_int64(attribute, value) > operand
	}
}
func int_gte(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_int64(attribute, value) >= operand
	}
}
func int_lt(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_int64(attribute, value) < operand
	}
}
func int_lte(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_int64(attribute, value) <= operand
	}
}
func int_not_equals(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_int64(attribute, value) != operand
	}
}
func int_equals(attribute string, operand int64) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_int64(attribute, value) == operand
	}
}

func get_float(value interface{}, name string) float {
	m := value.(map[string]interface{})
	return commons.Getfloat(m, name)
}

func float_gt(attribute string, operand float) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_float(attribute, value) > operand
	}
}
func float_gte(attribute string, operand float) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_float(attribute, value) >= operand
	}
}
func float_lt(attribute string, operand float) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_float(attribute, value) < operand
	}
}
func float_lte(attribute string, operand float) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_float(attribute, value) <= operand
	}
}
func float_not_equals(attribute string, operand float) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_float(attribute, value) != operand
	}
}
func float_equals(attribute string, operand float) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_float(attribute, value) == operand
	}
}

func get_string(value interface{}, name string) string {
	m := value.(map[string]interface{})
	return commons.GetString(m, name)
}

func string_not_contains(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return !strings.Contains(get_string(attribute, value), operand)
	}
}
func string_contains(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return strings.Contains(get_string(attribute, value), operand)
	}
}
func str_not_equals(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_string(attribute, value) != operand
	}
}
func str_equals(attribute string, operand string) jsonFunc {
	return func(value interface{}, res map[string]interface{}) bool {
		return get_string(attribute, value) == operand
	}
}
