package main

import (
	"errors"
	"flag"
	"fmt"
	"snmp"
	"strings"
	"time"
	"web"
)

var (
	snmpTimeout = flag.Int("snmp.timeout", 5*60, "maximun duration (second) of send/recv pdu timeout")
)

type SnmpDriver struct {
	snmp.ClientManager
}

func getTimeout(params map[string]string) time.Duration {
	v, ok := params["timeout"]
	if !ok {
		return time.Duration(*snmpTimeout) * time.Second
	}
	ret, err := snmp.ParseTime(v)
	if nil != err {
		panic(err)
	}
	return ret
}

func getVersion(params map[string]string) snmp.SnmpVersion {
	v, ok := params["version"]
	if !ok {
		return snmp.SNMP_V2C
	}
	switch v {
	case "v1", "V1", "1":
		return snmp.SNMP_V1
	case "v2", "V2", "v2c", "V2C", "2", "2c", "2C":
		return snmp.SNMP_V2C
	case "v3", "V3", "3":
		return snmp.SNMP_V3
	}
	panic(errors.New(fmt.Sprintf("error version: %s", v)))
	return 0
}

func getAction(params map[string]string) (snmp.SnmpType, error) {
	v, ok := params["action"]
	if !ok {
		return snmp.SNMP_PDU_GET, nil
	}
	switch v {
	case "get", "Get", "GET":
		return snmp.SNMP_PDU_GET, nil
	case "next", "Next", "NEXT", "getnext", "Getnext", "GETNEXT":
		return snmp.SNMP_PDU_GETNEXT, nil
	case "bulk", "Bulk", "BULK", "getbuld", "Getbuld", "GETBULD":
		return snmp.SNMP_PDU_GETBULK, nil
	case "set", "Set", "SET", "put", "Put", "PUT":
		return snmp.SNMP_PDU_SET, nil
	}
	return snmp.SNMP_PDU_GET, errors.New(fmt.Sprintf("error pdu type: %s", v))
}

func internalError(msg string, err error) error {
	return fmt.Errorf(msg + "-" + err.Error())
}

func (bridge *SnmpDriver) invoke(action snmp.SnmpType, params map[string]string) (interface{}, error) {
	host, ok := params["host"]
	if !ok {
		return nil, errors.New("'host' is required.")
	}

	oid, ok := params["oid"]
	if !ok {
		return nil, errors.New("'oid' is required.")
	}

	client, err := bridge.GetClient(host)
	if nil != err {
		return nil, internalError("create client failed", err)
	}

	req, err := client.CreatePDU(action, getVersion(params))
	if nil != err {
		return nil, internalError("create pdu failed", err)
	}

	err = req.Init(params)
	if nil != err {
		return nil, internalError("init pdu failed", err)
	}

	switch action {
	case snmp.SNMP_PDU_GET, snmp.SNMP_PDU_GETNEXT:
		err = req.GetVariableBindings().Append(oid, "")
	case snmp.SNMP_PDU_GETBULK:
		vbs := req.GetVariableBindings()
		for _, oi := range strings.Split(oid, "|") {
			err = vbs.Append(oi, "")
			if nil != err {
				break
			}
		}
	case snmp.SNMP_PDU_SET:
		txt, ok := params["body"]
		if !ok {
			err = errors.New("'body' is required in the set action.")
		} else {
			err = req.GetVariableBindings().Append(oid, txt)
		}
	default:
		err = fmt.Errorf("unknown pdu type %d", int(action))
	}

	if nil != err {
		return nil, internalError("append vb failed", err)
	}

	resp, err := client.SendAndRecv(req, getTimeout(params))
	if nil != err {
		return nil, internalError("snmp failed", err)
	}

	if 0 == resp.GetVariableBindings().Len() {
		return nil, internalError("result is empty", nil)
	}

	results := make(map[string]string)
	for _, vb := range resp.GetVariableBindings().All() {
		results[vb.Oid.GetString()] = vb.Value.String()
	}
	return results, nil
}

func (bridge *SnmpDriver) Get(params map[string]string) (interface{}, error) {
	action, err := getAction(params)
	if nil != err {
		return nil, internalError("get action failed", err)
	}
	return bridge.invoke(action, params)
}

func (bridge *SnmpDriver) Put(params map[string]string) (interface{}, error) {
	return bridge.invoke(snmp.SNMP_PDU_SET, params)
}

func (bridge *SnmpDriver) Create(map[string]string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (bridge *SnmpDriver) Delete(map[string]string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (client *SnmpDriver) Table(ctx *web.Context, host, oids string) {
	panic(fmt.Sprintf("not implemented"))
}
