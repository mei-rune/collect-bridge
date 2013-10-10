package snmp

import (
	"commons"
	"errors"
	"fmt"
	"github.com/runner-mei/snmpclient"
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
	v, ok := params["snmp.timeout"]
	if !ok || 0 == len(v) {
		if v, ok = params["timeout"]; !ok || 0 == len(v) {
			return timeout
		}
	}

	ret, err := time.ParseDuration(v)
	if nil != err {
		panic(err)
	}
	return ret
}

func getVersion(params map[string]string) (snmpclient.SnmpVersion, error) {
	v, ok := params["snmp.version"]
	if !ok {
		return snmpclient.SNMP_V2C, nil
	}
	return snmpclient.ParseSnmpVersion(v)
}

func getAction(params map[string]string) (snmpclient.SnmpType, error) {
	v, ok := params["snmp.action"]
	if !ok {
		return snmpclient.SNMP_PDU_GET, nil
	}
	return snmpclient.ParseSnmpAction(v)
}

func internalError(oid, msg string, err error) error {
	if nil == err {
		return errors.New(msg)
	}

	if e, ok := err.(snmpclient.SnmpError); ok {
		switch e.Code() {
		case snmpclient.SNMP_CODE_SYNTAX_NOSUCHOBJECT, /* exception */
			snmpclient.SNMP_CODE_SYNTAX_NOSUCHINSTANCE, /* exception */
			snmpclient.SNMP_CODE_SYNTAX_ENDOFMIBVIEW,   /* exception */
			snmpclient.SNMP_CODE_ERR_NOSUCHNAME:
			if 0 == len(msg) {
				return commons.NotFoundWithIdAndMessage(oid, err.Error())
			}
			return commons.NotFoundWithIdAndMessage(oid, msg+", "+err.Error())
		case snmpclient.SNMP_CODE_TIMEOUT:
			return commons.TimeoutErr
		}
	}

	if 0 == len(msg) {
		return err
	}
	if e, ok := err.(commons.RuntimeError); ok {
		return commons.NewApplicationError(e.Code(), msg+", "+err.Error())
	}
	return errors.New(msg + ", " + err.Error())
}

func internalErrorResult(oid, msg string, err error) commons.Result {
	if nil == err {
		return commons.ReturnWithInternalError(msg)
	}
	if e, ok := err.(snmpclient.SnmpError); ok {
		switch e.Code() {
		case snmpclient.SNMP_CODE_SYNTAX_NOSUCHOBJECT, /* exception */
			snmpclient.SNMP_CODE_SYNTAX_NOSUCHINSTANCE, /* exception */
			snmpclient.SNMP_CODE_SYNTAX_ENDOFMIBVIEW,   /* exception */
			snmpclient.SNMP_CODE_ERR_NOSUCHNAME:
			if 0 == len(msg) {
				return commons.ReturnWithNotFoundWithMessage(oid, err.Error())
			}
			return commons.ReturnWithNotFoundWithMessage(oid, msg+", "+err.Error())
		case snmpclient.SNMP_CODE_TIMEOUT:
			return commons.ReturnError(commons.TimeoutErr.Code(), commons.TimeoutErr.Error())
		}
	}

	if 0 == len(msg) {
		if e, ok := err.(commons.RuntimeError); ok {
			return commons.ReturnError(e.Code(), msg+", "+err.Error())
		}
		return commons.ReturnWithInternalError(err.Error())
	}
	if e, ok := err.(commons.RuntimeError); ok {
		return commons.ReturnError(e.Code(), msg+", "+err.Error())
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

func (self *SnmpDriver) invoke(action snmpclient.SnmpType, params map[string]string, body interface{}) commons.Result {
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
		return internalErrorResult("", "create client failed", err)
	}

	if snmpclient.SNMP_PDU_TABLE == action {
		columns, contains := params["snmp.columns"]

		if contains && 0 != len(columns) {
			return self.tableGetByColumns(params, client, oid, columns)
		}
		return self.tableGet(params, client, oid)
	}

	version, e := getVersion(params)
	if snmpclient.SNMP_Verr == version {
		return commons.ReturnWithBadRequest(e.Error())
	}

	req, err := client.CreatePDU(action, version)
	if nil != err {
		return internalErrorResult("", "create pdu failed", err)
	}

	err = req.Init(params)
	if nil != err {
		return internalErrorResult("", "init pdu failed", err)
	}

	switch action {
	case snmpclient.SNMP_PDU_GETNEXT:
		err = req.GetVariableBindings().Append(oid, "")
	case snmpclient.SNMP_PDU_GET, snmpclient.SNMP_PDU_GETBULK:
		vbs := req.GetVariableBindings()
		for _, single_oid := range strings.Split(oid, ",") {
			//single_oid = strings.TrimSpace(single_oid)
			if 0 == len(single_oid) {
				continue
			}

			err = vbs.Append(single_oid, "")
			if nil != err {
				break
			}
		}
	case snmpclient.SNMP_PDU_SET:
		if nil == body {
			err = errors.New("'body' is nil.")
		} else if bs, ok := body.([]byte); ok {
			err = req.GetVariableBindings().Append(oid, string(bs))
		} else if txt, ok := body.(string); ok {
			err = req.GetVariableBindings().Append(oid, txt)
		} else {
			err = fmt.Errorf("'body' is not a string - %T", body)
		}
	default:
		err = fmt.Errorf("unknown pdu type %d", int(action))
	}

	if nil != err {
		return commons.ReturnWithBadRequest(err.Error())
	}

	resp, err := client.SendAndRecv(req, getTimeout(params, self.timeout))
	if nil != err {
		return internalErrorResult(oid, "", err)
	}

	if 0 == resp.GetVariableBindings().Len() {
		return internalErrorResult("", "snmp result is empty", nil)
	}

	results := make(map[string]interface{})
	for _, vb := range resp.GetVariableBindings().All() {
		if vb.Value.IsError() && 1 == req.GetVariableBindings().Len() {
			switch vb.Value.GetSyntax() {
			case snmpclient.SNMP_SYNTAX_NOSUCHOBJECT:
				return commons.ReturnError(commons.NotFoundCode, "'"+vb.Oid.GetString()+"' is nosuchobject.")
			case snmpclient.SNMP_SYNTAX_NOSUCHINSTANCE:
				return commons.ReturnError(commons.NotFoundCode, "'"+vb.Oid.GetString()+"' is nosuchinstance.")
			case snmpclient.SNMP_SYNTAX_ENDOFMIBVIEW:
				return commons.ReturnError(commons.NotFoundCode, "'"+vb.Oid.GetString()+"' is endofmibview.")
			default:
				return commons.ReturnWithInternalError("result of '" + vb.Oid.GetString() + "' is error -- " + vb.Value.String())
			}
		}

		results[vb.Oid.GetString()] = vb.Value
	}

	return commons.Return(results)
}

func (self *SnmpDriver) Get(params map[string]string) commons.Result {
	action, err := getAction(params)
	if nil != err {
		return internalErrorResult("", "get action failed", err)
	}
	return self.invoke(action, params, nil)
}

func (self *SnmpDriver) Put(params map[string]string, body interface{}) commons.Result {
	return self.invoke(snmpclient.SNMP_PDU_SET, params, body)
}

func (self *SnmpDriver) Create(params map[string]string, body interface{}) commons.Result {
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
	NilVariableBinding = snmpclient.VariableBinding{Oid: *snmpclient.NewOid([]uint32{}), Value: snmpclient.NewSnmpNil()}
)

func (self *SnmpDriver) getNext(params map[string]string, client snmpclient.Client, next_oid snmpclient.SnmpOid,
	version snmpclient.SnmpVersion, timeout time.Duration) (snmpclient.VariableBinding, error) {
	var err error

	req, err := client.CreatePDU(snmpclient.SNMP_PDU_GETNEXT, version)
	if nil != err {
		return NilVariableBinding, internalError("", "create pdu failed", err)
	}

	err = req.Init(params)
	if nil != err {
		return NilVariableBinding, internalError("", "init pdu failed", err)
	}

	err = req.GetVariableBindings().AppendWith(next_oid, snmpclient.NewSnmpNil())
	if nil != err {
		return NilVariableBinding, internalError("", "append vb failed", err)
	}

	resp, err := client.SendAndRecv(req, timeout)
	if nil != err {
		return NilVariableBinding, internalError(next_oid.GetString(), "snmp failed", err)
	}

	if 0 == resp.GetVariableBindings().Len() {
		return NilVariableBinding, internalError("", "snmp result is empty", nil)
	}

	return resp.GetVariableBindings().Get(0), nil
}

func (self *SnmpDriver) tableGet(params map[string]string, client snmpclient.Client,
	oid string) commons.Result {

	start_oid, err := snmpclient.ParseOidFromString(oid)
	if nil != err {
		return internalErrorResult("", "param 'oid' is error", err)
	}
	oid_s := start_oid.GetString()
	version, e := getVersion(params)
	if snmpclient.SNMP_Verr == version {
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
		keys := snmpclient.NewOid(sub[1:]).GetString()

		row, _ := results[keys].(map[string]interface{})
		if nil == row {
			row = make(map[string]interface{})
			results[keys] = row
		}
		row[idx] = vb.Value

		next_oid = vb.Oid
	}
	if 0 == len(results) {
		return internalErrorResult("", "snmp result is empty", nil)
	}

	return commons.Return(results)
}

func (self *SnmpDriver) tableGetByColumns(params map[string]string, client snmpclient.Client,
	oid, columns_str string) commons.Result {

	start_oid, err := snmpclient.ParseOidFromString(oid)
	if nil != err {
		return internalErrorResult("", "param 'oid' is error", err)
	}
	columns, err := commons.ConvertToIntList(columns_str, ",")
	if nil != err {
		return internalErrorResult("", "param 'columns' is error", err)
	}

	version, e := getVersion(params)
	if snmpclient.SNMP_Verr == version {
		return commons.ReturnWithBadRequest(e.Error())
	}

	timeout := getTimeout(params, self.timeout)
	next_oids := make([]snmpclient.SnmpOid, 0, len(columns))
	next_oids_prefix := make([]string, 0, len(columns))
	for _, i := range columns {
		o := start_oid.Concat(i)
		next_oids = append(next_oids, o)
		next_oids_prefix = append(next_oids_prefix, o.GetString()+".")
	}

	results := make(map[string]interface{})
	for {
		var req snmpclient.PDU
		req, err = client.CreatePDU(snmpclient.SNMP_PDU_GETNEXT, version)
		if nil != err {
			return internalErrorResult("", "create pdu failed", err)
		}

		err = req.Init(params)
		if nil != err {
			return internalErrorResult("", "init pdu failed", err)
		}

		//fmt.Println("===============get==================")
		for _, next_oid := range next_oids {
			//fmt.Println(next_oid, ",")
			err = req.GetVariableBindings().AppendWith(next_oid, snmpclient.NewSnmpNil())
			if nil != err {
				return internalErrorResult("", "append vb failed", err)
			}
		}

		resp, err := client.SendAndRecv(req, timeout)
		if nil != err {
			return internalErrorResult(oid, "snmp failed", err)
		}

		if len(next_oids) != resp.GetVariableBindings().Len() {
			return internalErrorResult("", fmt.Sprintf("number of result is mismatch, excepted is %d, actual is %d",
				len(next_oids), resp.GetVariableBindings().Len()), nil)
		}

		//fmt.Println("===============result==================")
		offset := 0
		for i, vb := range resp.GetVariableBindings().All() {
			//fmt.Println(i, ",", vb.Oid, ",", vb.Value.String())
			vb_oid_str := vb.Oid.GetString()
			if !strings.HasPrefix(vb_oid_str, next_oids_prefix[i]) {
				continue
			}

			keys := vb_oid_str[len(next_oids_prefix[i]):]
			if 0 == len(keys) {
				return commons.ReturnError(commons.InternalErrorCode,
					fmt.Sprintf("read '%s' return '%s', it is incorrect - value is %s",
						start_oid.GetString(), vb.Oid.GetString(), vb.Value.String()))
			}

			row, _ := results[keys].(map[string]interface{})
			if nil == row {
				row = make(map[string]interface{})
				results[keys] = row
			}

			row[strconv.FormatInt(int64(columns[i]), 10)] = vb.Value

			next_oids[offset] = vb.Oid
			if offset != i {
				next_oids_prefix[offset] = next_oids_prefix[i]
				columns[offset] = columns[i]
			}

			offset++
		}
		if 0 == offset {
			break
		}
		next_oids_prefix = next_oids_prefix[0:offset]
		columns = columns[0:offset]
		next_oids = next_oids[0:offset]
	}
	if 0 == len(results) {
		return internalErrorResult("", "snmp result is empty", nil)
	}
	return commons.Return(results)
}
