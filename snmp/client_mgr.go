package snmp

import (
	"bytes"
	"commons"
	"errors"
	"expvar"
	"fmt"
	"github.com/runner-mei/snmpclient"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	is_unit_test_for_client_mgr       = false
	id_seq                      int32 = 0
)

type ClientManager struct {
	commons.Svc
	clients map[string]snmpclient.Client
}

func (svc *ClientManager) Init() {
	svc.Name = "client_manager" + strconv.Itoa(int(atomic.AddInt32(&id_seq, 1)))
	svc.clients = make(map[string]snmpclient.Client)
	svc.Set(func() {
		expvar.Publish(svc.Name+"_clients", expvar.Func(func() interface{} {
			return fmt.Sprint(len(svc.clients))
		}))

		expvar.Publish(svc.Name+"_requests", expvar.Func(func() interface{} {
			var buf bytes.Buffer
			for id, client := range svc.clients {
				buf.WriteString(id)
				buf.WriteString(":")
				buf.WriteString(client.Stats())
				buf.WriteString("\r\n")
			}
			return buf.String()
		}))

		go svc.heartbeat()
	}, nil, func() {
		svc.Test()
	})
}

func (svc *ClientManager) Test() error {

	clients := make(map[string]snmpclient.Client)
	for host, cl := range svc.clients {
		clients[host] = cl
	}

	go func() {
		// run at anther thread, don`t block main thread.
		// becase Test() may is expensively while request is too much.
		deleted := make([]string, 0, 10)
		for host, cl := range clients {
			if t, ok := cl.(commons.Testable); ok {
				if e := t.Test(); nil != e && !commons.IsTimeout(e) {
					deleted = append(deleted, host)
				}
			}
		}

		for _, host := range deleted {
			svc.RemoveClient(host)
		}
	}()

	return nil
}

func (svc *ClientManager) heartbeat() {
	count := 0
	for svc.IsRunning() {
		time.Sleep(1 * time.Second)
		count++
		if count%(5*60) == 0 {
			svc.SafelyCall(5*time.Minute, func() {
				svc.Test()
			})
		}
	}
}

func (mgr *ClientManager) GetClient(host string) (client snmpclient.Client, err error) {
	values := mgr.SafelyCall(5*time.Minute, func() (snmpclient.Client, error) { return mgr.createClient(host) })
	if 2 <= len(values) {
		if nil != values[0] {
			client = values[0].(snmpclient.Client)
		}
		if nil != values[1] {
			err = values[1].(error)
		}
	} else if 1 == len(values) {
		if nil != values[0] {
			err = values[0].(error)
		}
	} else {
		err = errors.New("return none values")
	}
	return
}

func (mgr *ClientManager) RemoveAllClients() {
	mgr.Send(func() { mgr.deleteAllClients() })
}

func (mgr *ClientManager) RemoveClient(host string) {
	mgr.Send(func() { mgr.deleteClient(host) })
}

func (mgr *ClientManager) deleteAllClients() {
	for _, cl := range mgr.clients {
		if start, ok := cl.(commons.Startable); ok {
			start.Stop()
		}
	}

	mgr.clients = make(map[string]snmpclient.Client)
}

func (mgr *ClientManager) deleteClient(addr string) {
	host := snmpclient.NormalizeAddress(addr)
	if cl, ok := mgr.clients[host]; ok {
		if start, ok := cl.(commons.Startable); ok {
			start.Stop()
		}
		delete(mgr.clients, host)
		mgr.INFO.Printf("host '" + host + "' is inactive, delete it from manager.")
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
			client, err = snmpclient.NewSnmpClient(host)
		}
		if nil != err {
			return
		}

		start, ok := client.(commons.Startable)
		if ok {
			err = start.Start()
			if nil != err {
				return
			}
		}
		mgr.clients[host] = client
	}
	return
}

type TestClient struct {
	start, stop, test bool
}

func (self *TestClient) Start() error {
	self.start = true
	return nil
}

func (self *TestClient) Stop() {
	self.stop = true
}

func (self *TestClient) Stats() string {
	return "test"
}
func (self *TestClient) Test() error {
	self.test = true
	return errors.New("it is expired.")
}

func (self *TestClient) CreatePDU(op snmpclient.SnmpType, version snmpclient.SnmpVersion) (snmpclient.PDU, snmpclient.SnmpError) {
	return nil, nil
}
func (self *TestClient) SendAndRecv(req snmpclient.PDU, timeout time.Duration) (snmpclient.PDU, snmpclient.SnmpError) {
	return nil, nil
}
