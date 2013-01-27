package errutils

import (
	"commons"
)

func InternalError(message string) commons.RuntimeError {
	return commons.NewRuntimeError(500, message)
}

func BadRequest(message string) commons.RuntimeError {
	return commons.NewRuntimeError(400, message)
}

func NotAcceptable(message string) commons.RuntimeError {
	return commons.NewRuntimeError(406, message)
}

func IsRequired(name string) commons.RuntimeError {
	return commons.NewRuntimeError(400, "'"+name+"' is required.")
}

func RecordNotFound(id string) commons.RuntimeError {
	return commons.NewRuntimeError(404, "'"+id+"' is not found.")
}

func RecordAlreadyExists(id string) commons.RuntimeError {
	return commons.NewRuntimeError(406, "'"+id+"' is already exists.")
}
