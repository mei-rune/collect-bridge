package snmp

import (
	"commons"
	"errors"
	"time"
)

var (
	is_unit_test_for_client_mgr = false
)

type ClientManager struct {
	commons.Svc
	clients map[string]Client
}

func (svc *ClientManager) Init() {
	svc.clients = make(map[string]Client)
	svc.Set(func() {
		go svc.heartbeat()
	}, nil, func() {
		svc.Test()
	})
}

func (svc *ClientManager) Test() error {

	clients := make(map[string]Client)
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

func (mgr *ClientManager) GetClient(host string) (client Client, err error) {
	values := mgr.Call(5*time.Minute, func() (Client, error) { return mgr.createClient(host) })
	if nil != values[0] {
		client = values[0].(Client)
	}
	if nil != values[1] {
		err = values[1].(error)
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

	mgr.clients = make(map[string]Client)
}

func (mgr *ClientManager) deleteClient(addr string) {
	host := NormalizeAddress(addr)
	if cl, ok := mgr.clients[host]; ok {
		if start, ok := cl.(commons.Startable); ok {
			start.Stop()
		}
		delete(mgr.clients, host)
		mgr.INFO.Printf("host '" + host + "' is inactive, delete it from manager.")
	}
}

func (mgr *ClientManager) createClient(addr string) (client Client, err error) {
	host := NormalizeAddress(addr)
	cl, ok := mgr.clients[host]
	if ok {
		client = cl
		err = nil
	} else {
		if is_unit_test_for_client_mgr {
			client, err = &TestClient{}, nil
		} else {
			client, err = NewSnmpClient(host)
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

func (self *TestClient) Test() error {
	self.test = true
	return errors.New("time out")
}

func (self *TestClient) CreatePDU(op SnmpType, version SnmpVersion) (PDU, SnmpCodeError) {
	return nil, nil
}
func (self *TestClient) SendAndRecv(req PDU, timeout time.Duration) (PDU, SnmpCodeError) {
	return nil, nil
}
