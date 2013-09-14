package sampling

import (
	"commons"
)

var TableNotExists = commons.NotFound("table is not exists.")
var NotFound = commons.NotExists
var TypeError = commons.TypeError("value isn't the specific type.")

type BackgroundWorker interface {
	Stats() map[string]interface{}
	Close() // call close while server is shutdown, but not call while worker is removed
	OnTick()
}

type BackgroundWorkers interface {
	Add(id string, bw BackgroundWorker)
	Remove(id string)
}

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
	Paths       []P               `json:"route_paths,omitempty"`
}

type Filter struct {
	Method    string   `json:"method"`
	Arguments []string `json:"arguments"`
}

type P [2]string

type Sampling interface {
	Get(metric_name string, paths []P, params MContext) (interface{}, error)
	GetBool(metric_name string, paths []P, params MContext) (bool, error)
	GetInt(metric_name string, paths []P, params MContext) (int, error)
	GetInt32(metric_name string, paths []P, params MContext) (int32, error)
	GetInt64(metric_name string, paths []P, params MContext) (int64, error)
	GetUint(metric_name string, paths []P, params MContext) (uint, error)
	GetUint32(metric_name string, paths []P, params MContext) (uint32, error)
	GetUint64(metric_name string, paths []P, params MContext) (uint64, error)
	GetString(metric_name string, paths []P, params MContext) (string, error)
	GetObject(metric_name string, paths []P, params MContext) (map[string]interface{}, error)
	GetArray(metric_name string, paths []P, params MContext) ([]interface{}, error)
	GetObjects(metric_name string, paths []P, params MContext) ([]map[string]interface{}, error)
}

type MContext interface {
	Set(key string, value interface{})

	GetBool(key string) (bool, error)
	GetInt(key string) (int, error)
	GetInt32(key string) (int32, error)
	GetInt64(key string) (int64, error)
	GetUint(key string) (uint, error)
	GetUint32(key string) (uint32, error)
	GetUint64(key string) (uint64, error)
	GetFloat(key string) (float64, error)
	GetString(key string) (string, error)
	GetObject(key string) (map[string]interface{}, error)
	GetArray(key string) ([]interface{}, error)
	GetObjects(key string) ([]map[string]interface{}, error)

	GetWithDefault(key string, defaultValue interface{}) interface{}
	GetBoolWithDefault(key string, defaultValue bool) bool
	GetIntWithDefault(key string, defaultValue int) int
	GetInt32WithDefault(key string, defaultValue int32) int32
	GetInt64WithDefault(key string, defaultValue int64) int64
	GetUintWithDefault(key string, defaultValue uint) uint
	GetUint32WithDefault(key string, defaultValue uint32) uint32
	GetUint64WithDefault(key string, defaultValue uint64) uint64
	GetFloatWithDefault(key string, defaultValue float64) float64
	GetStringWithDefault(key, defaultValue string) string
	GetArrayWithDefault(key string, defaultValue []interface{}) []interface{}
	GetObjectWithDefault(key string, defaultValue map[string]interface{}) map[string]interface{}
	GetObjectsWithDefault(key string, defaultValue []map[string]interface{}) []map[string]interface{}

	CreateCtx(metric_name string, managed_type, managed_id string) (MContext, error)
	Read() Sampling
	Body() (interface{}, error)
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
	Paths       []P
	Match       Matchers
	Categories  []string
	Init        func(rs *RouteSpec, params map[string]interface{}) (Method, error)
}

var (
	Methods = map[string]*RouteSpec{}
)
