package sampling

import (
	"bytes"
	"commons"
	"encoding/json"
	"strconv"
	"text/template"
	"time"
)

func MakeChannelName(metric, managedType, managedId, pathStr string, params map[string]string) string {
	return metric + "/" + managedId + "/" + pathStr
}

type ExchangeRequest struct {
	ChannelName string            `json:"channel"`
	Id          uint64            `json:"request_id"`
	Action      string            `json:"action"`
	Name        string            `json:"metric-name"`
	ManagedType string            `json:"managed_type,omitempty"`
	ManagedId   string            `json:"managed_id,omitempty"`
	Address     string            `json:"address,omitempty"`
	Paths       []P               `json:"paths,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
}

func (self *ExchangeRequest) ToJson(w *bytes.Buffer, id uint64) error {
	old := w.Len()

	w.WriteString(`{"channel":"`)
	template.JSEscape(w, []byte(self.ChannelName))
	w.WriteString(`","request_id":`)
	w.WriteString(strconv.FormatUint(id, 10))
	w.WriteString(`,"action":"`)
	w.WriteString(self.Action)
	w.WriteString(`","metric-name":"`)
	w.WriteString(self.Name)
	if 0 != len(self.ManagedId) {
		w.WriteString(`","managed_type":"`)
		w.WriteString(self.ManagedType)
		w.WriteString(`","managed_id":"`)
		w.WriteString(self.ManagedId)
	} else {
		w.WriteString(`","address":"`)
		w.WriteString(self.Address)
	}
	w.WriteString(`",`)

	if nil != self.Paths && 0 != len(self.Paths) {
		w.WriteString(`paths":[`)
		for _, p := range self.Paths {
			w.WriteByte('"')
			w.WriteString(p[0])
			w.WriteString(`","`)
			w.WriteString(p[1])
			w.WriteString(`",`)
		}
		w.Truncate(w.Len() - 1)
		w.WriteString(`],`)
	}
	if nil != self.Params && 0 != len(self.Params) {
		w.WriteString(`params":{`)
		for k, v := range self.Params {
			w.WriteString(`"`)
			w.WriteString(k)
			w.WriteString(`":"`)
			template.JSEscape(w, []byte(v))
			w.WriteString(`",`)
		}
		w.Truncate(w.Len() - 1)
		w.WriteString(`},`)
	}

	if nil != self.Body {
		w.WriteString(`"body":`)
		if e := json.NewEncoder(w).Encode(self.Body); nil != e {
			w.Truncate(old)
			return e
		}
	} else {
		w.Truncate(w.Len() - 1)
	}

	w.WriteByte('}')
	return nil
}

type ExchangeResponse struct {
	ChannelName string                    `json:"channel"`
	Id          uint64                    `json:"request_id"`
	EcreatedAt  time.Time                 `json:"created_at"`
	Eerror      *commons.ApplicationError `json:"error,omitempty"`
	Evalue      interface{}               `json:"value,omitempty"`

	value commons.AnyData
}

func (self *ExchangeResponse) ToMap() map[string]interface{} {
	res := map[string]interface{}{}
	res["created_at"] = self.EcreatedAt
	if nil != self.Eerror {
		res["error"] = map[string]interface{}{"code": self.Eerror.Ecode, "message": self.Eerror.Emessage}
	}
	if nil != self.Evalue {
		res["value"] = self.Evalue
	}
	return res
}

func (self *ExchangeResponse) ErrorCode() int {
	if nil != self.Eerror {
		return self.Eerror.Ecode
	}
	return -1
}

func (self *ExchangeResponse) ErrorMessage() string {
	if nil != self.Eerror {
		return self.Eerror.Emessage
	}
	return ""
}

func (self *ExchangeResponse) Error() commons.RuntimeError {
	if nil == self.Eerror {
		return nil
	}
	return self.Eerror
}

func (self *ExchangeResponse) HasError() bool {
	return nil != self.Eerror && (0 != self.Eerror.Ecode || 0 != len(self.Eerror.Emessage))
}

func (self *ExchangeResponse) CreatedAt() time.Time {
	return self.EcreatedAt
}

func (self *ExchangeResponse) InterfaceValue() interface{} {
	return self.Evalue
}

func (self *ExchangeResponse) Value() commons.Any {
	self.value.Value = self.Evalue
	return &self.value
}

type requestCtx struct {
	created_at     time.Time
	cached_timeout time.Duration
	is_subscribed  bool
	c              chan interface{}
	request        *ExchangeRequest
	grp            *channelGroup
}

type ChannelClient interface {
	Id() string
	Send()
	Close()
}

type Client interface {
	Id() string
	Invoke(timeout time.Duration) (interface{}, error)
}

type clientImpl struct {
	id      string
	broker  *SamplingBroker
	request ExchangeRequest
	ctx     requestCtx
}

func (self *clientImpl) Id() string {
	return self.id
}

func (self *clientImpl) Invoke(timeout time.Duration) (interface{}, error) {
	c := make(chan interface{}, 1)
	self.broker.exchange_c <- &requestCtx{created_at: time.Now(), cached_timeout: timeout / 2, c: c,
		request: &self.request}

	select {
	case res := <-c:
		return res, nil
	case <-time.After(timeout):
		return nil, timeoutError
	}
}

func (self *clientImpl) Send() {
	self.broker.exchange_c <- &self.ctx
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
