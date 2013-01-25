package snmp

import (
	"fmt"
	"net"
	"strings"
)

func NormalizeIP(s string) string {
	if "" == s {
		return s
	}
	addr := net.ParseIP(s)
	if nil != addr {
		return addr.String()
	}

	addr = net.ParseIP(strings.Replace(s, "_", ".", -1))
	if nil != addr {
		return addr.String()
	}
	return s
}

func NormalizePort(s string) string {
	//if nil == s || 0 == len(s) {
	return s
	//}
}

func NormalizeAddress(s string) string {
	if "" == s {
		return s
	}

	idx := strings.IndexRune(s, ':')
	if -1 == idx {
		idx = strings.IndexRune(s, ',')
		if -1 == idx {
			return NormalizeIP(s)
		}
	}
	return NormalizeIP(s[0:idx]) + ":" + NormalizePort(s[idx+1:])
}

type snmpCodeException struct {
	code    SnmpResult
	message string
}

func (err *snmpCodeException) Error() string {
	return err.message
}

func (err *snmpCodeException) Code() SnmpResult {
	return err.code
}

// Errorf formats according to a format specifier and returns the string 
// as a value that satisfies error.
func Errorf(code SnmpResult, format string, a ...interface{}) SnmpCodeError {
	return &snmpCodeException{code: code, message: fmt.Sprintf(format, a...)}
}

func Error(code SnmpResult, msg string) SnmpCodeError {
	return &snmpCodeException{code: code, message: msg}
}

func newError(code SnmpResult, err error, msg string) SnmpCodeError {
	if "" == msg {
		return &snmpCodeException{code: code, message: err.Error()}
	}
	return &snmpCodeException{code: code, message: msg + " - " + err.Error()}
}
