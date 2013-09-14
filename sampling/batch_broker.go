package sampling

import (
	"bytes"
	"commons"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"text/template"
	"time"
)

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
	ChannelName string               `json:"channel"`
	Id          uint64               `json:"request_id"`
	CreatedAt   time.Time            `json:"created_at"`
	Error       commons.RuntimeError `json:"error,omitempty"`
	Evalue      interface{}          `json:"value,omitempty"`

	value commons.AnyData
}

func (self *ExchangeResponse) Value() commons.Any {
	self.value.Value = self.Evalue
	return &self.value
}

type RequestCtx struct {
	CreatedAt time.Time
	C         chan interface{}
	Request   *ExchangeRequest
}

type BatchBroker struct {
	name       string
	action     string
	url        string
	exchange_c chan *RequestCtx
	chan_c     chan *channelRequest

	closed     int32
	wait       sync.WaitGroup
	last_error *expvar.String

	request_id      uint64
	pendingRequests map[uint64]*RequestCtx
	channels        map[string]map[string]chan interface{}
}

type channelRequest struct {
	channelName, id string
	c               chan interface{}
	reply           chan error
}

var timeoutError = commons.NewApplicationError(http.StatusGatewayTimeout, "time out")
var closedError = errors.New("it is already closed.")
var existsError = errors.New("it is already exists.")

func (self *BatchBroker) Subscribe(channelName, id string, c chan interface{}) error {
	if 1 == atomic.LoadInt32(&self.closed) {
		return closedError
	}

	reply := make(chan error, 1)
	self.chan_c <- &channelRequest{id: id, channelName: channelName, c: c, reply: reply}
	select {
	case e := <-reply:
		return e
	case <-time.After(5 * time.Second):
		return commons.TimeoutErr
	}
}

func (self *BatchBroker) Unsubscribe(channelName, id string) {
	if 1 == atomic.LoadInt32(&self.closed) {
		return
	}

	self.chan_c <- &channelRequest{id: id, channelName: channelName}
}

func (self *BatchBroker) doChannelRequest(chanD *channelRequest) {
	var e error = nil
	grp, ok := self.channels[chanD.channelName]
	if !ok {
		if nil == chanD.c {
			goto end
		}

		grp = make(map[string]chan interface{})
		self.channels[chanD.channelName] = grp
	}

	if nil == chanD.c {
		delete(grp, chanD.id)
		if 0 == len(grp) {
			delete(self.channels, chanD.channelName)
		}
	} else {
		if _, ok := grp[chanD.id]; ok {
			e = existsError
			goto end
		}
		grp[chanD.id] = chanD.c
	}
end:
	if nil != chanD.reply {
		chanD.reply <- e
	}
}

func (self *BatchBroker) IsClosed() bool {
	return 1 == atomic.LoadInt32(&self.closed)
}

func (self *BatchBroker) Close() {
	if !atomic.CompareAndSwapInt32(&self.closed, 0, 1) {
		return
	}

	close(self.exchange_c)
	self.wait.Wait()
}

func (self *BatchBroker) run() {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			msg := buffer.String()
			self.last_error.Set(msg)
			log.Println(msg)

			if !self.IsClosed() {
				os.Exit(-1)
			}
		}

		log.Println("BatchBroker is exit.")

		atomic.StoreInt32(&self.closed, 1)
		self.wait.Done()
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	cached_objects := make([]*RequestCtx, 0, 1000)
	cached_requests := make([]*ExchangeRequest, 0, 1000)

	is_running := true
	for is_running {
		select {
		case <-ticker.C:
			self.onIdle()
		case chanD, ok := <-self.chan_c:
			if !ok {
				is_running = false
				break
			}
			self.doChannelRequest(chanD)
		case ctx, ok := <-self.exchange_c:
			if !ok {
				is_running = false
				break
			}

			self.runOnce(append(cached_objects[0:0], ctx), cached_requests[0:0], 1000)
		}

	}
}

func (self *BatchBroker) onIdle() {
	if 0 == len(self.pendingRequests) {
		return
	}

	now := time.Now()
	expired := make([]uint64, 0, 30)
	for k, ctx := range self.pendingRequests {
		interval := now.Sub(ctx.CreatedAt)
		if interval > *snmp_timeout {
			self.replyError(ctx.C, timeoutError)
			expired = append(expired, k)
		}
	}

	for _, k := range expired {
		delete(self.pendingRequests, k)
	}
}

func (self *BatchBroker) runOnce(cached_array []*RequestCtx, cached_requests []*ExchangeRequest, max_size int) {
	objects := self.recvObjects(cached_array, max_size)
	if 0 == len(objects) {
		return
	}

	for _, obj := range objects {
		self.request_id += 1
		obj.Request.Id = self.request_id
		cached_requests = append(cached_requests, obj.Request)
	}
	resposes, e := self.exchange(cached_requests)
	if nil != e {
		for _, obj := range objects {
			if nil == obj.C {
				continue
			}
			self.replyError(obj.C, e)
		}
		return
	}

	now := time.Now()
	for _, obj := range objects {
		if nil == obj.C {
			continue
		}

		obj.CreatedAt = now
		self.pendingRequests[obj.Request.Id] = obj
	}

	failed := make([]string, 0, 10)
	for _, res := range resposes {
		if pending, ok := self.pendingRequests[res.Id]; ok {
			delete(self.pendingRequests, res.Id)
			self.reply(pending.C, res)
		}

		if c_array, ok := self.channels[res.ChannelName]; ok {
			failed = failed[0:0]
			for k, c := range c_array {
				if e := self.reply(c, res); nil != e {
					failed = append(failed, k)
				}
			}
			for _, k := range failed {
				delete(c_array, k)
			}
		}
	}
}

func (self *BatchBroker) replyError(c chan<- interface{}, e commons.RuntimeError) error {
	return self.reply(c, &ExchangeResponse{CreatedAt: time.Now(), Error: e})
}

func (self *BatchBroker) reply(c chan<- interface{}, response *ExchangeResponse) (e error) {
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			e = errors.New(buffer.String())
		}
	}()
	c <- response
	return nil
}

func (self *BatchBroker) recvObjects(objects []*RequestCtx, max_size int) []*RequestCtx {
	req, ok := <-self.exchange_c
	if !ok {
		return objects
	}

	if nil == req {
		return objects
	}

	objects = append(objects, req)
	for {
		select {
		case req, ok := <-self.exchange_c:
			if !ok {
				return objects
			}
			if nil == req {
				return objects
			}
			objects = append(objects, req)
			if max_size < len(objects) {
				return objects
			}
		default:
			return objects
		}
	}
	return objects
}

func exchangeTo(method, url string, requests []*ExchangeRequest) ([]*ExchangeResponse, commons.RuntimeError) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1000))
	e := json.NewEncoder(buffer).Encode(requests)
	if nil != e {
		return nil, commons.NewApplicationError(http.StatusBadRequest, e.Error())
	}
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return nil, commons.NewApplicationError(http.StatusBadRequest, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "Keep-Alive")
	resp, e := http.DefaultClient.Do(req)
	if nil != e {
		return nil, commons.NewApplicationError(http.StatusInternalServerError, "get failed, "+e.Error())
	}

	defer func() {
		if nil != resp.Body {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusAccepted {
		if http.StatusNoContent == resp.StatusCode {
			return nil, nil
		}

		resp_body, _ := ioutil.ReadAll(resp.Body)
		if nil == resp_body || 0 == len(resp_body) {
			return nil, commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
		}

		return nil, commons.NewApplicationError(resp.StatusCode, string(resp_body))
	}

	if nil == resp.Body {
		return nil, commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
	}

	var result []*ExchangeResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	e = decoder.Decode(&result)
	if nil != e {
		return nil, commons.NewApplicationError(http.StatusInternalServerError, e.Error())
	}
	return result, nil
}

func (self *BatchBroker) exchange(requests []*ExchangeRequest) (responses []*ExchangeResponse, e commons.RuntimeError) {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			msg := buffer.String()
			e = commons.NewApplicationError(http.StatusInternalServerError, msg)
		}
	}()

	return exchangeTo(self.action, self.url, requests)
}

func newBatchClient(name, url string) (*BatchBroker, error) {
	var varString *expvar.String = nil

	varE := expvar.Get("batch_client." + name)
	if nil != varE {
		varString, _ = varE.(*expvar.String)
		if nil == varString {
			varString = expvar.NewString("foreign_db." + name + "." + time.Now().String())
		}
	} else {
		varString = expvar.NewString("foreign_db." + name)
	}

	db := &BatchBroker{name: name,
		action:          "POST",
		url:             url,
		exchange_c:      make(chan *RequestCtx, 1000),
		chan_c:          make(chan *channelRequest),
		closed:          0,
		last_error:      varString,
		request_id:      0,
		pendingRequests: make(map[uint64]*RequestCtx),
		channels:        make(map[string]map[string]chan interface{})}

	go db.run()
	db.wait.Add(1)
	return db, nil
}
