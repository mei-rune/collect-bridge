package web

import (
	"bytes"
	"net/url"
	"os"
	"strings"
	"time"
)

func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

func dirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}

	return true
}

func fileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func Urlencode(data map[string]string) string {
	var buf bytes.Buffer
	for k, v := range data {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
		buf.WriteByte('&')
	}
	s := buf.String()
	return s[0 : len(s)-1]
}

const (
	HTTP_UNKOWN       int = -1
	HTTP_GET          int = 0
	HTTP_PUT          int = 1
	HTTP_DELETE       int = 2
	HTTP_POST         int = 3
	HTTP_HEAD         int = 4
	HTTP_OPTIONS      int = 5
	HTTP_TRACE        int = 6
	HTTP_CONNECT      int = 7
	HTTP_OTHER_METHOD int = 8
	HTTP_METHOD_END   int = 9
)

var http_methods = map[string]int{
	"GET":     HTTP_GET,
	"PUT":     HTTP_PUT,
	"DELETE":  HTTP_DELETE,
	"POST":    HTTP_POST,
	"HEAD":    HTTP_HEAD,
	"OPTIONS": HTTP_OPTIONS,
	"TRACE":   HTTP_TRACE,
	"CONNECT": HTTP_CONNECT,
}

func HashMethod(t string) int {
	ret, found := http_methods[t]
	if found {
		return ret
	}
	return HTTP_OTHER_METHOD
}
