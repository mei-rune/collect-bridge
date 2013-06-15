package ds

import (
	"commons"
	"encoding/json"
	"fmt"
)

type Client struct {
	*commons.HttpClient
}

func NewClient(url string) *Client {
	return &Client{HttpClient: &commons.HttpClient{Url: url}}
}

func marshalError(e error) commons.RuntimeError {
	return commons.NewRuntimeError(commons.BadRequestCode, "marshal failed, "+e.Error())
}

func (self *Client) Create(target string, body map[string]interface{}) (string, commons.RuntimeError) {
	msg, e := json.Marshal(body)
	if nil != e {
		return "", marshalError(e)
	}
	return self.CreateJson(self.CreateUrl().Concat(target).ToUrl(), msg)
}

func (self *Client) SaveBy(target string, params map[string]string, body map[string]interface{}) (string, commons.RuntimeError) {
	url := self.CreateUrl().
		Concat(target).
		WithQueries(params, "@").
		WithQuery("save", "true").
		ToUrl()

	msg, e := json.Marshal(body)
	if nil != e {
		return "", marshalError(e)
	}
	return self.CreateJson(url, msg)
}

func (self *Client) CreateJson(url string, msg []byte) (string, commons.RuntimeError) {
	res := self.Invoke("POST", url, msg, 201)
	if res.HasError() {
		return "", res.Error()
	}

	if nil == res.LastInsertId() {
		return "", commons.InternalError("lastInsertId is nil")
	}

	result := fmt.Sprint(res.LastInsertId())
	if "-1" == res.LastInsertId() {
		return "", commons.InternalError("lastInsertId is -1")
	}

	return result, nil
}

func (self *Client) UpdateById(target, id string, body map[string]interface{}) commons.RuntimeError {
	msg, e := json.Marshal(body)
	if nil != e {
		return marshalError(e)
	}
	c, err := self.UpdateJson(self.CreateUrl().Concat(target, id).ToUrl(), msg)
	if nil != err {
		return err
	}
	if 1 != c {
		return commons.InternalError(fmt.Sprintln("update row with id is '%s', effected row is %d", id, c))
	}
	return nil
}

func (self *Client) UpdateBy(target string, params map[string]string, body map[string]interface{}) (int64, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target).WithQueries(params, "@").ToUrl()
	msg, e := json.Marshal(body)
	if nil != e {
		return -1, marshalError(e)
	}
	return self.UpdateJson(url, msg)
}

func (self *Client) UpdateJson(url string, msg []byte) (int64, commons.RuntimeError) {
	res := self.Invoke("PUT", url, msg, 200)
	if res.HasError() {
		return -1, res.Error()
	}
	result := res.Effected()
	if -1 == result {
		return -1, commons.InternalError("effected rows is -1")
	}
	return result, nil
}

func (self *Client) DeleteById(target, id string) commons.RuntimeError {
	res := self.Invoke("DELETE", self.CreateUrl().Concat(target, id).ToUrl(), nil, 200)
	return res.Error()
}

func (self *Client) DeleteBy(target string, params map[string]string) (int64, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target).WithQueries(params, "@").ToUrl()
	res := self.Invoke("DELETE", url, nil, 200)
	if res.HasError() {
		return -1, res.Error()
	}
	result := res.Effected()
	if -1 == result {
		return -1, commons.InternalError("effected rows is -1")
	}
	return result, nil
}

func (self *Client) Count(target string, params map[string]string) (int64, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target, "@count").WithQueries(params, "@").ToUrl()
	res := self.Invoke("GET", url, nil, 200)
	if res.HasError() {
		return -1, res.Error()
	}

	result, err := res.Value().AsInt64()
	if nil != err {
		return -1, commons.InternalError(err.Error())
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
		Concat(target).
		WithQueries(params, "@").
		WithQuery("includes", includes).ToUrl()
	res := self.Invoke("GET", url, nil, 200)
	if res.HasError() {
		return nil, res.Error()
	}
	result, err := res.Value().AsObjects()
	if nil != err {
		return nil, commons.InternalError(err.Error())
	}
	return result, nil
}

func (self *Client) FindByIdWithIncludes(target, id string, includes string) (
	map[string]interface{}, commons.RuntimeError) {
	url := self.CreateUrl().Concat(target, id).WithQuery("includes", includes).ToUrl()
	res := self.Invoke("GET", url, nil, 200)
	if res.HasError() {
		return nil, res.Error()
	}
	result, err := res.Value().AsObject()
	if nil != err {
		return nil, commons.InternalError(err.Error())
	}
	//fmt.Printf("res = %#v\r\n\r\n", res)
	//fmt.Printf("value = %#v\r\n\r\n", res.InterfaceValue())
	return result, nil
}

func (self *Client) Children(parent, parent_id, target string, params map[string]string) ([]map[string]interface{},
	commons.RuntimeError) {
	url := self.CreateUrl().
		Concat(parent, parent_id, "children", target).
		WithQueries(params, "@").ToUrl()
	res := self.Invoke("GET", url, nil, 200)
	if res.HasError() {
		return nil, res.Error()
	}
	result, err := res.Value().AsObjects()
	if nil != err {
		return nil, commons.InternalError(err.Error())
	}
	return result, nil
}

func (self *Client) Parent(child, child_id, target string) (map[string]interface{},
	commons.RuntimeError) {
	url := self.CreateUrl().
		Concat(child, child_id, "parent", target).ToUrl()
	res := self.Invoke("GET", url, nil, 200)
	if res.HasError() {
		return nil, res.Error()
	}
	result, err := res.Value().AsObject()
	if nil != err {
		return nil, commons.InternalError(err.Error())
	}
	return result, nil
}
