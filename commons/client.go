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
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return networkError(e.Error())
	}

	// found := false
	// for _, code := range exceptedCode {
	//  if resp.StatusCode == code {
	//    found = true
	//  }
	// }
	// if !found {
	if resp.StatusCode != exceptedCode {
		return httpError(resp.StatusCode, string(readAllBytes(resp.Body)))
	}
	resp_body := readAllBytes(resp.Body)
	result := map[string]interface{}{}
	e = json.Unmarshal(resp_body, &result)
	if nil != e {
		// // Please remove it after refactor DELETE action
		// if "OK" == string(resp_body) {
		//  return commons.Return(1), nil
		// }
		//fmt.Println(string(resp_body))
		return unmarshalError(e)
	}
	if warnings, ok := result["warnings"]; ok {
		self.Warnings = warnings
	}
	return Return(result)
}
