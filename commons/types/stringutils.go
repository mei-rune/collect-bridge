package types

import (
	"github.com/grsmv/inflect"
)

func CamelCase(name string) string {
	return inflect.Camelize(name)
}

func Underscore(name string) string {
	return inflect.Underscore(name)
}

func Pluralize(str string) string {
	return inflect.Pluralize(str)
}

func Tableize(className string) string {
	return inflect.Tableize(className)
}
