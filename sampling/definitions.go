package sampling

import (
	"commons"
)

type RouteDefinition struct {
	Method      string            `json:"method"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Level       []string          `json:"level"`
	File        string            `json:"file"`
	Action      map[string]string `json:"action"`
	Match       []Filter          `json:"match"`
	Categories  []string          `json:"categories"`
}

type Filter struct {
	Method    string   `json:"method"`
	Arguments []string `json:"arguments"`
}

type MContext interface {
	commons.Map
	Body() interface{}
}

type Method interface {
	Call(params MContext) commons.Result
}

type RouteSpec struct {
	Method      string
	Name        string
	Description string
	Author      string
	License     string
	Level       []string
	File        string
	Match       Matchers
	Categories  []string
	Call        func(rs *RouteSpec, params map[string]interface{}) (Method, error)
}

var (
	Methods = map[string]*RouteSpec{}
)
