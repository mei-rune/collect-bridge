package commons

import (
	"bytes"
	"fmt"
	"strings"
)

//url := self.CreateURL().Concat(target, "query").WithQueries(params, "@").WithQuery("save", "true").ToUrl()

type UrlBuilder struct {
	bytes.Buffer
	is_query bool
}

func NewUrlBuilder(base string) *UrlBuilder {
	builder := &UrlBuilder{is_query: false}
	if '/' == base[len(base)-1] {
		builder.WriteString(base[:len(base)-1])
	} else {
		builder.WriteString(base)
	}
	if strings.ContainsRune(base, '?') {
		builder.is_query = true
	}
	return builder
}

func (self *UrlBuilder) Concat(paths ...string) *UrlBuilder {
	if self.is_query {
		panic("[panic] don`t append path to the query")
	}

	for _, pa := range paths {
		if 0 == len(pa) {
			continue
		}

		if '/' != pa[0] {
			self.WriteString("/")
		}

		if '/' == pa[len(pa)-1] {
			self.WriteString(pa[:len(pa)-1])
		} else {
			self.WriteString(pa)
		}
	}
	return self
}

func (self *UrlBuilder) WithQuery(key, value string) *UrlBuilder {
	if 0 == len(key) {
		return self
	}
	if !self.is_query {
		self.WriteString("?")
		self.is_query = true
	} else {
		self.WriteString("&")
	}
	self.WriteString(key)
	self.WriteString("=")
	self.WriteString(value)
	return self
}

func (self *UrlBuilder) WithQueries(params map[string]string, prefix string) *UrlBuilder {
	if 0 == len(params) {
		return self
	}
	if !self.is_query {
		self.WriteString("?")
		self.is_query = true
	} else {
		self.WriteString("&")
	}
	for k, v := range params {
		self.WriteString(prefix)
		self.WriteString(k)
		self.WriteString("=")
		self.WriteString(v)
		self.WriteString("&")
	}
	self.Truncate(self.Len() - 1)
	return self
}

func (self *UrlBuilder) WithAnyQueries(params map[string]interface{}, prefix string) *UrlBuilder {
	if 0 == len(params) {
		return self
	}
	if !self.is_query {
		self.WriteString("?")
		self.is_query = true
	} else {
		self.WriteString("&")
	}
	for k, v := range params {
		self.WriteString(prefix)
		self.WriteString(k)
		self.WriteString("=")
		if s, ok := v.(string); ok {
			self.WriteString(s)
		} else {
			self.WriteString(fmt.Sprint(v))
		}
		self.WriteString("&")
	}
	self.Truncate(self.Len() - 1)
	return self
}

func (self *UrlBuilder) ToUrl() string {
	return self.String()
}
