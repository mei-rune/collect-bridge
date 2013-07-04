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

	//fmt.Println(action, url)
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

	var result SimpleResult
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	e = decoder.Decode(&result)
	if nil != e {
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

func (self *Client) Create(params map[string]string) Result {
	body := params["body"]

	if 0 == len(body) {
		return ReturnWithIsRequired("body")
	}

	delete(params, "body")
	return self.Invoke("POST", self.CreateUrl().WithQueries(params, "").ToUrl(), []byte(body), 201)
}

func (self *Client) Put(params map[string]string) Result {
	id := params["id"]
	body := params["body"]

	if 0 == len(id) {
		return ReturnWithIsRequired("id")
	}

	if 0 == len(body) {
		return ReturnWithIsRequired("body")
	}

	delete(params, "id")
	delete(params, "body")
	return self.Invoke("PUT", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), []byte(body), 200)
}

func (self *Client) Delete(params map[string]string) Result {
	id := params["id"]

	if 0 == len(id) {
		return ReturnWithIsRequired("id")
	}

	delete(params, "id")
	return self.Invoke("DELETE", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), []byte(""), 200)
}

func (self *Client) Get(params map[string]string) Result {
	id := params["id"]

	if 0 == len(id) {
		return ReturnWithIsRequired("id")
	}

	delete(params, "id")
	return self.Invoke("GET", self.CreateUrl().Concat(id).WithQueries(params, "").ToUrl(), []byte(""), 200)
}
