package mdb

import (
	"commons"
	"commons/errutils"
	"encoding/json"
	"fmt"
)

type Client struct {
	*commons.HttpClient
}

func NewClient(url string) *Client {
	return &Client{HttpClient: &commons.HttpClient{Url: url}}
}

func (self *Client) Create(target string, body map[string]interface{}) (string, commons.RuntimeError) {
	msg, e := json.Marshal(body)
	if nil != e {
		return "", commons.MarshalError(e)
	}
	return self.CreateJson(self.CreateUrl().Concat(target).ToUrl(), msg)
}

func (self *Client) SaveBy(target string, params map[string]string, body map[string]interface{}) (string, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target).WithQueries(params, "@").WithQuery("save", "true").ToUrl()

	msg, e := json.Marshal(body)
	if nil != e {
		return "", commons.MarshalError(e)
	}
	return self.CreateJson(url, msg)
}

func (self *Client) CreateJson(url string, msg []byte) (string, commons.RuntimeError) {
	res, e := self.Invoke("POST", url, msg, 201)
	if nil != e {
		return "", e
	}
	result, err := res.GetReturnAsString()
	if nil != err {
		return "", errutils.InternalError(err.Error())
	}
	return result, nil
}

func (self *Client) UpdateById(target, id string, body map[string]interface{}) commons.RuntimeError {
	msg, e := json.Marshal(body)
	if nil != e {
		return commons.MarshalError(e)
	}
	c, err := self.UpdateJson(self.CreateUrl().Concat(target, id).ToUrl(), msg)
	if nil != err {
		return err
	}
	if 1 != c {
		return errutils.InternalError(fmt.Sprintln("update '%s', excepted row %d", id, c))
	}
	return nil
}

func (self *Client) UpdateBy(target string, params map[string]string, body map[string]interface{}) (int, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target, "query").WithQueries(params, "@").ToUrl()
	msg, e := json.Marshal(body)
	if nil != e {
		return -1, commons.MarshalError(e)
	}
	return self.UpdateJson(url, msg)
}

func (self *Client) UpdateJson(url string, msg []byte) (int, commons.RuntimeError) {
	res, e := self.Invoke("PUT", url, msg, 200)
	if nil != e {
		return -1, e
	}
	result, err := res.GetReturnAsInt()
	if nil != err {
		return -1, errutils.InternalError(err.Error())
	}
	return result, nil
}

func (self *Client) DeleteById(target, id string) commons.RuntimeError {
	_, e := self.Invoke("DELETE", self.CreateUrl().Concat(target, id).ToUrl(), nil, 200)
	return e
}

func (self *Client) DeleteBy(target string, params map[string]string) (int, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target, "query").WithQueries(params, "@").ToUrl()
	res, e := self.Invoke("DELETE", url, nil, 200)
	if nil != e {
		return -1, e
	}
	result, err := res.GetReturnAsInt()
	if nil != err {
		return -1, errutils.InternalError(err.Error())
	}
	return result, nil
}

func (self *Client) Count(target string, params map[string]string) (int, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target, "count").WithQueries(params, "@").ToUrl()

	res, e := self.Invoke("GET", url, nil, 200)
	if nil != e {
		return -1, e
	}

	result, err := res.GetReturnAsInt()
	if nil != err {
		return -1, errutils.InternalError(err.Error())
	}
	return result, nil
}

func (self *Client) FindById(target, id string) (map[string]interface{},
	commons.RuntimeError) {
	return self.FindByIdWithIncludes(target, id, "")
}

func (self *Client) FindBy(target string, params map[string]string) ([]map[string]interface{},
	commons.RuntimeError) {
	return self.FindByWithIncludes(target, params, "")
}

func (self *Client) FindByWithIncludes(target string, params map[string]string, includes string) (
	[]map[string]interface{}, commons.RuntimeError) {
	url := self.CreateUrl().
		Concat(target, "query").
		WithQueries(params, "@").
		WithQuery("includes", includes).ToUrl()
	res, e := self.Invoke("GET", url, nil, 200)
	if nil != e {
		return nil, e
	}
	result, err := res.GetReturnAsObjects()
	if nil != err {
		return nil, errutils.InternalError(err.Error())
	}
	return result, nil
}

func (self *Client) FindByIdWithIncludes(target, id string, includes string) (
	map[string]interface{}, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target, id).WithQuery("includes", includes).ToUrl()
	res, e := self.Invoke("GET", url, nil, 200)
	if nil != e {
		return nil, e
	}

	result, err := res.GetReturnAsObject()
	if nil != err {
		return nil, errutils.InternalError(err.Error())
	}
	return result, nil
}
