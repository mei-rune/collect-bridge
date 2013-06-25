package metrics

import (
	"commons"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type MatchFunc func(matcher *Matcher, params commons.Map, value string) bool

type Matcher struct {
	Method      string
	Attribute   string
	Arguments   []string
	valueMap    map[string]string
	re          *regexp.Regexp
	f           MatchFunc
	Description string
}

func toString(method string, arguments []string) string {
	return fmt.Sprintf("%s(%s)", method, strings.Join(arguments, ","))
}

func checkArguments(arguments []string, excepted int) error {
	if nil == arguments {
		return fmt.Errorf("wrong number of arguments (0 for %d)", excepted)
	}
	if excepted != len(arguments) {
		return fmt.Errorf("wrong number of arguments (%d for %d)", len(arguments), excepted)
	}
	return nil
}

func toLower(arguments []string, offset int) []string {
	res := make([]string, len(arguments))
	for i, v := range arguments[offset:] {
		res[i] = strings.ToLower(v)
	}
	return res
}
func dup(arguments []string, offset int) []string {
	res := make([]string, len(arguments)-1)
	copy(res, arguments[offset:])
	return res
}

func NewMatcher(method string, arguments []string) (*Matcher, error) {

	res := &Matcher{Method: method,
		Arguments:   arguments,
		Description: toString(method, arguments)}

	switch method {
	case "in":
		if nil == arguments || 1 >= len(arguments) {
			return nil, errors.New("wrong number of arguments (0 for 2+)")
		}
		res.Attribute = arguments[0]
		res.Arguments = dup(arguments, 1)
		res.valueMap = make(map[string]string)
		for _, s := range res.Arguments {
			res.valueMap[s] = s
		}
		res.f = match_in
		return res, nil
	case "equal":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = dup(arguments, 1)
		res.f = match_equal
		return res, nil
	case "start_with":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = dup(arguments, 1)
		res.f = match_start_with
		return res, nil
	case "end_with":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = dup(arguments, 1)
		res.f = match_end_with
		return res, nil
	case "contains":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = dup(arguments, 1)
		res.f = match_contains
		return res, nil
	case "in_with_ignore_case":
		if nil == arguments || 1 >= len(arguments) {
			return nil, errors.New("wrong number of arguments (0 for 1+)")
		}
		res.Attribute = arguments[0]
		res.Arguments = toLower(arguments, 1)
		res.valueMap = make(map[string]string)
		for _, s := range res.Arguments {
			res.valueMap[s] = s
		}
		res.f = match_in_with_ignore_case
	case "equal_with_ignore_case":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = toLower(arguments, 1)
		res.f = match_equal_with_ignore_case
		return res, nil
	case "start_with_and_ignore_case":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = toLower(arguments, 1)
		res.f = match_start_with_and_ignore_case
		return res, nil
	case "end_with_and_ignore_case":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = toLower(arguments, 1)
		res.f = match_end_with_and_ignore_case
		return res, nil
	case "contains_with_ignore_case":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		res.Arguments = toLower(arguments, 1)
		res.f = match_contains_with_ignore_case
		return res, nil
	case "match":
		if e := checkArguments(arguments, 2); nil != e {
			return nil, e
		}
		res.Attribute = arguments[0]
		var e error = nil
		res.re, e = regexp.Compile(arguments[1])
		if nil != e {
			return nil, e
		}
		res.Arguments = dup(arguments, 1)
		res.f = match_regex
		return res, nil
	}
	return nil, errors.New("Unsupported method - '" + method + "'")
}

func match_in(self *Matcher, params commons.Map, value string) bool {
	_, ok := self.valueMap[value]
	return ok
}

func match_equal(self *Matcher, params commons.Map, value string) bool {
	return self.Arguments[0] == value
}

func match_start_with(self *Matcher, params commons.Map, value string) bool {
	return strings.HasPrefix(value, self.Arguments[0])
}

func match_end_with(self *Matcher, params commons.Map, value string) bool {
	return strings.HasSuffix(value, self.Arguments[0])
}

func match_contains(self *Matcher, params commons.Map, value string) bool {
	return -1 != strings.Index(value, self.Arguments[0])
}

func match_in_with_ignore_case(self *Matcher, params commons.Map, value string) bool {
	_, ok := self.valueMap[strings.ToLower(value)]
	return ok
}

func match_equal_with_ignore_case(self *Matcher, params commons.Map, value string) bool {
	return self.Arguments[0] == strings.ToLower(value)
}

func match_start_with_and_ignore_case(self *Matcher, params commons.Map, value string) bool {
	return strings.HasPrefix(strings.ToLower(value), self.Arguments[0])
}

func match_end_with_and_ignore_case(self *Matcher, params commons.Map, value string) bool {
	return strings.HasSuffix(strings.ToLower(value), self.Arguments[0])
}

func match_contains_with_ignore_case(self *Matcher, params commons.Map, value string) bool {
	return -1 != strings.Index(strings.ToLower(value), self.Arguments[0])
}

func match_regex(self *Matcher, params commons.Map, value string) bool {
	return self.re.MatchString(value)
}

type Matchers []*Matcher

func NewMatchers() Matchers {
	return make([]*Matcher, 0, 5)
}

func (self Matchers) Match(params commons.Map, debugging bool) (bool, error) {
	if nil == self || 0 == len(self) {
		return true, nil
	}

	if debugging {
		error := make([]string, 0)
		for _, m := range self {
			value := params.GetString(m.Attribute, "")
			if 0 == len(value) {
				error = append(error, commons.IsRequired(m.Attribute).Error())
			} else if !m.f(m, params, value) {
				error = append(error, "'"+m.Attribute+"' is not match - "+m.Description+"!")
			}
		}
		if 0 != len(error) {
			return false, errors.New("match failed:\n" + strings.Join(error, "\n"))
		}
	} else {
		for _, m := range self {
			value := params.GetString(m.Attribute, "")
			if 0 == len(value) {
				return false, commons.IsRequired(m.Attribute)
			}

			if !m.f(m, params, value) {
				return false, errors.New("'" + m.Attribute + "' is not match - " + m.Description + "!")
			}
		}
	}
	return true, nil
}
