package sampling

import (
	"commons"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type MatchFunc func(matcher *Matcher, params MContext, value string) bool

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

func match_in(self *Matcher, params MContext, value string) bool {
	_, ok := self.valueMap[value]
	return ok
}

func match_equal(self *Matcher, params MContext, value string) bool {
	return self.Arguments[0] == value
}

func match_start_with(self *Matcher, params MContext, value string) bool {
	return strings.HasPrefix(value, self.Arguments[0])
}

func match_end_with(self *Matcher, params MContext, value string) bool {
	return strings.HasSuffix(value, self.Arguments[0])
}

func match_contains(self *Matcher, params MContext, value string) bool {
	return -1 != strings.Index(value, self.Arguments[0])
}

func match_in_with_ignore_case(self *Matcher, params MContext, value string) bool {
	_, ok := self.valueMap[strings.ToLower(value)]
	return ok
}

func match_equal_with_ignore_case(self *Matcher, params MContext, value string) bool {
	return self.Arguments[0] == strings.ToLower(value)
}

func match_start_with_and_ignore_case(self *Matcher, params MContext, value string) bool {
	return strings.HasPrefix(strings.ToLower(value), self.Arguments[0])
}

func match_end_with_and_ignore_case(self *Matcher, params MContext, value string) bool {
	return strings.HasSuffix(strings.ToLower(value), self.Arguments[0])
}

func match_contains_with_ignore_case(self *Matcher, params MContext, value string) bool {
	return -1 != strings.Index(strings.ToLower(value), self.Arguments[0])
}

func match_regex(self *Matcher, params MContext, value string) bool {
	return self.re.MatchString(value)
}

type Matchers []*Matcher

func NewMatchers() Matchers {
	return make([]*Matcher, 0, 5)
}

func ToFilters(matcher Matchers) []Filter {
	return nil
}

func isNotFoundOrTypeError(e error) bool {
	if re, ok := e.(commons.RuntimeError); ok {
		if commons.TypeErrorCode == re.Code() {
			return true
		}
		if commons.NotFoundCode == re.Code() {
			return true
		}
	}
	return false
}

func (self Matchers) Match(skipped int, path_params map[string]string, params MContext, debugging bool) (bool, error) {
	if nil == self || skipped >= len(self) {
		return true, nil
	}

	if debugging {
		error_messages := make([]string, 0)
		for _, m := range self[skipped:] {
			if nil != path_params {
				if s, ok := path_params[m.Attribute]; ok {
					if !m.f(m, params, s) {
						error_messages = append(error_messages, "'"+m.Attribute+"' is not match - "+m.Description+"!")
					}
					continue
				}
			}
			value, e := params.GetString("$" + m.Attribute)
			if nil != e {
				error_messages = append(error_messages, "get '"+m.Attribute+"' failed,"+e.Error())
			} else if 0 == len(value) {
				error_messages = append(error_messages, commons.IsRequired(m.Attribute).Error())
			} else if !m.f(m, params, value) {
				error_messages = append(error_messages, "'"+m.Attribute+"' is not match - "+m.Description+"!")
			}
		}
		if 0 != len(error_messages) {
			return false, errors.New("match failed:\n" + strings.Join(error_messages, "\n"))
		}
	} else {
		for _, m := range self[skipped:] {
			if nil != path_params {
				if s, ok := path_params[m.Attribute]; ok {
					if !m.f(m, params, s) {
						return false, nil //errors.New("'" + m.Attribute + "' is not match - " + m.Description + "!")
					}
					continue
				}
			}

			value, e := params.GetString("$" + m.Attribute)
			if nil != e {
				if isNotFoundOrTypeError(e) {
					//fmt.Println("'" + m.Attribute + "' is not found - " + m.Description + "!")
					return false, nil
				}
				return false, errors.New("get '" + m.Attribute + "' failed," + e.Error())
			}
			if 0 == len(value) {
				//fmt.Println("'" + m.Attribute + "' is empty - " + m.Description + "!")
				return false, nil
			}

			if !m.f(m, params, value) {
				//fmt.Println("'" + m.Attribute + "' is not match - " + m.Description + "!")
				return false, nil //errors.New("'" + m.Attribute + "' is not match - " + m.Description + "!")
			}
		}
	}
	return true, nil
}

type FilterBuilder struct {
	matchers Matchers
}

func Match() *FilterBuilder {
	return &FilterBuilder{}
}

func (self *FilterBuilder) Oid(oid string) *FilterBuilder {
	m, e := NewMatcher("start_with", []string{"sys.oid", oid})
	if nil != e {
		panic(e.Error())
	}
	self.matchers = append(self.matchers, m)
	return self
}

func (self *FilterBuilder) concat(method string, arguments []string) *FilterBuilder {
	m, e := NewMatcher(method, arguments)
	if nil != e {
		panic(e.Error())
	}
	self.matchers = append(self.matchers, m)
	return self
}

func (self *FilterBuilder) In(attributeName string, arguments []string) *FilterBuilder {
	args := make([]string, len(arguments)+1)
	args[0] = attributeName
	copy(args[1:], arguments)
	return self.concat("in", args)
}

func (self *FilterBuilder) Equals(attributeName, arguments string) *FilterBuilder {
	return self.concat("equal", []string{attributeName, arguments})
}

func (self *FilterBuilder) StartWith(attributeName, arguments string) *FilterBuilder {
	return self.concat("start_with", []string{attributeName, arguments})
}

func (self *FilterBuilder) EndWith(attributeName, arguments string) *FilterBuilder {
	return self.concat("end_with", []string{attributeName, arguments})
}

func (self *FilterBuilder) Contains(attributeName, arguments string) *FilterBuilder {
	return self.concat("contains", []string{attributeName, arguments})
}

func (self *FilterBuilder) InIgnoreCase(attributeName string, arguments []string) *FilterBuilder {
	args := make([]string, len(arguments)+1)
	args[0] = attributeName
	copy(args[1:], arguments)
	return self.concat("in_with_ignore_case", args)
}

func (self *FilterBuilder) EqualsIgnoreCase(attributeName, arguments string) *FilterBuilder {
	return self.concat("equal_with_ignore_case", []string{attributeName, arguments})
}

func (self *FilterBuilder) StartWithIgnoreCase(attributeName, arguments string) *FilterBuilder {
	return self.concat("start_with_and_ignore_case", []string{attributeName, arguments})
}

func (self *FilterBuilder) EndWithIgnoreCase(attributeName, arguments string) *FilterBuilder {
	return self.concat("end_with_and_ignore_case", []string{attributeName, arguments})
}

func (self *FilterBuilder) ContainsIgnoreCase(attributeName, arguments string) *FilterBuilder {
	return self.concat("contains_with_ignore_case", []string{attributeName, arguments})
}

func (self *FilterBuilder) Build() Matchers {
	return self.matchers
}
