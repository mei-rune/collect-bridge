package commons

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type HttpClient struct {
	Url      string
	Warnings interface{}
}

func MarshalError(e error) Result {
	return ReturnError(BadRequestCode, "marshal failed, "+e.Error())
}

func unmarshalError(e error) Result {
	return ReturnError(InternalErrorCode, "unmarshal failed, "+e.Error())
}

func networkError(fmts string, params ...interface{}) Result {
	if 0 == len(params) {
		return ReturnError(NetworkErrorCode, fmts)
	}
	return ReturnError(NetworkErrorCode, fmt.Sprintf(fmts, params))
}

func httpError(code int, fmts string, params ...interface{}) Result {
	if 0 == len(params) {
		return ReturnError(code, fmts)
	}
	return ReturnError(code, fmt.Sprintf(fmts, params))
}

func readAllBytes(r io.Reader) []byte {
	bs, e := ioutil.ReadAll(r)
	if nil != e {
		panic(e.Error())
	}
	return bs
}

func (self *HttpClient) CreateUrl() *UrlBuilder {
	return NewUrlBuilder(self.Url)
}

func (self *HttpClient) Invoke(action, url string, msg []byte, exceptedCode int) Result {
	self.Warnings = nil

	req, err := http.NewRequest(action, url, bytes.NewBuffer(msg))
	if err != nil {
		return ReturnError(BadRequestCode, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return networkError(e.Error())
	}

	resp_body := readAllBytes(resp.Body)
	if resp.StatusCode != exceptedCode {
		if nil == resp_body || 0 == len(resp_body) {
			return httpError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
		}
		return httpError(resp.StatusCode, string(resp_body))
	}
	var result SimpleResult
	e = json.Unmarshal(resp_body, &result)
	if nil != e {
		return unmarshalError(e)
	}
	return &result
}
