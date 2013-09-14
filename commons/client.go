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

func (self *HttpClient) InvokeWithObject(action, url string, body interface{}, exceptedCode int) Result {
	if nil == body {
		return self.InvokeWith(action, url, nil, exceptedCode)
	} else {
		buffer := bytes.NewBuffer(make([]byte, 0, 1000))
		e := json.NewEncoder(buffer).Encode(body)
		if nil != e {
			return ReturnError(BadRequestCode, e.Error())
		}
		return self.InvokeWith(action, url, buffer, exceptedCode)
	}
}

func (self *HttpClient) InvokeWithBytes(action, url string, msg []byte, exceptedCode int) Result {
	if nil == msg {
		return self.InvokeWith(action, url, nil, exceptedCode)
	} else {
		return self.InvokeWith(action, url, bytes.NewBuffer(msg), exceptedCode)
	}
}

func (self *HttpClient) InvokeWith(action, url string, body io.Reader, exceptedCode int) Result {
	self.Warnings = nil

	//fmt.Println(action, url)
	req, err := http.NewRequest(action, url, body)
	if err != nil {
		return ReturnError(BadRequestCode, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "Keep-Alive")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return networkError(e.Error())
	}

	// Install closing the request body (if any)
	defer func() {
		if nil != resp.Body {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != exceptedCode {
		resp_body := readAllBytes(resp.Body)
		if nil == resp_body || 0 == len(resp_body) {
			return httpError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
		}

		return httpError(resp.StatusCode, string(resp_body))
		//return httpError(resp.StatusCode, fmt.Sprintf("[%v]%v", resp.StatusCode, string(resp_body)))
	}

	// resp_body := readAllBytes(resp.Body)
	// if nil == resp_body || 0 == len(resp_body) {
	// 	return httpError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
	// }

	//decoder := json.NewDecoder(bytes.NewBuffer(resp_body))
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	var result SimpleResult
	e = decoder.Decode(&result)
	if nil != e {
		//fmt.Println(string(resp_body))
		return unmarshalError(e)
	}
	return &result
}

type Client struct {
	*HttpClient
}

func NewClient(url, target string) *Client {
	if 0 == len(url) {
		panic("'url' is empty")
	}

	if 0 == len(target) {
		panic("'target' is empty")
	}

	return &Client{HttpClient: &HttpClient{Url: NewUrlBuilder(url).Concat(target).ToUrl()}}
}

func (self *Client) Create(params map[string]string, body interface{}) Result {
	return self.InvokeWithObject("POST", self.CreateUrl().WithQueries(params, "").ToUrl(), body, 201)
}

func (self *Client) Put(params map[string]string, body interface{}) Result {
	id := params["id"]
	if 0 == len(id) {
		return ReturnWithIsRequired("id")
	}
	delete(params, "id")

	return self.InvokeWithObject("PUT", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), body, 200)
}

func (self *Client) Delete(params map[string]string) Result {
	id := params["id"]

	if 0 == len(id) {
		return self.InvokeWith("DELETE", self.CreateUrl().WithQueries(params, "").ToUrl(), nil, 200)
	}

	delete(params, "id")
	return self.InvokeWith("DELETE", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), nil, 200)
}

func (self *Client) Get(params map[string]string) Result {
	id := params["id"]

	if 0 == len(id) {
		return self.InvokeWith("GET", self.CreateUrl().WithQueries(params, "").ToUrl(), nil, 200)
	}

	delete(params, "id")
	return self.InvokeWith("GET", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), nil, 200)
}
