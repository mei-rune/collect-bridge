package metrics

import (
	"commons"
	"commons/as"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type TABLE_CB func(table map[string]interface{}, key string, old_row map[string]interface{}) error
type ONE_CB func(old_row map[string]interface{}) (map[string]interface{}, commons.RuntimeError)

var (
	metricNotExistsError = commons.NewRuntimeError(commons.BadRequestCode, "'metric' is required.")
)

type Base struct {
	drvMgr  *commons.DriverManager
	metrics *commons.DriverManager
	drv     commons.Driver
}

func (self *Base) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	ok := false
	self.drvMgr, ok = params["drvMgr"].(*commons.DriverManager)
	if !ok {
		return commons.NotFound("drvMgr")
	}
	self.metrics, ok = params["metrics"].(*commons.DriverManager)
	if !ok {
		return commons.NotFound("metrics")
	}
	self.drv, _ = self.drvMgr.Connect(drvName)
	if nil == self.drv {
		return commons.NotFound(drvName)
	}
	return nil
}

func (self *Base) GetMetricAsString(params map[string]string, metric string) (string, commons.RuntimeError) {
	if s, ok := params[metric]; ok {
		return s, nil
	}
	drv, _ := self.metrics.Connect(metric)
	if nil == drv {
		return "", metricNotExistsError
	}
	res := drv.Get(params)
	if res.HasError() {
		return "", res.Error()
	}
	s, e := res.Value().AsString()
	if nil != e {
		return s, commons.InternalError(e.Error())
	}
	return s, nil
}

func (self *Base) GetMetricAsInt32(params map[string]string, metric string, defaultValue int32) (int32, commons.RuntimeError) {
	if s, ok := params[metric]; ok {
		i, e := strconv.ParseInt(s, 10, 32)
		if nil != e {
			return int32(i), commons.ValueIsNil
		}
	}
	drv, _ := self.metrics.Connect(metric)
	if nil == drv {
		return defaultValue, metricNotExistsError
	}
	res := drv.Get(params)
	if res.HasError() {
		return defaultValue, res.Error()
	}

	i, e := res.Value().AsInt32()
	if nil != e {
		return defaultValue, commons.InternalError(e.Error())
	}
	return i, nil
}

func (self *Base) GetMetricAsUint32(params map[string]string, metric string, defaultValue uint32) (uint32, commons.RuntimeError) {
	if s, ok := params[metric]; ok {
		i, e := strconv.ParseUint(s, 10, 64)
		if nil != e {
			return uint32(i), commons.ValueIsNil
		}
	}
	drv, _ := self.metrics.Connect(metric)
	if nil == drv {
		return defaultValue, metricNotExistsError
	}
	res := drv.Get(params)
	if res.HasError() {
		return defaultValue, res.Error()
	}

	ui, e := res.Value().AsUint32()
	if nil != e {
		return defaultValue, commons.InternalError(e.Error())
	}
	return ui, nil
}

type SnmpBase struct {
	Base
}

func (self *SnmpBase) GetString(params map[string]string, oid string) commons.Result {
	res, s := self.GetStringValue(params, oid)
	if !res.HasError() {
		return res.Return(s)
	}
	return res
}

func (self *SnmpBase) GetStringValue(params map[string]string, oid string) (commons.Result, string) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res := self.drv.Get(params)
	if res.HasError() {
		return res, ""
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.Return(nil).SetError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), ""
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)), ""
	}
	value, e := TryGetString(params, values, oid)
	if nil == e {
		return res.Return(value), value
	}
	return commons.ReturnError(commons.InternalErrorCode, e.Error()), value
}

func (self *SnmpBase) GetOid(params map[string]string, oid string) commons.Result {
	res, s := self.GetOidValue(params, oid)
	if !res.HasError() {
		return res.Return(s)
	}
	return res
}

func (self *SnmpBase) GetOidValue(params map[string]string, oid string) (commons.Result, string) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res := self.drv.Get(params)
	if res.HasError() {
		return res, ""
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.ReturnError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), ""
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)), ""
	}
	value, e := TryGetOid(params, values, oid)
	if nil == e {
		return res.Return(value), value
	}
	return commons.ReturnError(commons.InternalErrorCode, e.Error()), ""
}

func (self *SnmpBase) GetInt32(params map[string]string, oid string) commons.Result {
	res, i := self.GetInt32Value(params, oid, -1)
	if !res.HasError() {
		return res.Return(i)
	}
	return res
}

func (self *SnmpBase) GetInt32Value(params map[string]string, oid string, defaultValue int32) (commons.Result, int32) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res := self.drv.Get(params)
	if res.HasError() {
		return res, defaultValue
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.Return(defaultValue).SetError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), defaultValue
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)), defaultValue
	}
	value, e := TryGetInt32(params, values, oid, defaultValue)
	if nil == e {
		return res.Return(value), value
	}
	return commons.ReturnError(commons.InternalErrorCode, e.Error()), defaultValue
}

func (self *SnmpBase) GetInt64(params map[string]string, oid string) commons.Result {
	res, i := self.GetInt64Value(params, oid, -1)
	if !res.HasError() {
		return res.Return(i)
	}
	return res
}

func (self *SnmpBase) GetInt64Value(params map[string]string, oid string, defaultValue int64) (commons.Result, int64) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res := self.drv.Get(params)
	if res.HasError() {
		return res, defaultValue
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.Return(defaultValue).SetError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), defaultValue
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)), defaultValue
	}
	value, e := TryGetInt64(params, values, oid, defaultValue)
	if nil == e {
		return res.Return(value), value
	}
	return commons.ReturnError(commons.InternalErrorCode, e.Error()), defaultValue
}

func (self *SnmpBase) GetUint32(params map[string]string, oid string) commons.Result {
	res, i := self.GetUint32Value(params, oid, 0)
	if !res.HasError() {
		return res.Return(i)
	}
	return res
}

func (self *SnmpBase) GetUint32Value(params map[string]string, oid string, defaultValue uint32) (commons.Result, uint32) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res := self.drv.Get(params)
	if res.HasError() {
		return res, defaultValue
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.Return(defaultValue).SetError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), defaultValue
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)), defaultValue
	}
	value, e := TryGetUint32(params, values, oid, defaultValue)
	if nil == e {
		return res.Return(value), value
	}
	return commons.ReturnError(commons.InternalErrorCode, e.Error()), defaultValue
}

func (self *SnmpBase) GetValues(params map[string]string, oids []string) (commons.Result, map[string]interface{}) {
	params["snmp.oid"] = strings.Join(oids, "|")
	params["snmp.action"] = "bulk"
	res := self.drv.Get(params)
	if res.HasError() {
		return res, nil
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.ReturnError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), nil
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)), nil
	}
	return res.Return(values), values
}

func (self *SnmpBase) GetUint64(params map[string]string, oid string) commons.Result {
	res, i := self.GetUint64Value(params, oid, 0)
	if !res.HasError() {
		return res.Return(i)
	}
	return res
}
func (self *SnmpBase) GetUint64Value(params map[string]string, oid string, defaultValue uint64) (commons.Result, uint64) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res := self.drv.Get(params)
	if res.HasError() {
		return res, defaultValue
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.Return(defaultValue).SetError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), defaultValue
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)), defaultValue
	}
	value, e := TryGetUint64(params, values, oid, defaultValue)
	if nil == e {
		return res.Return(value), value
	}
	return commons.ReturnError(commons.InternalErrorCode, e.Error()), defaultValue
}

func (self *SnmpBase) GetTable(params map[string]string, oid, columns string,
	cb TABLE_CB) commons.Result {
	res, t := self.GetTableValue(params, oid, columns, cb)
	if !res.HasError() {
		return res.Return(t)
	}
	return res
}

func (self *SnmpBase) GetTableValue(params map[string]string, oid, columns string,
	cb TABLE_CB) (result commons.Result, sv map[string]interface{}) {

	defer func() {
		if e := recover(); nil != e {
			result = commons.ReturnError(commons.InternalErrorCode, fmt.Sprint(e))
		}
	}()

	params["snmp.oid"] = oid
	params["snmp.action"] = "table"
	params["snmp.columns"] = columns
	res := self.drv.Get(params)
	if res.HasError() {
		return res, nil
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.ReturnError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), nil
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result must is not a map[string]interface{} - [%T]%v.", rv, rv)), nil
	}

	table := map[string]interface{}{}
	for key, r := range values {
		row, ok := r.(map[string]interface{})
		if !ok {
			return commons.ReturnError(commons.InternalErrorCode,
				fmt.Sprintf("row with key is '%s' process failed, it is not a map[string]interface{} - [%T]%v.", key, r, r)), nil
		}
		e := cb(table, key, row)
		if nil != e {
			return commons.ReturnError(commons.InternalErrorCode,
				"row with key is '"+key+"' process failed, "+e.Error()), nil
		}
	}
	return res.Return(table), table
}

func (self *SnmpBase) GetOne(params map[string]string, oid, columns string, cb ONE_CB) commons.Result {
	res, t := self.GetOneValue(params, oid, columns, cb)
	if !res.HasError() {
		return res.Return(t)
	}
	return res
}

func (self *SnmpBase) GetOneValue(params map[string]string, oid, columns string, cb ONE_CB) (result commons.Result, sv map[string]interface{}) {

	defer func() {
		if e := recover(); nil != e {
			result = commons.ReturnError(commons.InternalErrorCode, fmt.Sprint(e))
		}
	}()

	params["snmp.oid"] = oid
	params["snmp.action"] = "table"
	params["snmp.columns"] = columns
	res := self.drv.Get(params)
	if res.HasError() {
		return res, nil
	}
	rv := res.InterfaceValue()
	if nil == rv {
		return commons.ReturnError(commons.ValueIsNil.Code(), commons.ValueIsNil.Error()), nil
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return commons.ReturnError(commons.InternalErrorCode,
			fmt.Sprintf("snmp result is not a map[string]interface{} - [%T]%v.", rv, rv)), nil
	}
	if 0 == len(values) {
		return commons.ReturnError(commons.InternalErrorCode, "result is empty"), nil
	}
	for _, r := range values {
		old_row, ok := r.(map[string]interface{})
		if !ok {
			return commons.ReturnError(commons.InternalErrorCode,
				fmt.Sprintf("result is not a map[string]interface{} - [%T]%v.", r, r)), nil
		}
		row, err := cb(old_row)
		if nil != row || nil != err {
			return res.Return(row), row
		}
	}
	return commons.ReturnError(commons.InternalErrorCode,
		"Record not found - getonevalue"), nil
}

func (self *SnmpBase) Get(params map[string]string) commons.Result {
	return commons.ReturnError(commons.NotImplementedCode, "not implemented")
}

func (self *SnmpBase) Put(params map[string]string) commons.Result {
	return commons.ReturnError(commons.NotImplementedCode, "not implemented")
}

func (self *SnmpBase) Create(params map[string]string) commons.Result {
	return commons.ReturnError(commons.NotImplementedCode, "not implemented")
}

func (self *SnmpBase) Delete(params map[string]string) commons.Result {
	return commons.ReturnError(commons.NotImplementedCode, "not implemented")
}

type systemOid struct {
	SnmpBase
}

func (self *systemOid) Get(params map[string]string) commons.Result {
	return self.GetOid(params, "1.3.6.1.2.1.1.2.0")
}

type systemDescr struct {
	SnmpBase
}

func (self *systemDescr) Get(params map[string]string) commons.Result {
	return self.GetString(params, "1.3.6.1.2.1.1.1.0")
}

type systemName struct {
	SnmpBase
}

func (self *systemName) Get(params map[string]string) commons.Result {
	return self.GetString(params, "1.3.6.1.2.1.1.5.0")
}

type systemUpTime struct {
	SnmpBase
}

func (self *systemUpTime) Get(params map[string]string) commons.Result {
	return self.GetUint32(params, "1.3.6.1.2.1.1.3.0")
}

type systemLocation struct {
	SnmpBase
}

func (self *systemLocation) Get(params map[string]string) commons.Result {
	return self.GetString(params, "1.3.6.1.2.1.1.6.0")
}

type systemServices struct {
	SnmpBase
}

func (self *systemServices) Get(params map[string]string) commons.Result {
	return self.GetInt32(params, "1.3.6.1.2.1.1.7.0")
}

type systemInfo struct {
	SnmpBase
}

func (self *systemInfo) Get(params map[string]string) commons.Result {
	return self.GetOne(params, "1.3.6.1.2.1.1", "", func(old_row map[string]interface{}) (map[string]interface{}, commons.RuntimeError) {

		oid := GetOid(params, old_row, "2")
		services := GetUint32(params, old_row, "7", 0)

		new_row := map[string]interface{}{}
		new_row["sys.descr"] = GetString(params, old_row, "1")
		new_row["sys.oid"] = oid
		new_row["sys.upTime"] = GetUint32(params, old_row, "3", 0)
		new_row["sys.contact"] = GetString(params, old_row, "4")
		new_row["sys.name"] = GetString(params, old_row, "5")
		new_row["sys.location"] = GetString(params, old_row, "6")
		new_row["sys.services"] = services

		params["sys.oid"] = oid
		params["sys.services"] = strconv.Itoa(int(services))

		t, e := self.GetMetricAsString(params, "sys.type")
		if nil == e {
			new_row["sys.type"] = t
		}

		return new_row, nil
	})
}

type interfaceAll struct {
	SnmpBase
}

func (self *interfaceAll) Get(params map[string]string) commons.Result {
	//params["columns"] = "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21"

	return self.GetTable(params, "1.3.6.1.2.1.2.2.1", "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21",
		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32(params, old_row, "1", -1)
			new_row["ifDescr"] = GetString(params, old_row, "2")
			new_row["ifType"] = GetInt32(params, old_row, "3", -1)
			new_row["ifMtu"] = GetInt32(params, old_row, "4", -1)
			new_row["ifSpeed"] = GetUint64(params, old_row, "5", 0)
			new_row["ifPhysAddress"] = GetHardwareAddress(params, old_row, "6")
			new_row["ifAdminStatus"] = GetInt32(params, old_row, "7", -1)
			new_row["ifOpStatus"] = GetInt32(params, old_row, "8", -1)
			new_row["ifLastChange"] = GetInt32(params, old_row, "9", -1)
			new_row["ifInOctets"] = GetUint64(params, old_row, "10", 0)
			new_row["ifInUcastPkts"] = GetUint64(params, old_row, "11", 0)
			new_row["ifInNUcastPkts"] = GetUint64(params, old_row, "12", 0)
			new_row["ifInDiscards"] = GetUint64(params, old_row, "13", 0)
			new_row["ifInErrors"] = GetUint64(params, old_row, "14", 0)
			new_row["ifInUnknownProtos"] = GetUint64(params, old_row, "15", 0)
			new_row["ifOutOctets"] = GetUint64(params, old_row, "16", 0)
			new_row["ifOutUcastPkts"] = GetUint64(params, old_row, "17", 0)
			new_row["ifOutNUcastPkts"] = GetUint64(params, old_row, "18", 0)
			new_row["ifOutDiscards"] = GetUint64(params, old_row, "19", 0)
			new_row["ifOutErrors"] = GetUint64(params, old_row, "20", 0)
			new_row["ifOutQLen"] = GetUint64(params, old_row, "21", 0)
			table[key] = new_row
			return nil
		})
}

type interfaceDescr struct {
	SnmpBase
}

func (self *interfaceDescr) Get(params map[string]string) commons.Result {
	//params["columns"] = "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21"

	return self.GetTable(params, "1.3.6.1.2.1.2.2.1", "1,2,3,4,5,6",
		func(table map[string]interface{}, key string, old_row map[string]interface{}) error {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32(params, old_row, "1", -1)
			new_row["ifDescr"] = GetString(params, old_row, "2")
			new_row["ifType"] = GetInt32(params, old_row, "3", -1)
			new_row["ifMtu"] = GetInt32(params, old_row, "4", -1)
			new_row["ifSpeed"] = GetUint64(params, old_row, "5", 0)
			new_row["ifPhysAddress"] = GetHardwareAddress(params, old_row, "6")
			table[key] = new_row
			return nil
		})
}

type systemType struct {
	SnmpBase
	device2id map[string]int
}

func ErrorIsRestric(msg string, restric bool, log *commons.Logger) commons.RuntimeError {
	if !restric {
		log.DEBUG.Print(msg)
		return nil
	}
	return commons.NewRuntimeError(commons.InternalErrorCode, msg)
}

func (self *systemType) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	e := self.SnmpBase.Init(params, drvName)
	if nil != e {
		return e
	}
	log, ok := params["log"].(*commons.Logger)
	if !ok {
		log = commons.Log
	}

	restric := false
	v, ok := params["restric"]
	if ok {
		restric = as.AsBoolWithDefaultValue(v, restric)
	}

	dt := commons.SearchFile("etc/device_types.json")
	if "" == dt {
		return ErrorIsRestric("'etc/device_types.json' is not exists.", restric, log)
	}

	f, err := ioutil.ReadFile(dt)
	if nil != err {
		return ErrorIsRestric(fmt.Sprintf("read file '%s' failed, %s", dt, err.Error()), restric, log)
	}

	self.device2id = make(map[string]int)
	err = json.Unmarshal(f, &self.device2id)
	if nil != err {
		return ErrorIsRestric(fmt.Sprintf("unmarshal json '%s' failed, %s", dt, err.Error()), restric, log)
	}

	return nil
}

func (self *systemType) Get(params map[string]string) commons.Result {
	oid, e := self.GetMetricAsString(params, "sys.oid")
	if nil == e {
		if dt, ok := self.device2id[oid]; ok {
			return commons.Return(dt)
		}
	}

	//return nil, commons.NotImplemented

	t := 0
	res, dt := self.GetInt32Value(params, "1.3.6.1.2.1.4.1.0", -1)
	if res.HasError() {
		goto SERVICES
	}

	if 1 == dt {
		t += 4
	}
	res, dt = self.GetInt32Value(params, "1.3.6.1.2.1.17.1.2.0", -1)
	if res.HasError() {
		goto SERVICES
	}
	if dt > 0 {
		t += 2
	}

	if 0 != t {
		return commons.Return(t >> 1)
	}
SERVICES:
	services, e := self.GetMetricAsInt32(params, "sys.services", 0)
	if nil != e {
		return commons.ReturnWithError(e)
	}
	return commons.Return((services & 0x7) >> 1)
}

func init() {
	commons.METRIC_DRVS["sys.oid"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemOid{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.descr"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemDescr{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.name"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemName{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.services"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemServices{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.upTime"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemUpTime{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.type"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemType{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.location"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemLocation{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &systemInfo{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["interface"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &interfaceAll{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["interfaceDescr"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &interfaceDescr{}
		return drv, drv.Init(params, "snmp")
	}

}

type DispatchFunc func(params map[string]string) commons.Result

var emptyResult = make(map[string]interface{})

type dispatcherBase struct {
	SnmpBase
	get, set    DispatchFunc
	get_methods map[uint]map[string]DispatchFunc
	set_methods map[uint]map[string]DispatchFunc
}

func splitsystemOid(oid string) (uint, string) {
	if !strings.HasPrefix(oid, "1.3.6.1.4.1.") {
		return 0, oid
	}
	oid = oid[12:]
	idx := strings.IndexRune(oid, '.')
	if -1 == idx {
		u, e := strconv.ParseUint(oid, 10, 0)
		if nil != e {
			panic(e.Error())
		}
		return uint(u), ""
	}

	u, e := strconv.ParseUint(oid[:idx], 10, 0)
	if nil != e {
		panic(e.Error())
	}
	return uint(u), oid[idx+1:]
}

func (self *dispatcherBase) RegisterGetFunc(oids []string, get DispatchFunc) {
	for _, oid := range oids {
		main, sub := splitsystemOid(oid)
		methods := self.get_methods[main]
		if nil == methods {
			methods = map[string]DispatchFunc{}
			self.get_methods[main] = methods
		}
		methods[sub] = get
	}
}

func (self *dispatcherBase) RegisterSetFunc(oids []string, set DispatchFunc) {
	for _, oid := range oids {
		main, sub := splitsystemOid(oid)
		methods := self.set_methods[main]
		if nil == methods {
			methods = map[string]DispatchFunc{}
			self.set_methods[main] = methods
		}
		methods[sub] = set
	}
}

func findFunc(oid string, funcs map[uint]map[string]DispatchFunc) DispatchFunc {
	main, sub := splitsystemOid(oid)
	methods := funcs[main]
	if nil == methods {
		return nil
	}
	get := methods[sub]
	if nil != get {
		return get
	}
	if "" == sub {
		return nil
	}
	return methods[""]
}

func (self *dispatcherBase) FindGetFunc(oid string) DispatchFunc {
	return findFunc(oid, self.get_methods)
}

func (self *dispatcherBase) FindSetFunc(oid string) DispatchFunc {
	return findFunc(oid, self.set_methods)
}

func findDefaultFunc(oid string, funcs map[uint]map[string]DispatchFunc) DispatchFunc {
	main, sub := splitsystemOid(oid)
	methods := funcs[main]
	if nil == methods {
		return nil
	}
	if "" == sub {
		return nil
	}
	return methods[""]
}

func (self *dispatcherBase) FindDefaultGetFunc(oid string) DispatchFunc {
	return findDefaultFunc(oid, self.get_methods)
}

func (self *dispatcherBase) FindDefaultSetFunc(oid string) DispatchFunc {
	return findDefaultFunc(oid, self.set_methods)
}

func (self *dispatcherBase) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	self.get_methods = make(map[uint]map[string]DispatchFunc, 1000)
	self.set_methods = make(map[uint]map[string]DispatchFunc, 1000)
	return self.SnmpBase.Init(params, drvName)
}

func (self *dispatcherBase) invoke(params map[string]string, funcs map[uint]map[string]DispatchFunc) commons.Result {
	oid, e := self.GetMetricAsString(params, "sys.oid")
	if nil != e {
		return commons.ReturnError(e.Code(), "get system oid failed, "+e.Error())
	}
	f := findFunc(oid, funcs)
	if nil != f {
		res := f(params)
		if !res.HasError() {
			return res
		}
		if commons.ContinueCode != res.ErrorCode() {
			return res
		}

		f = findDefaultFunc(oid, funcs)
		if nil != f {
			res := f(params)
			if !res.HasError() {
				return res
			}
			if commons.ContinueCode != res.ErrorCode() {
				return res
			}
		}
	}
	if nil != self.get {
		return self.get(params)
	}
	return commons.ReturnError(commons.NotAcceptableCode, "Unsupported device - "+oid)
}

func (self *dispatcherBase) Get(params map[string]string) commons.Result {
	return self.invoke(params, self.get_methods)
}

func (self *dispatcherBase) Put(params map[string]string) commons.Result {
	return self.invoke(params, self.set_methods)
}
