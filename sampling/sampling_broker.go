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
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SamplingBroker struct {
	name          string
	action        string
	url           string
	exchange_c    chan *requestCtx
	chan_c        chan *channelRequest
	cached_buffer bytes.Buffer

	closed     int32
	wait       sync.WaitGroup
	last_error *expvar.String

	request_id      uint64
	pendingRequests map[uint64]*requestCtx
	channelGroups   map[string]*channelGroup
}

type channelGroup struct {
	name          string
	last_begin_at time.Time
	last_end_at   time.Time
	channels      map[string]chan interface{}
}

func (self *SamplingBroker) createClient(channelName string, c chan interface{},
	method, metric_name, managedType, managedId, pathStr string, params map[string]string,
	body interface{}, cached_timeout time.Duration) (*clientImpl, error) {
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

	cl := &clientImpl{id: channelName + "/" + commons.GenerateId(),
		broker: self,
		request: ExchangeRequest{ChannelName: channelName,
			Action:      method,
			Name:        metric_name,
			ManagedType: managedType,
			ManagedId:   managedId,
			Paths:       paths,
			Params:      params,
			Body:        body}}

	cl.ctx.created_at = time.Now()
	cl.ctx.cached_timeout = cached_timeout
	cl.ctx.request = &cl.request
	return cl, nil
}

func (self *SamplingBroker) CreateClient(method, metric_name, managedType, managedId, pathStr string,
	params map[string]string, body interface{}) (Client, error) {
	cl, e := self.createClient("", nil, method, metric_name, managedType, managedId, pathStr, params, body, 0)
	if nil != e {
		return nil, e
	}
	return cl, nil
}
func (self *SamplingBroker) SubscribeClient(channelName string, c chan interface{},
	method, metric_name, managedType, managedId, pathStr string, params map[string]string,
	body interface{}, cached_timeout time.Duration) (ChannelClient, error) {
	if nil == c {
		return nil, errors.New("chan is nil.")
	}
	cl, e := self.createClient(channelName, c, method, metric_name, managedType, managedId, pathStr, params, body, cached_timeout)
	if nil != e {
		return nil, e
	}

	grp, e := self.Subscribe(channelName, cl.id, c)
	if nil != e {
		return nil, e
	}
	cl.ctx.is_subscribed = true
	cl.ctx.grp = grp
	cl.ctx.c = c

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

		grp = &channelGroup{name: chanD.channelName, channels: make(map[string]chan interface{})}
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

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	cached_objects := make([]*requestCtx, 0, 1000)
	cached_requests := make([]*ExchangeRequest, 0, 1000)

	count := 0
	is_running := true
	for is_running {
		select {
		case <-ticker.C:
			count += 1
			if 0 == count%20 {
				self.onIdle()
			}

			if 0 == len(self.exchange_c) {
				select {
				case self.exchange_c <- nil:
				default:
				}
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

func (self *SamplingBroker) recvObjects(objects []*requestCtx, max_size int) []*requestCtx {
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
		interval := now.Sub(ctx.created_at)
		if interval > *snmp_timeout {
			self.replyError(ctx.c, timeoutError)
			expired = append(expired, k)
		}
	}

	for _, k := range expired {
		delete(self.pendingRequests, k)
	}
}

var empty_id_list = []uint64{}

func (self *SamplingBroker) runOnce(cached_array []*requestCtx, cached_requests []*ExchangeRequest, max_size int) {
	objects := self.recvObjects(cached_array, max_size)

	id_list := empty_id_list
	if 0 != len(objects) {
		id_list = make([]uint64, len(objects))
		now := time.Now()
		for idx, obj := range objects {
			if nil != obj.grp {
				if now.Sub(obj.grp.last_begin_at) < obj.cached_timeout {
					continue
				}

				if now.Sub(obj.grp.last_end_at) < obj.cached_timeout {
					continue
				}

				obj.grp.last_begin_at = now
			}
			if obj.is_subscribed {
				id_list[idx] = 0
			} else {
				self.request_id += 1
				id_list[idx] = self.request_id
			}
			cached_requests = append(cached_requests, obj.request)
		}
	}

	resposes, e := self.exchange(id_list, cached_requests)
	if nil != e {
		for _, obj := range objects {
			if nil == obj.c {
				continue
			}
			self.replyError(obj.c, e)
		}
		return
	}

	now := time.Now()
	for idx, obj := range objects {
		if nil == obj.c {
			continue
		}

		if !obj.is_subscribed {
			obj.created_at = now
			self.pendingRequests[id_list[idx]] = obj
		}
	}

	grp_failed := make([]string, 0, 10)
	failed := make([]string, 0, 10)
	for _, res := range resposes {
		if pending, ok := self.pendingRequests[res.Id]; ok {
			delete(self.pendingRequests, res.Id)
			self.reply(pending.c, res)
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

	// TODO: 因为和trigger产生死锁，暂时在这里解决一下，请尽快重构 trigger.
	// select {
	// case c <- response:
	// default:
	// }
	return nil
}

func exchangeTo(method, url string, id_list []uint64, requests []*ExchangeRequest, buffer *bytes.Buffer) ([]*ExchangeResponse, commons.RuntimeError) {
	buffer.Reset()
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
	return exchangeTo(self.action, self.url, id_list, requests, &self.cached_buffer)
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
		exchange_c:      make(chan *requestCtx, 1000),
		chan_c:          make(chan *channelRequest),
		closed:          0,
		last_error:      varString,
		request_id:      0,
		pendingRequests: make(map[uint64]*requestCtx),
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
