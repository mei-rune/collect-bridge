package mdb

import (
	"bytes"
	"commons"
	"commons/as"
	"commons/errutils"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func marshalError(e error) commons.RuntimeError {
	return errutils.BadRequest("marshal failed, " + e.Error())
}

func unmarshalError(e error) commons.RuntimeError {
	return errutils.InternalError("unmarshal failed, " + e.Error())
}

func networkError(fmts string, params ...interface{}) commons.RuntimeError {
	if 0 == len(params) {
		return commons.NewRuntimeError(commons.NetworkErrorCode, fmts)
	}
	return commons.NewRuntimeError(commons.NetworkErrorCode, fmt.Sprintf(fmts, params))
}

func httpError(code int, fmts string, params ...interface{}) commons.RuntimeError {
	if 0 == len(params) {
		return commons.NewRuntimeError(code, fmts)
	}
	return commons.NewRuntimeError(code, fmt.Sprintf(fmts, params))
}

func typeError(key, t string) commons.RuntimeError {
	return errutils.InternalError("'" + key + "' is not a " + t + ".")
}

type Client struct {
	url      string
	Warnings interface{}
}

func NewClient(url string) *Client {
	return &Client{url: url}
}

func readAllBytes(r io.Reader) []byte {
	bs, e := ioutil.ReadAll(r)
	if nil != e {
		panic(e.Error())
	}
	return bs
}

func (self *Client) invoke(action, url string, msg []byte, exceptedCode int) (map[string]interface{}, commons.RuntimeError) {
	self.Warnings = nil

	req, err := http.NewRequest(action, url, bytes.NewBuffer(msg))
	if err != nil {
		return nil, errutils.BadRequest(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return nil, networkError(e.Error())
	}

	// found := false
	// for _, code := range exceptedCode {
	//	if resp.StatusCode == code {
	//		found = true
	//	}
	// }
	// if !found {
	if resp.StatusCode != exceptedCode {
		return nil, httpError(resp.StatusCode, string(readAllBytes(resp.Body)))
	}
	resp_body := readAllBytes(resp.Body)
	result := map[string]interface{}{}
	e = json.Unmarshal(resp_body, &result)
	if nil != e {
		// // Please remove it after refactor DELETE action
		// if "OK" == string(resp_body) {
		//	return commons.Return(1), nil
		// }
		//fmt.Println(string(resp_body))
		return nil, unmarshalError(e)
	}
	if warnings, ok := result["warnings"]; ok {
		self.Warnings = warnings
	}
	return result, nil
}

func (self *Client) Create(target string, body map[string]interface{}) (string, commons.RuntimeError) {
	msg, e := json.Marshal(body)
	if nil != e {
		return "", marshalError(e)
	}
	return self.CreateJson(self.url+target, msg)
}

func (self *Client) SaveBy(target string, params map[string]string, body map[string]interface{}) (string, commons.RuntimeError) {
	url := self.url + target + "/query?"
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	url += "save=true"

	msg, e := json.Marshal(body)
	if nil != e {
		return "", marshalError(e)
	}
	return self.CreateJson(url, msg)
}

func (self *Client) CreateJson(url string, msg []byte) (string, commons.RuntimeError) {
	res, e := self.invoke("POST", url, msg, 201)
	if nil != e {
		return "", e
	}
	v := commons.GetReturn(res)
	if nil == v {
		return "", commons.ValueIsNil
	}

	if id, ok := v.(string); ok {
		return id, nil
	}
	return "", typeError("id", "string")
}

func (self *Client) UpdateById(target, id string, body map[string]interface{}) commons.RuntimeError {
	msg, e := json.Marshal(body)
	if nil != e {
		return marshalError(e)
	}
	c, err := self.UpdateJson(self.url+target+"/"+id, msg)
	if nil != err {
		return err
	}
	if 1 != c {
		return errutils.InternalError(fmt.Sprintln("update '%s', excepted row %d", id, c))
	}
	return nil
}

func (self *Client) UpdateBy(target string, params map[string]string, body map[string]interface{}) (int, commons.RuntimeError) {
	url := self.url + target + "/query?"
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	msg, e := json.Marshal(body)
	if nil != e {
		return -1, marshalError(e)
	}
	return self.UpdateJson(url[:len(url)-1], msg)
}

func (self *Client) UpdateJson(url string, msg []byte) (int, commons.RuntimeError) {
	res, e := self.invoke("PUT", url, msg, 200)
	if nil != e {
		return -1, e
	}
	v := commons.GetReturn(res)
	if nil == v {
		return -1, commons.ValueIsNil
	}

	//effected
	if c, e := as.AsInt(v); nil == e {
		return c, nil
	}
	return -1, typeError("effected", "string")
}

func (self *Client) DeleteById(target, id string) commons.RuntimeError {
	_, e := self.invoke("DELETE", self.url+target+"/"+id, nil, 200)
	return e
}

func (self *Client) DeleteBy(target string, params map[string]string) (int, commons.RuntimeError) {
	url := self.url + target + "/query?"
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	res, e := self.invoke("DELETE", url[:len(url)-1], nil, 200)
	if nil != e {
		return -1, e
	}

	v := commons.GetReturn(res)
	if nil == v {
		return -1, commons.ValueIsNil
	}

	//effected
	if c, e := as.AsInt(v); nil == e {
		return c, nil
	}
	return -1, typeError("effected", "string")
}

func (self *Client) Count(t string, params map[string]string) (int, commons.RuntimeError) {
	url := self.url + t + "/count?"
	for k, v := range params {
		url += ("@" + k + "=" + v + "&")
	}
	res, e := self.invoke("GET", url, nil, 200)
	if nil != e {
		return -1, e
	}
	v := commons.GetReturn(res)
	if nil == v {
		return -1, commons.ValueIsNil
	}

	//effected
	if c, e := as.AsInt(v); nil == e {
		return c, nil
	}
	return -1, typeError("count", "string")
}

func (self *Client) FindById(target, id string) (map[string]interface{},
	commons.RuntimeError) {
	return self.FindByIdWithIncludes(target, id, "")
}

func (self *Client) FindByIdWithIncludes(target, id string, includes string) (
	map[string]interface{}, commons.RuntimeError) {
	url := self.url + target + "/" + id
	if "" != includes {
		url += ("?includes=" + includes)
	}

	res, e := self.invoke("GET", url, nil, 200)
	if nil != e {
		return nil, e
	}
	v := commons.GetReturn(res)
	if nil == v {
		return nil, commons.ValueIsNil
	}

	if result, ok := v.(map[string]interface{}); ok {
		return result, nil
	}

	return nil, typeError("result", "map[string]interface{}")
}
