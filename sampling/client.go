package sampling

import (
	// "commons"
	// "encoding/json"
	// "errors"
	// "fmt"
	"time"
)

type ChannelClient interface {
	Send()
	Close()
}

type Client interface {
	Invoke(timeout time.Duration) (interface{}, error)
}

type clientImpl struct {
	id      string
	broker  *SamplingBroker
	c       chan *RequestCtx
	request ExchangeRequest
	ctx     RequestCtx
}

func (self *clientImpl) Invoke(timeout time.Duration) (interface{}, error) {
	c := make(chan interface{}, 1)
	self.c <- &RequestCtx{CreatedAt: time.Now(), C: c,
		Request: &self.request}

	select {
	case res := <-c:
		return res, nil
	case <-time.After(timeout):
		return nil, timeoutError
	}
}

func (self *clientImpl) Send() {
	self.c <- &self.ctx
}

func (self *clientImpl) Close() {
	self.broker.Unsubscribe(self.request.ChannelName, self.id)
}

// func NewClient(url string) *Client {
// 	return &Client{HttpClient: &commons.HttpClient{Url: url}}
// }

// func marshalError(e error) commons.RuntimeError {
// 	return commons.NewRuntimeError(commons.BadRequestCode, "marshal failed, "+e.Error())
// }

// func (self *Client) Create(target string, body map[string]interface{}) (string, commons.RuntimeError) {
// 	msg, e := json.Marshal(body)
// 	if nil != e {
// 		return "", marshalError(e)
// 	}
// 	_, id, err := self.CreateJson(self.CreateUrl().Concat(target).ToUrl(), msg)
// 	return id, err
// }

// func (self *Client) CreateJson(url string, msg []byte) commons.Result {
// 	res := self.Invoke("POST", url, msg, 201)
// 	if res.HasError() {
// 		return res
// 	}

// 	if nil == res.LastInsertId() {
// 		return commons.ReturnWithInternalError("lastInsertId is nil")
// 	}

// 	result := fmt.Sprint(res.LastInsertId())
// 	if "-1" == res.LastInsertId() {
// 		return commons.ReturnWithInternalError("lastInsertId is -1")
// 	}

// 	return res
// }

// func (self *Client) Get(managed_type, managed_id, target string) commons.Result {
// 	[]map[string]interface{}, commons.RuntimeError) {
// 	url := self.CreateUrl().
// 		Concat(managed_type, managed_id, target)

// 	return self.Invoke("GET", url.ToUrl(), nil, 200)
// }
