package main

import (
	"commons"
	"errors"
	"flag"
	"fmt"
	"snmp"
	"strconv"
	"strings"
	"time"
)

var (
	snmpTimeout = flag.Int("snmp.timeout", 60, "maximun duration (second) of send/recv pdu timeout")
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
	panic(fmt.Sprintf("error version: %s", v))
	return 0
}

func getAction(params map[string]string) (snmp.SnmpType, error) {
	v, ok := params["action"]
	if !ok {
		return snmp.SNMP_PDU_GET, nil
	}
	switch v {
	case "table", "Table", "TABLE":
		return snmp.SNMP_PDU_TABLE, nil
	case "get", "Get", "GET":
		return snmp.SNMP_PDU_GET, nil
	case "next", "Next", "NEXT", "getnext", "Getnext", "GETNEXT":
		return snmp.SNMP_PDU_GETNEXT, nil
	case "bulk", "Bulk", "BULK", "getbuld", "Getbuld", "GETBULD":
		return snmp.SNMP_PDU_GETBULK, nil
	case "set", "Set", "SET", "put", "Put", "PUT":
		return snmp.SNMP_PDU_SET, nil
	}
	return snmp.SNMP_PDU_GET, fmt.Errorf("error pdu type: %s", v)
}

func internalError(msg string, err error) error {
	if nil == err {
		return errors.New(msg)
	}
	return fmt.Errorf(msg + "-" + err.Error())
}

func (bridge *SnmpDriver) invoke(action snmp.SnmpType, params map[string]string) (map[string]interface{}, error) {
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

	if snmp.SNMP_PDU_TABLE == action {
		_, contains := params["columns"]

		if contains {
			return bridge.tableGetByColumns(params, client, oid)
		}

		return bridge.tableGet(params, client, oid)
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
	return map[string]interface{}{"value": results}, nil
}

func (bridge *SnmpDriver) Get(params map[string]string) (map[string]interface{}, error) {
	action, err := getAction(params)
	if nil != err {
		return nil, internalError("get action failed", err)
	}
	return bridge.invoke(action, params)
}

func (bridge *SnmpDriver) Put(params map[string]string) (map[string]interface{}, error) {
	return bridge.invoke(snmp.SNMP_PDU_SET, params)
}

func (bridge *SnmpDriver) Create(params map[string]string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (bridge *SnmpDriver) Delete(params map[string]string) (bool, error) {
	v, ok := params["remove_clients"]
	if ok && "true" == v {
		host, _ := params["client"]
		if "" != host {
			bridge.RemoveClient(host)
		} else {
			bridge.RemoveAllClients()
		}

		return true, nil
	}

	return false, fmt.Errorf("not implemented")
}

var (
	NilVariableBinding = snmp.VariableBinding{Oid: *snmp.NewOid([]uint32{}), Value: snmp.NewSnmpNil()}
)

func (bridge *SnmpDriver) getNext(params map[string]string, client snmp.Client, next_oid snmp.SnmpOid,
	version snmp.SnmpVersion, timeout time.Duration) (snmp.VariableBinding, error) {
	var err error

	req, err := client.CreatePDU(snmp.SNMP_PDU_GETNEXT, version)
	if nil != err {
		return NilVariableBinding, internalError("create pdu failed", err)
	}

	err = req.Init(params)
	if nil != err {
		return NilVariableBinding, internalError("init pdu failed", err)
	}

	err = req.GetVariableBindings().AppendWith(next_oid, snmp.NewSnmpNil())
	if nil != err {
		return NilVariableBinding, internalError("append vb failed", err)
	}

	resp, err := client.SendAndRecv(req, timeout)
	if nil != err {
		return NilVariableBinding, internalError("snmp failed", err)
	}

	if 0 == resp.GetVariableBindings().Len() {
		return NilVariableBinding, internalError("result is empty", nil)
	}

	return resp.GetVariableBindings().Get(0), nil
}

func (bridge *SnmpDriver) tableGet(params map[string]string, client snmp.Client,
	oid string) (map[string]interface{}, error) {

	start_oid, err := snmp.ParseOidFromString(oid)
	if nil != err {
		return nil, internalError("param 'oid' is error", err)
	}
	oid_s := start_oid.GetString()
	version := getVersion(params)
	timeout := getTimeout(params)
	next_oid := start_oid
	results := make(map[string]map[string]string)
	for {
		vb, err := bridge.getNext(params, client, next_oid, version, timeout)
		if nil != err {
			return nil, err
		}

		if !strings.HasPrefix(vb.Oid.GetString(), oid_s) {
			break
		}

		sub := vb.Oid.GetUint32s()[len(start_oid):]
		if 2 > len(sub) {
			return nil, fmt.Errorf("read '%s' return '%s', it is incorrect", next_oid.GetString(), vb.Oid.GetString())
		}

		idx := strconv.FormatUint(uint64(sub[0]), 10)
		keys := snmp.NewOid(sub[1:]).GetString()

		row, _ := results[keys]
		if nil == row {
			row = make(map[string]string)
			results[keys] = row
		}
		row[idx] = vb.Value.String()

		next_oid = vb.Oid
	}

	return map[string]interface{}{"value": results}, nil
}

func (bridge *SnmpDriver) tableGetByColumns(params map[string]string, client snmp.Client,
	oid string) (map[string]interface{}, error) {

	start_oid, err := snmp.ParseOidFromString(oid)
	if nil != err {
		return nil, internalError("param 'oid' is error", err)
	}
	columns, err := commons.GetIntList(params, "columns")
	if nil != err {
		return nil, internalError("param 'columns' is error", err)
	}

	version := getVersion(params)
	timeout := getTimeout(params)
	next_oids := make([]snmp.SnmpOid, 0, len(columns))
	next_oids_s := make([]string, 0, len(columns))
	for _, i := range columns {
		o := start_oid.Concat(i)
		next_oids = append(next_oids, o)
		next_oids_s = append(next_oids_s, o.GetString())
	}

	results := make(map[string]map[string]string)

	for {
		var req snmp.PDU
		req, err = client.CreatePDU(snmp.SNMP_PDU_GETNEXT, version)
		if nil != err {
			return nil, internalError("create pdu failed", err)
		}

		err = req.Init(params)
		if nil != err {
			return nil, internalError("init pdu failed", err)
		}

		for _, next_oid := range next_oids {
			err = req.GetVariableBindings().AppendWith(next_oid, snmp.NewSnmpNil())
			if nil != err {
				return nil, internalError("append vb failed", err)
			}
		}

		resp, err := client.SendAndRecv(req, timeout)
		if nil != err {
			return nil, internalError("snmp failed", err)
		}

		if len(next_oids) != resp.GetVariableBindings().Len() {
			return nil, internalError(fmt.Sprintf("number of result is mismatch, excepted is %d, actual is %d",
				len(next_oids), resp.GetVariableBindings().Len()), nil)
		}

		offset := 0
		for i, vb := range resp.GetVariableBindings().All() {

			if !strings.HasPrefix(vb.Oid.GetString(), next_oids_s[i]) {
				copy(next_oids[i:], next_oids[i+1:])
				copy(columns[i:], columns[i+1:])
				continue
			}

			sub := vb.Oid.GetUint32s()[len(start_oid)+1:]
			if 1 > len(sub) {
				return nil, fmt.Errorf("read '%s' return '%s', it is incorrect", start_oid, vb.Oid.GetString())
			}

			keys := snmp.NewOid(sub).GetString()

			row, _ := results[keys]
			if nil == row {
				row = make(map[string]string)
				results[keys] = row
			}
			row[strconv.FormatInt(int64(columns[i]), 10)] = vb.Value.String()

			next_oids[offset] = vb.Oid
			offset++
		}

		if 0 == offset {
			break
		}
		next_oids = next_oids[0:offset]
	}
	return map[string]interface{}{"value": results}, nil
}
