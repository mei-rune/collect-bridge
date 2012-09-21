package snmp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	time_format = "time format error with valuw is '%s', excepted format is 'xxx[unit]', xxx is a number, unit must is in (ms, s, m)."
)

func ParseTime(s string) (time.Duration, error) {
	idx := strings.IndexFunc(s, func(r rune) bool {
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return false
		}
		return true
	})

	if idx == 0 {
		return 0, fmt.Errorf(time_format, s)
	}

	unit := time.Second
	if -1 != idx {
		switch s[idx:] {
		case "ms", "MS":
			unit = time.Millisecond
		case "s", "S":
			unit = time.Second
		case "m", "M":
			unit = time.Minute
		default:
			return 0, fmt.Errorf(time_format, s)
		}
		s = s[:idx]
	}

	i, err := strconv.ParseInt(s, 10, 0)
	if nil != err {
		return 0, fmt.Errorf(time_format, s, err.Error())
	}
	return time.Duration(i) * unit, nil
}

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
