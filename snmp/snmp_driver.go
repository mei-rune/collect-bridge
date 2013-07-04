package snmp

import (
	"commons"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SnmpDriver struct {
	ClientManager
	drvMgr  *commons.DriverManager
	timeout time.Duration
}

func NewSnmpDriver(timeout time.Duration, drvMgr *commons.DriverManager) *SnmpDriver {
	drv := &SnmpDriver{drvMgr: drvMgr, timeout: timeout}
	drv.ClientManager.Init()
	return drv
}

func getTimeout(params map[string]string, timeout time.Duration) time.Duration {
	v, ok := params["timeout"]
	if !ok {
		return timeout
	}

	ret, err := time.ParseDuration(v)
	if nil != err {
		panic(err)
	}
	return ret
}

func getVersion(params map[string]string) (SnmpVersion, error) {
	v, ok := params["snmp.version"]
	if !ok {
		return SNMP_V2C, nil
	}
	return parseVersion(v)
}

func parseVersion(v string) (SnmpVersion, error) {
	switch v {
	case "v1", "V1", "1":
		return SNMP_V1, nil
	case "v2", "V2", "v2c", "V2C", "2", "2c", "2C":
		return SNMP_V2C, nil
	case "v3", "V3", "3":
		return SNMP_V3, nil
	}
	return SNMP_Verr, errors.New("Unsupported version - " + v)
}

func getAction(params map[string]string) (SnmpType, error) {
	v, ok := params["snmp.action"]
	if !ok {
		return SNMP_PDU_GET, nil
	}
	switch v {
	case "table", "Table", "TABLE":
		return SNMP_PDU_TABLE, nil
	case "get", "Get", "GET":
		return SNMP_PDU_GET, nil
	case "next", "Next", "NEXT", "getnext", "Getnext", "GETNEXT":
		return SNMP_PDU_GETNEXT, nil
	case "bulk", "Bulk", "BULK", "getbuld", "Getbuld", "GETBULD":
		return SNMP_PDU_GETBULK, nil
	case "set", "Set", "SET", "put", "Put", "PUT":
		return SNMP_PDU_SET, nil
	}
	return SNMP_PDU_GET, fmt.Errorf("error pdu type: %s", v)
}

func internalError(msg string, err error) error {
	if nil == err {
		return errors.New(msg)
	}
	if 0 == len(msg) {
		return err
	}
	return errors.New(msg + ", " + err.Error())
}

func internalErrorResult(msg string, err error) commons.Result {
	if nil == err {
		return commons.ReturnWithInternalError(msg)
	}
	if 0 == len(msg) {
		return commons.ReturnWithInternalError(err.Error())
	}
	return commons.ReturnWithInternalError(msg + ", " + err.Error())
}

var HostIsRequired = commons.IsRequired("snmp.host")
var OidIsRequired = commons.IsRequired("snmp.oid")

func getHost(params map[string]string) (string, error) {
	host, ok := params["snmp.host"]
	if !ok {
		if address, ok := params["snmp.address"]; ok {
			host = address
			if port, ok := params["snmp.port"]; ok && 0 != len(port) {
				host += (":" + port)
			} else {
				host += ":161"
			}
		} else if host, ok = params["id"]; !ok {
			return "", HostIsRequired
		}
	}
	return host, nil
}

func (self *SnmpDriver) invoke(action SnmpType, params map[string]string) commons.Result {
	host, e := getHost(params)
	if nil != e {
		return commons.ReturnWithBadRequest(e.Error())
	}

	oid, ok := params["snmp.oid"]
	if !ok || 0 == len(oid) {
		return commons.ReturnWithIsRequired("snmp.oid")
	}

	client, err := self.GetClient(host)
	if nil != err {
		return internalErrorResult("create client failed", err)
	}

	if SNMP_PDU_TABLE == action {
		columns, contains := params["snmp.columns"]

		if contains && 0 != len(columns) {
			return self.tableGetByColumns(params, client, oid, columns)
		}
		return self.tableGet(params, client, oid)
	}

	version, e := getVersion(params)
	if SNMP_Verr == version {
		return commons.ReturnWithBadRequest(e.Error())
	}

	req, err := client.CreatePDU(action, version)
	if nil != err {
		return internalErrorResult("create pdu failed", err)
	}

	err = req.Init(params)
	if nil != err {
		return internalErrorResult("init pdu failed", err)
	}

	switch action {
	case SNMP_PDU_GET, SNMP_PDU_GETNEXT:
		err = req.GetVariableBindings().Append(oid, "")
	case SNMP_PDU_GETBULK:
		vbs := req.GetVariableBindings()
		for _, oi := range strings.Split(oid, "|") {
			err = vbs.Append(oi, "")
			if nil != err {
				break
			}
		}
	case SNMP_PDU_SET:
		txt, ok := params["body"]
		if !ok {
			err = commons.BodyNotExists
		} else {
			err = req.GetVariableBindings().Append(oid, txt)
		}
	default:
		err = fmt.Errorf("unknown pdu type %d", int(action))
	}

	if nil != err {
		return internalErrorResult("append vb failed", err)
	}
	resp, err := client.SendAndRecv(req, getTimeout(params, self.timeout))
	if nil != err {
		return internalErrorResult("", err)
	}

	if 0 == resp.GetVariableBindings().Len() {
		return internalErrorResult("snmp result is empty", nil)
	}
	results := make(map[string]interface{})
	for _, vb := range resp.GetVariableBindings().All() {
		if vb.Value.IsError() && 1 == req.GetVariableBindings().Len() {
			return internalErrorResult("result is error", nil)
		}

		results[vb.Oid.GetString()] = vb.Value
	}

	return commons.Return(results)
}

func (self *SnmpDriver) Get(params map[string]string) commons.Result {
	action, err := getAction(params)
	if nil != err {
		return internalErrorResult("get action failed", err)
	}
	return self.invoke(action, params)
}

func (self *SnmpDriver) Put(params map[string]string) commons.Result {
	return self.invoke(SNMP_PDU_SET, params)
}

func (self *SnmpDriver) Create(params map[string]string) commons.Result {
	return commons.ReturnWithNotImplemented()
}

func (self *SnmpDriver) Delete(params map[string]string) commons.Result {
	action, ok := params["snmp.action"]
	if ok && "remove_client" == action {
		host, _ := getHost(params)
		if "" != host && "all" != host {
			self.RemoveClient(host)
		} else {
			self.RemoveAllClients()
		}

		return commons.Return(true)
	}

	return commons.ReturnWithNotImplemented()
}

var (
	NilVariableBinding = VariableBinding{Oid: *NewOid([]uint32{}), Value: NewSnmpNil()}
)

func (self *SnmpDriver) getNext(params map[string]string, client Client, next_oid SnmpOid,
	version SnmpVersion, timeout time.Duration) (VariableBinding, error) {
	var err error

	req, err := client.CreatePDU(SNMP_PDU_GETNEXT, version)
	if nil != err {
		return NilVariableBinding, internalError("create pdu failed", err)
	}

	err = req.Init(params)
	if nil != err {
		return NilVariableBinding, internalError("init pdu failed", err)
	}

	err = req.GetVariableBindings().AppendWith(next_oid, NewSnmpNil())
	if nil != err {
		return NilVariableBinding, internalError("append vb failed", err)
	}

	resp, err := client.SendAndRecv(req, timeout)
	if nil != err {
		return NilVariableBinding, internalError("snmp failed", err)
	}

	if 0 == resp.GetVariableBindings().Len() {
		return NilVariableBinding, internalError("snmp result is empty", nil)
	}

	return resp.GetVariableBindings().Get(0), nil
}

func (self *SnmpDriver) tableGet(params map[string]string, client Client,
	oid string) commons.Result {

	start_oid, err := ParseOidFromString(oid)
	if nil != err {
		return internalErrorResult("param 'oid' is error", err)
	}
	oid_s := start_oid.GetString()
	version, e := getVersion(params)
	if SNMP_Verr == version {
		return commons.ReturnWithBadRequest(e.Error())
	}

	timeout := getTimeout(params, self.timeout)
	next_oid := start_oid
	results := make(map[string]interface{})

	for {
		vb, err := self.getNext(params, client, next_oid, version, timeout)
		if nil != err {
			return commons.ReturnWithInternalError(err.Error())
		}

		if !strings.HasPrefix(vb.Oid.GetString(), oid_s) {
			break
		}

		sub := vb.Oid.GetUint32s()[len(start_oid):]
		if 2 > len(sub) {
			return commons.ReturnError(commons.InternalErrorCode,
				fmt.Sprintf("read '%s' return '%s', it is incorrect - value is %s",
					next_oid.GetString(), vb.Oid.GetString(), vb.Value.String()))
		}

		idx := strconv.FormatUint(uint64(sub[0]), 10)
		keys := NewOid(sub[1:]).GetString()

		row, _ := results[keys].(map[string]interface{})
		if nil == row {
			row = make(map[string]interface{})
			results[keys] = row
		}
		row[idx] = vb.Value

		next_oid = vb.Oid
	}
	if 0 == len(results) {
		return internalErrorResult("snmp result is empty", nil)
	}

	return commons.Return(results)
}

func (self *SnmpDriver) tableGetByColumns(params map[string]string, client Client,
	oid, columns_str string) commons.Result {

	start_oid, err := ParseOidFromString(oid)
	if nil != err {
		return internalErrorResult("param 'oid' is error", err)
	}
	columns, err := commons.ConvertToIntList(columns_str, ",")
	if nil != err {
		return internalErrorResult("param 'columns' is error", err)
	}

	version, e := getVersion(params)
	if SNMP_Verr == version {
		return commons.ReturnWithBadRequest(e.Error())
	}

	timeout := getTimeout(params, self.timeout)
	next_oids := make([]SnmpOid, 0, len(columns))
	next_oids_s := make([]string, 0, len(columns))
	for _, i := range columns {
		o := start_oid.Concat(i)
		next_oids = append(next_oids, o)
		next_oids_s = append(next_oids_s, o.GetString()+".")
	}

	results := make(map[string]interface{})
	for {
		var req PDU
		req, err = client.CreatePDU(SNMP_PDU_GETNEXT, version)
		if nil != err {
			return internalErrorResult("create pdu failed", err)
		}

		err = req.Init(params)
		if nil != err {
			return internalErrorResult("init pdu failed", err)
		}

		for _, next_oid := range next_oids {
			err = req.GetVariableBindings().AppendWith(next_oid, NewSnmpNil())
			if nil != err {
				return internalErrorResult("append vb failed", err)
			}
		}

		resp, err := client.SendAndRecv(req, timeout)
		if nil != err {
			return internalErrorResult("snmp failed", err)
		}

		if len(next_oids) != resp.GetVariableBindings().Len() {
			return internalErrorResult(fmt.Sprintf("number of result is mismatch, excepted is %d, actual is %d",
				len(next_oids), resp.GetVariableBindings().Len()), nil)
		}

		offset := 0
		for i, vb := range resp.GetVariableBindings().All() {

			if !strings.HasPrefix(vb.Oid.GetString(), next_oids_s[i]) {
				continue
			}

			sub := vb.Oid.GetUint32s()[len(start_oid)+1:]
			if 1 > len(sub) {
				return commons.ReturnError(commons.InternalErrorCode,
					fmt.Sprintf("read '%s' return '%s', it is incorrect - value is %s",
						start_oid.GetString(), vb.Oid.GetString(), vb.Value.String()))
			}

			keys := NewOid(sub).GetString()

			row, _ := results[keys].(map[string]interface{})
			if nil == row {
				row = make(map[string]interface{})
				results[keys] = row
			}

			row[strconv.FormatInt(int64(columns[i]), 10)] = vb.Value

			next_oids[offset] = vb.Oid
			if offset != i {
				next_oids_s[offset] = next_oids_s[i]
				columns[offset] = columns[i]
			}

			offset++
		}
		if 0 == offset {
			break
		}
		next_oids_s = next_oids_s[0:offset]
		columns = columns[0:offset]
		next_oids = next_oids[0:offset]
	}
	if 0 == len(results) {
		return internalErrorResult("snmp result is empty", nil)
	}
	return commons.Return(results)
}
