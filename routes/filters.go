package routes

import (
	"fmt"
	"regexp"
	"strings"
)

type MatchFunc func(matcher *Matcher, params map[string]string, value string) bool

type ToStringFunc func(matcher *Matcher) string

type Matcher struct {
	Method      string
	Attribute   string
	Arguments   []string
	valueMap    map[string]string
	re          *regexp.Regexp
	f           MatchFunc
	Description string
}

func toString(method, attribute string, arguments []string) string {
	return fmt.Sprintf("%s(%s, %s)", method, attribute, strings.Join(arguments, ","))
}

func checkArguments(arguments []string, excepted int) error {
	if nil == arguments {
		return errors.New("wrong number of arguments (0 for 1)")
	}
	if excepted != len(arguments) {
		return errors.New("wrong number of arguments (" + len(arguments) + " for 1)")
	}
	return nil
}

func toLower(arguments []string) []string {
	res := make([]string, len(arguments))
	for i, v := range arguments {
		res[i] = strings.ToLower(v)
	}
	return res
}
func dup(arguments []string) []string {
	res := make([]string, len(arguments))
	copy(res, arguments)
	return res
}

func NewMatcher(method, attribute string, arguments []string) (*Matcher, error) {
	res := &Matcher{Method: method,
		Attribute:   attribute,
		Arguments:   arguments,
		Description: toString(method, attribute, arguments)}

	switch method {
	case "in":
		if nil == arguments || 0 == len(arguments) {
			return errors.New("wrong number of arguments (0 for 1+)")
		}
		res.Arguments = dup(arguments)
		res.valueMap = make(map[string]string)
		for _, s := range res.Arguments {
			res.valueMap[s] = s
		}
		res.f = match_in
		return res, nil
	case "equal":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = dup(arguments)
		res.f = match_equal
		return res, nil
	case "start_with":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = dup(arguments)
		res.f = match_start_with
		return res, nil
	case "end_with":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = dup(arguments)
		res.f = match_end_with
		return res, nil
	case "contains":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = dup(arguments)
		res.f = match_contains
		return res, nil
	case "in_with_ignore_case":
		if nil == arguments || 0 == len(arguments) {
			return errors.New("wrong number of arguments (0 for 1+)")
		}
		res.Arguments = toLower(arguments)
		res.valueMap = make(map[string]string)
		for _, s := range res.Arguments {
			res.valueMap[s] = s
		}
		res.f = match_in_with_ignore_case
	case "equal_with_ignore_case":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = toLower(arguments)
		res.f = match_equal_with_ignore_case
		return res, nil
	case "start_with_and_ignore_case":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = toLower(arguments)
		res.f = match_start_with_and_ignore_case
		return res, nil
	case "end_with_and_ignore_case":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = toLower(arguments)
		res.f = match_end_with_and_ignore_case
		return res, nil
	case "contains_with_ignore_case":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		res.Arguments = toLower(arguments)
		res.f = match_contains_with_ignore_case
		return res, nil
	case "match":
		if e := checkArguments(arguments, 1); nil != e {
			return nil, e
		}
		var e error = nil
		res.re, e = regexp.Compile(arguments[0])
		if nil != e {
			return nil, e
		}
		res.Arguments = dup(arguments)
		res.f = match_regex
		return res, nil
	}
	return nil, errors.New("Unsupported method - '" + method + "'")
}

func match_in(self *Matcher, params map[string]string, value string) bool {
	_, ok := self.valueMap[value]
	return ok
}

func match_equal(self *Matcher, params map[string]string, value string) bool {
	return self.Arguments[0] == value
}

func match_start_with(self *Matcher, params map[string]string, value string) bool {
	return strings.HasPrefix(value, self.Arguments[0])
}

func match_end_with(self *Matcher, params map[string]string, value string) bool {
	return strings.HasSuffix(value, self.Arguments[0])
}

func match_contains(self *Matcher, params map[string]string, value string) bool {
	return -1 != strings.Index(value, self.Arguments[0])
}

func match_in_with_ignore_case(self *Matcher, params map[string]string, value string) bool {
	_, ok := self.valueMap[strings.ToLower(value)]
	return ok
}

func match_equal_with_ignore_case(self *Matcher, params map[string]string, value string) bool {
	return self.Arguments[0] == strings.ToLower(value)
}

func match_start_with_and_ignore_case(self *Matcher, params map[string]string, value string) bool {
	return strings.HasPrefix(strings.ToLower(value), self.Arguments[0])
}

func match_end_with_and_ignore_case(self *Matcher, params map[string]string, value string) bool {
	return strings.HasSuffix(strings.ToLower(value), self.Arguments[0])
}

func match_contains_with_ignore_case(self *Matcher, params map[string]string, value string) bool {
	return -1 != strings.Index(strings.ToLower(value), self.Arguments[0])
}

func match_regex(self *Matcher, params map[string]string, value string) bool {
	return self.re.MatchString(value)
}

type Matchers []*Matcher

func NewMatchers() Matchers {
	return make([]*Matcher, 0, 5)
}

func (self Matchers) Match(params map[string]string) (bool, error) {
	if nil == self || 0 == len(self) {
		return true, nil
	}

	for _, m := range self {
		value, ok := params[m.Attribute]
		if !ok {
			return false, errors.New("'" + m.Attribute + "' is not exists!")
		}

		if !m.f(m, params, value) {
			return false, errors.New("'" + m.Attribute + "' is not match - " + m.Description + "!")
		}
	}
	return true, nil
}
