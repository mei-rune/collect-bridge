package snmp

import (
	"commons"
	"time"
)

type ClientManager struct {
	commons.Svc
	clients map[string]Client
}

func (svc *ClientManager) Init() {
	svc.clients = make(map[string]Client)
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

func (mgr *ClientManager) createClient(host string) (client Client, err error) {
	cl, ok := mgr.clients[host]
	if ok {
		client = cl
		err = nil
	} else {
		client, err = NewSnmpClient(host)
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
