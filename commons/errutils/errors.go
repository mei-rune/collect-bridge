package errutils

import (
	"commons"
)

func InternalError(message string) commons.RuntimeError {
	return commons.NewRuntimeError(commons.InternalErrorCode, message)
}

func BadRequest(message string) commons.RuntimeError {
	return commons.NewRuntimeError(commons.BadRequestCode, message)
}

func NotAcceptable(message string) commons.RuntimeError {
	return commons.NewRuntimeError(commons.NotAcceptableCode, message)
}

func IsRequired(name string) commons.RuntimeError {
	return commons.NewRuntimeError(commons.BadRequestCode, "'"+name+"' is required.")
}

func RecordNotFound(id string) commons.RuntimeError {
	return commons.NotFound(id)
}

func RecordAlreadyExists(id string) commons.RuntimeError {
	return commons.NewRuntimeError(commons.NotAcceptableCode, "'"+id+"' is already exists.")
}
