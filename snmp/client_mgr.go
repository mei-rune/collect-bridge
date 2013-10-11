package snmp

import (
	"bytes"
	"expvar"
	"fmt"
	"github.com/runner-mei/snmpclient"
	"log"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	is_unit_test_for_client_mgr       = false
	id_seq                      int32 = 0
)

type mgrRequest struct {
	c      chan *mgrRequest
	host   string
	client snmpclient.Client
	err    error
}

var (
	requests_mutex sync.Mutex
	requests_cache = newRequestBuffer(make([]*mgrRequest, 200))
)

func init() {
	expvar.Publish("mgr_request_cache", expvar.Func(func() interface{} {
		requests_mutex.Lock()
		size := requests_cache.Size()
		requests_mutex.Unlock()
		return size
	}))
}

func newRequest() *mgrRequest {
	requests_mutex.Lock()
	cached := requests_cache.Pop()
	requests_mutex.Unlock()
	if nil != cached {
		return cached
	}
	return &mgrRequest{c: make(chan *mgrRequest, 1)}
}

func releaseRequest(will_cache *mgrRequest) {
	will_cache.host = ""
	will_cache.client = nil
	will_cache.err = nil

	requests_mutex.Lock()
	requests_cache.Push(will_cache)
	requests_mutex.Unlock()
}

type ClientManager struct {
	is_closed     int32
	wait          sync.WaitGroup
	c             chan *mgrRequest
	remove_c      chan string
	clients       map[string]snmpclient.Client
	poll_interval time.Duration

	last_error string
}

func (mgr *ClientManager) init() error {
	mgr.c = make(chan *mgrRequest, 10)
	mgr.remove_c = make(chan string, 10)
	mgr.clients = make(map[string]snmpclient.Client)
	if 1*time.Second > mgr.poll_interval {
		mgr.poll_interval = 1 * time.Second
	}

	name := "client_manager" + strconv.Itoa(int(atomic.AddInt32(&id_seq, 1)))
	expvar.Publish(name+"_clients", expvar.Func(func() interface{} {
		return fmt.Sprint(len(mgr.clients))
	}))

	expvar.Publish(name+"_requests", expvar.Func(func() interface{} {
		results := map[string]interface{}{"queue_size": len(mgr.c)}
		for id, client := range mgr.clients {
			results[id] = client.Stats()
		}
		return results
	}))

	expvar.Publish(name+"_last_error", expvar.Func(func() interface{} {
		return mgr.last_error
	}))

	go mgr.serve()
	mgr.wait.Add(1)
	return nil
}

func (mgr *ClientManager) Close() {
	if !atomic.CompareAndSwapInt32(&mgr.is_closed, 0, 1) {
		return
	}

	close(mgr.c)
	mgr.wait.Wait()
}

func (mgr *ClientManager) serve() {
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
			mgr.last_error = buffer.String()
		}
		atomic.StoreInt32(&mgr.is_closed, 1)
		mgr.wait.Done()
	}()

	count := 0
	ticker := time.NewTicker(mgr.poll_interval)
	defer ticker.Stop()

	is_running := true
	for is_running {
		select {
		case request, ok := <-mgr.c:
			if !ok {
				is_running = false
				break
			}
			request.client, request.err = mgr.createClient(request.host)
			request.c <- request
		case nm := <-mgr.remove_c:
			if 0 == len(nm) {
				mgr.deleteAllClients()
			} else {
				mgr.deleteClient(nm)
			}
		case <-ticker.C:
			count += 1
			mgr.fireTrick(count)
		}
	}
}

func (mgr *ClientManager) fireTrick(count int) {
	if 0 != count%10 {
		return
	}

	clients := make(map[string]snmpclient.Client)
	for host, cl := range mgr.clients {
		clients[host] = cl
	}

	go func() {
		// run at anther thread, don`t block main thread.
		// becase Test() may is expensively while request is too much.
		deleted := make([]string, 0, 10)
		for host, cl := range clients {
			if cl.IsExpired() {
				deleted = append(deleted, host)
			}
		}

		for _, host := range deleted {
			mgr.RemoveClient(host)
		}
	}()
}

func (mgr *ClientManager) GetClient(host string) (snmpclient.Client, error) {
	request := &mgrRequest{
		c:    make(chan *mgrRequest, 1),
		host: host}
	mgr.c <- request
	request = <-request.c
	return request.client, request.err
}

func (mgr *ClientManager) RemoveAllClients() {
	mgr.remove_c <- ""
}

func (mgr *ClientManager) RemoveClient(host string) {
	mgr.remove_c <- host
}

func (mgr *ClientManager) deleteAllClients() {
	for _, cl := range mgr.clients {
		cl.Close()
	}

	mgr.clients = make(map[string]snmpclient.Client)
	log.Printf("delete all client from manager.")
}

func (mgr *ClientManager) deleteClient(addr string) {
	host := snmpclient.NormalizeAddress(addr)
	if cl, ok := mgr.clients[host]; ok {
		cl.Close()
		delete(mgr.clients, host)
		log.Printf("host '" + host + "' is inactive, delete it from manager.")
	}
}

func (mgr *ClientManager) createClient(addr string) (client snmpclient.Client, err error) {
	host := snmpclient.NormalizeAddress(addr)
	cl, ok := mgr.clients[host]
	if ok {
		client = cl
		err = nil
	} else {
		if is_unit_test_for_client_mgr {
			client, err = &TestClient{}, nil
		} else {
			client, err = snmpclient.NewSnmpClientWith(host, mgr.poll_interval,
				&snmpclient.NullWriter{}, &snmpclient.LogWriter{})
		}
		if nil != err {
			return
		}

		mgr.clients[host] = client
	}
	return
}

type TestClient struct {
	stop, test bool
}

func (self *TestClient) Close() {
	self.stop = true
}

func (self *TestClient) Stats() interface{} {
	return "test"
}
func (self *TestClient) IsExpired() bool {
	self.test = true
	return true
}

func (self *TestClient) CreatePDU(op snmpclient.SnmpType, version snmpclient.SnmpVersion) (snmpclient.PDU, snmpclient.SnmpError) {
	return nil, nil
}
func (self *TestClient) SendAndRecv(req snmpclient.PDU, timeout time.Duration) (snmpclient.PDU, snmpclient.SnmpError) {
	return nil, nil
}
