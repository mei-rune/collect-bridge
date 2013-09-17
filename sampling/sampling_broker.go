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
	"strings"
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

type RequestCtx struct {
	CreatedAt time.Time
	C         chan interface{}
	Request   *ExchangeRequest
	grp       *channelGroup
}

type SamplingBroker struct {
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
	channelGroups   map[string]*channelGroup
}

type channelGroup struct {
	last_begin_at time.Time
	last_end_at   time.Time
	channels      map[string]chan interface{}
}

func (self *SamplingBroker) createClient(channelName string, c chan interface{},
	method, metric_name, managedType, managedId, pathStr string, params map[string]string, body interface{}) (*clientImpl, error) {
	pathStr = strings.Trim(pathStr, "/")
	var paths []P
	if 0 != len(pathStr) {
		ss := strings.Split(pathStr, "/")
		if 0 != len(ss)%2 {
			return nil, errors.New("paths is style error - `" + pathStr + "`")
		}
		paths = make([]P, 0, len(ss)/2)
		for i := 0; i < len(ss); i++ {
			paths = append(paths, P{ss[0], ss[1]})
		}
	}

	cl := &clientImpl{id: commons.GenerateId(),
		broker: self,
		c:      self.exchange_c,
		request: ExchangeRequest{ChannelName: channelName,
			Action:      method,
			Name:        metric_name,
			ManagedType: managedType,
			ManagedId:   managedId,
			Paths:       paths,
			Params:      params,
			Body:        body}}
	cl.ctx.Request = &cl.request

	if nil != c {
		grp, e := self.Subscribe(channelName, cl.id, c)
		if nil != e {
			return nil, e
		}
		cl.ctx.grp = grp
		cl.ctx.C = c
	}
	return cl, nil
}

func (self *SamplingBroker) CreateClient(channelName, method, metric_name, managedType, managedId, pathStr string,
	params map[string]string, body interface{}) (Client, error) {
	cl, e := self.createClient(channelName, nil, method, metric_name, managedType, managedId, pathStr, params, body)
	if nil != e {
		return nil, e
	}
	return cl, nil
}
func (self *SamplingBroker) SubscribeClient(channelName string, c chan interface{},
	method, metric_name, managedType, managedId, pathStr string, params map[string]string, body interface{}) (ChannelClient, error) {
	cl, e := self.createClient(channelName, c, method, metric_name, managedType, managedId, pathStr, params, body)
	if nil != e {
		return nil, e
	}
	return cl, nil
}

type channelRequest struct {
	channelName, id string
	c               chan interface{}
	reply           chan *channelRequest
	e               error
	grp             *channelGroup
}

var timeoutError = commons.NewApplicationError(http.StatusGatewayTimeout, "time out")
var closedError = errors.New("it is already closed.")
var existsError = errors.New("it is already exists.")

func (self *SamplingBroker) Subscribe(channelName, id string, c chan interface{}) (*channelGroup, error) {
	if 1 == atomic.LoadInt32(&self.closed) {
		return nil, closedError
	}

	reply := make(chan *channelRequest, 1)
	self.chan_c <- &channelRequest{id: id, channelName: channelName, c: c, reply: reply}
	select {
	case r := <-reply:
		return r.grp, r.e
	case <-time.After(5 * time.Second):
		return nil, commons.TimeoutErr
	}
}

func (self *SamplingBroker) Unsubscribe(channelName, id string) {
	if 1 == atomic.LoadInt32(&self.closed) {
		return
	}

	self.chan_c <- &channelRequest{id: id, channelName: channelName}
}

func (self *SamplingBroker) doChannelRequest(chanD *channelRequest) {
	grp, ok := self.channelGroups[chanD.channelName]
	if !ok {
		if nil == chanD.c {
			goto end
		}

		grp = &channelGroup{channels: make(map[string]chan interface{})}
		self.channelGroups[chanD.channelName] = grp
	}

	if nil == chanD.c {
		delete(grp.channels, chanD.id)
		if 0 == len(grp.channels) {
			delete(self.channelGroups, chanD.channelName)
		}
	} else {
		if _, ok := grp.channels[chanD.id]; ok {
			chanD.e = existsError
			goto end
		}

		chanD.grp = grp
		grp.channels[chanD.id] = chanD.c
	}
end:
	if nil != chanD.reply {
		chanD.reply <- chanD
	}
}

func (self *SamplingBroker) IsClosed() bool {
	return 1 == atomic.LoadInt32(&self.closed)
}

func (self *SamplingBroker) Close() {
	if !atomic.CompareAndSwapInt32(&self.closed, 0, 1) {
		return
	}

	close(self.exchange_c)
	self.wait.Wait()
}

func (self *SamplingBroker) run() {
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

		log.Println("SamplingBroker is exit.")

		atomic.StoreInt32(&self.closed, 1)
		self.wait.Done()
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	cached_objects := make([]*RequestCtx, 0, 1000)
	cached_requests := make([]*ExchangeRequest, 0, 1000)

	count := 0
	is_running := true
	for is_running {
		select {
		case <-ticker.C:
			if 0 == count%10 {
				self.onIdle()
			}
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
			if nil == ctx {
				self.runOnce(cached_objects[0:0], cached_requests[0:0], 1000)
			} else {
				self.runOnce(append(cached_objects[0:0], ctx), cached_requests[0:0], 1000)
			}
		}
	}
}

func (self *SamplingBroker) recvObjects(objects []*RequestCtx, max_size int) []*RequestCtx {
	for {
		select {
		case req, ok := <-self.exchange_c:
			if !ok {
				return objects
			}
			if nil != req {
				objects = append(objects, req)
				if max_size < len(objects) {
					return objects
				}
			}
		default:
			return objects
		}
	}
	return objects
}

func (self *SamplingBroker) onIdle() {
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

var empty_id_list = []uint64{}

func (self *SamplingBroker) runOnce(cached_array []*RequestCtx, cached_requests []*ExchangeRequest, max_size int) {
	objects := self.recvObjects(cached_array, max_size)

	id_list := empty_id_list
	if 0 != len(objects) {
		id_list = make([]uint64, len(objects))
		now := time.Now()
		for idx, obj := range objects {
			if nil != obj.grp {
				if now.Sub(obj.grp.last_begin_at).Seconds() < 15 {
					continue
				}

				if now.Sub(obj.grp.last_end_at).Seconds() < 15 {
					continue
				}

				obj.grp.last_begin_at = now
			}
			self.request_id += 1
			id_list[idx] = self.request_id
			cached_requests = append(cached_requests, obj.Request)
		}
	}
	resposes, e := self.exchange(id_list, cached_requests)
	if nil != e {
		//fmt.Println("----", e, len(objects))
		for _, obj := range objects {
			if nil == obj.C {
				//fmt.Println("skip")
				continue
			}
			self.replyError(obj.C, e)
		}
		return
	}
	//fmt.Println(len(resposes))

	now := time.Now()
	for idx, obj := range objects {
		if nil == obj.C {
			continue
		}

		obj.CreatedAt = now
		self.pendingRequests[id_list[idx]] = obj
	}

	grp_failed := make([]string, 0, 10)
	failed := make([]string, 0, 10)
	for _, res := range resposes {
		if pending, ok := self.pendingRequests[res.Id]; ok {
			delete(self.pendingRequests, res.Id)
			self.reply(pending.C, res)
		}

		if grp, ok := self.channelGroups[res.ChannelName]; ok {
			grp.last_end_at = now

			failed = failed[0:0]
			for k, c := range grp.channels {
				if e := self.reply(c, res); nil != e {
					failed = append(failed, k)
				}
			}
			for _, k := range failed {
				delete(grp.channels, k)
			}

			if 0 == len(grp.channels) {
				grp_failed = append(grp_failed, res.ChannelName)
			}
		}
	}

	if 0 != len(grp_failed) {
		for _, k := range grp_failed {
			delete(self.channelGroups, k)
		}
	}
}

func (self *SamplingBroker) replyError(c chan<- interface{}, e commons.RuntimeError) error {
	return self.reply(c, &ExchangeResponse{EcreatedAt: time.Now(), Eerror: &commons.ApplicationError{Ecode: e.Code(), Emessage: e.Error()}})
}

func (self *SamplingBroker) reply(c chan<- interface{}, response *ExchangeResponse) (e error) {
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

func exchangeTo(method, url string, id_list []uint64, requests []*ExchangeRequest) ([]*ExchangeResponse, commons.RuntimeError) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1000))
	buffer.WriteByte('[')
	if nil != requests && 0 != len(requests) {
		for idx, r := range requests {
			if e := r.ToJson(buffer, id_list[idx]); nil != e {
				return nil, commons.NewApplicationError(http.StatusBadRequest, e.Error())
			}
			buffer.WriteByte(',')
		}
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteByte(']')
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

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		if http.StatusNoContent == resp.StatusCode {
			return nil, nil
		}

		resp_body, _ := ioutil.ReadAll(resp.Body)
		if nil == resp_body || 0 == len(resp_body) {
			return nil, commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: error", resp.StatusCode))
		}

		return nil, commons.NewApplicationError(resp.StatusCode, fmt.Sprintf("%v: %v", resp.StatusCode, string(resp_body)))
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

func (self *SamplingBroker) exchange(id_list []uint64, requests []*ExchangeRequest) (responses []*ExchangeResponse, e commons.RuntimeError) {
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

	return exchangeTo(self.action, self.url, id_list, requests)
}

func NewBroker(name, url string) (*SamplingBroker, error) {
	var varString *expvar.String = nil

	varE := expvar.Get("sampling_broker." + name)
	if nil != varE {
		varString, _ = varE.(*expvar.String)
		if nil == varString {
			varString = expvar.NewString("sampling_broker." + name + "." + time.Now().String())
		}
	} else {
		varString = expvar.NewString("sampling_broker." + name)
	}

	db := &SamplingBroker{name: name,
		action:          "POST",
		url:             url,
		exchange_c:      make(chan *RequestCtx, 1000),
		chan_c:          make(chan *channelRequest),
		closed:          0,
		last_error:      varString,
		request_id:      0,
		pendingRequests: make(map[uint64]*RequestCtx),
		channelGroups:   make(map[string]*channelGroup)}

	if nil != expvar.Get("sampling_broker."+name) {
		expvar.Publish("sampling_broker."+name+"."+time.Now().String()+".queue", expvar.Func(func() interface{} {
			return len(db.exchange_c)
		}))
	} else {
		expvar.Publish("sampling_broker."+name+".queue", expvar.Func(func() interface{} {
			return len(db.exchange_c)
		}))
	}
	go db.run()
	db.wait.Add(1)
	return db, nil
}
