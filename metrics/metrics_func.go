package metrics

import (
	"commons"
	"commons/as"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"snmp"
	"strconv"
)

type RECORD map[string]interface{}
type TABLE_CB func(table map[string]RECORD, key string, old_row RECORD) error
type ONE_CB func(old_row RECORD) (RECORD, commons.RuntimeError)

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

func (self *Base) GetStringMetric(params map[string]string, metric string) (string, commons.RuntimeError) {
	if s, ok := params[metric]; ok {
		return s, commons.ValueIsNil
	}
	drv, _ := self.metrics.Connect(metric)
	if nil == drv {
		return "", MetricNotExists
	}
	res, err := drv.Get(params)
	if nil == res || nil != err {
		return "", err
	}
	value := commons.GetReturn(res)
	if nil == value {
		return "", commons.ValueIsNil
	}
	s, e := as.AsString(value)
	if nil != e {
		return "", commons.NewRuntimeError(commons.InternalErrorCode, e.Error())
	}
	return s, nil
}

func (self *Base) GetInt32Metric(params map[string]string, metric string, defaultValue int32) (int32, commons.RuntimeError) {
	if s, ok := params[metric]; ok {
		i, e := strconv.ParseInt(s, 10, 32)
		if nil != e {
			return int32(i), commons.ValueIsNil
		}
	}
	drv, _ := self.metrics.Connect(metric)
	if nil == drv {
		return defaultValue, MetricNotExists
	}
	res, err := drv.Get(params)
	if nil == res || nil != err {
		return defaultValue, err
	}
	value := commons.GetReturn(res)
	if nil == value {
		return defaultValue, commons.ValueIsNil
	}
	i, e := as.AsInt32(value)
	if nil != e {
		return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, e.Error())
	}
	return i, nil
}

func (self *Base) GetUint32Metric(params map[string]string, metric string, defaultValue uint32) (uint32, commons.RuntimeError) {
	if s, ok := params[metric]; ok {
		i, e := strconv.ParseUint(s, 10, 64)
		if nil != e {
			return uint32(i), commons.ValueIsNil
		}
	}
	drv, _ := self.metrics.Connect(metric)
	if nil == drv {
		return defaultValue, MetricNotExists
	}
	res, err := drv.Get(params)
	if nil == res || nil != err {
		return defaultValue, err
	}
	value := commons.GetReturn(res)
	if nil == value {
		return defaultValue, commons.ValueIsNil
	}
	i, e := as.AsUint32(value)
	if nil != e {
		return defaultValue, commons.NewRuntimeError(commons.InternalErrorCode, e.Error())
	}
	return i, nil
}

func (self *Base) GetString(params map[string]string, oid string) (map[string]interface{}, commons.RuntimeError) {
	res, s, err := self.GetStringValue(params, oid)
	if nil == err {
		return commons.ReturnWithValue(res, s), err
	}
	return res, err
}

func (self *Base) GetStringValue(params map[string]string, oid string) (map[string]interface{}, string, commons.RuntimeError) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, "", err
	}
	value, err := GetString(params, res, oid)
	return res, value, err
}

func (self *Base) GetOid(params map[string]string, oid string) (map[string]interface{}, commons.RuntimeError) {
	res, s, err := self.GetOidValue(params, oid)
	if nil == err {
		return commons.ReturnWithValue(res, s), err
	}
	return res, err
}

func (self *Base) GetOidValue(params map[string]string, oid string) (map[string]interface{}, string, commons.RuntimeError) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, "", err
	}
	value, err := GetOid(params, res, oid)
	return res, value, err
}

func (self *Base) GetInt32(params map[string]string, oid string) (map[string]interface{}, commons.RuntimeError) {
	res, i, err := self.GetInt32Value(params, oid, -1)
	if nil == err {
		return commons.ReturnWithValue(res, i), err
	}
	return res, err
}

func (self *Base) GetInt32Value(params map[string]string, oid string, defaultValue int32) (map[string]interface{}, int32, commons.RuntimeError) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, defaultValue, err
	}
	value, err := GetInt32(params, res, oid, defaultValue)
	return res, value, err
}

func (self *Base) GetUint32(params map[string]string, oid string) (map[string]interface{}, commons.RuntimeError) {
	res, i, err := self.GetUint32Value(params, oid, 0)
	if nil == err {
		return commons.ReturnWithValue(res, i), err
	}
	return res, err
}

func (self *Base) GetUint32Value(params map[string]string, oid string, defaultValue uint32) (map[string]interface{}, uint32, commons.RuntimeError) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "get"
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, defaultValue, err
	}
	value, err := GetUint32(params, res, oid, defaultValue)
	return res, value, err
}

func (self *Base) GetTable(params map[string]string, oid, columns string, cb TABLE_CB) (map[string]interface{}, commons.RuntimeError) {
	res, t, err := self.GetTableValue(params, oid, columns, cb)
	if nil == err {
		return commons.ReturnWithValue(res, t), err
	}
	return res, err
}

func (self *Base) GetTableValue(params map[string]string, oid, columns string, cb TABLE_CB) (map[string]interface{}, map[string]RECORD, commons.RuntimeError) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "table"
	params["snmp.columns"] = columns
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, nil, err
	}
	rv := commons.GetReturn(res)
	if nil == rv {
		return res, nil, err
	}
	switch values := rv.(type) {
	case map[string]map[string]snmp.SnmpValue:
		table := map[string]RECORD{}
		for key, r := range values {
			row := RECORD{}
			for k, v := range r {
				row[k] = v
			}
			e := cb(table, key, row)
			if nil != e {
				return res, nil, commons.NewRuntimeError(commons.InternalErrorCode,
					"row with key is '"+key+"' process failed, "+e.Error())
			}
		}
		return res, table, err
	case map[string]interface{}:
		table := map[string]RECORD{}
		for key, r := range values {
			row, ok := r.(RECORD)
			if !ok {
				return res, nil, commons.NewRuntimeError(commons.InternalErrorCode,
					"row with key is '"+key+"' process failed, it is not a RECORD")
			}
			e := cb(table, key, row)
			if nil != e {
				return res, nil, commons.NewRuntimeError(commons.InternalErrorCode,
					"row with key is '"+key+"' process failed, "+e.Error())
			}
		}
		return res, table, err
	}
	return res, nil, commons.NewRuntimeError(commons.InternalErrorCode,
		fmt.Sprintf("snmp result must is not a map[string]snmp.SnmpValue or map[string]interface{} - [%T]%v.", rv, rv))
}

func (self *Base) GetOne(params map[string]string, oid, columns string, cb ONE_CB) (map[string]interface{}, commons.RuntimeError) {
	res, t, err := self.GetOneValue(params, oid, columns, cb)
	if nil == err {
		return commons.ReturnWithValue(res, t), err
	}
	return res, err
}

func (self *Base) GetOneValue(params map[string]string, oid, columns string, cb ONE_CB) (map[string]interface{}, RECORD, commons.RuntimeError) {
	params["snmp.oid"] = oid
	params["snmp.action"] = "table"
	params["snmp.columns"] = columns
	res, err := self.drv.Get(params)
	if nil == res || nil != err {
		return res, nil, err
	}
	rv := commons.GetReturn(res)
	if nil == rv {
		return res, nil, err
	}
	switch values := rv.(type) {
	case map[string]map[string]snmp.SnmpValue:
		if 0 == len(values) {
			return res, nil, commons.NewRuntimeError(commons.InternalErrorCode, "result is empty")
		}
		for _, r := range values {
			old_row := RECORD{}
			for k, v := range r {
				old_row[k] = v
			}

			row, err := cb(old_row)
			return res, row, err
		}
	case map[string]interface{}:
		if 0 == len(values) {
			return res, nil, commons.NewRuntimeError(commons.InternalErrorCode, "result is empty")
		}
		for _, r := range values {
			old_row, ok := r.(map[string]interface{})
			if !ok {
				return res, nil, commons.NewRuntimeError(commons.InternalErrorCode,
					fmt.Sprintf("result is not a map[string]interface{} - [%T]%v.", r, r))
			}
			row, err := cb(old_row)
			return res, row, err
		}
	}
	return res, nil, commons.NewRuntimeError(commons.InternalErrorCode,
		fmt.Sprintf("snmp result must is not a map[string]map[string]snmp.SnmpValue or map[string]interface{} - [%T]%v.", rv, rv))
}

func (self *Base) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Create(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return nil, commons.NotImplemented
}

func (self *Base) Delete(params map[string]string) (bool, commons.RuntimeError) {
	return false, commons.NotImplemented
}

type SystemOid struct {
	Base
}

func (self *SystemOid) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return self.GetOid(params, "1.3.6.1.2.1.1.2.0")
}

type SystemDescr struct {
	Base
}

func (self *SystemDescr) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return self.GetString(params, "1.3.6.1.2.1.1.1.0")
}

type SystemName struct {
	Base
}

func (self *SystemName) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return self.GetString(params, "1.3.6.1.2.1.1.5.0")
}

type SystemUpTime struct {
	Base
}

func (self *SystemUpTime) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return self.GetUint32(params, "1.3.6.1.2.1.1.3.0")
}

type SystemServices struct {
	Base
}

func (self *SystemServices) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return self.GetInt32(params, "1.3.6.1.2.1.1.7.0")
}

type SystemInfo struct {
	Base
}

func (self *SystemInfo) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return self.GetOne(params, "1.3.6.1.2.1.1", "", func(old_row RECORD) (RECORD, commons.RuntimeError) {
		new_row := RECORD{}
		new_row["sys.descr"] = GetStringColumn(params, old_row, "1")
		new_row["sys.oid"] = GetOidColumn(params, old_row, "2")
		new_row["sys.upTime"] = GetUint32Column(params, old_row, "3", 0)
		new_row["sys.contact"] = GetStringColumn(params, old_row, "4")
		new_row["sys.name"] = GetStringColumn(params, old_row, "5")
		new_row["sys.location"] = GetStringColumn(params, old_row, "6")
		new_row["sys.services"] = GetUint32Column(params, old_row, "7", 0)
		return new_row, nil
	})
}

type Interface struct {
	Base
}

func (self *Interface) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	//params["columns"] = "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21"

	return self.GetTable(params, "1.3.6.1.2.1.2.2.1", "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21",
		func(table map[string]RECORD, key string, old_row RECORD) error {
			new_row := RECORD{}
			new_row["ifIndex"] = GetInt32Column(params, old_row, "1", -1)
			new_row["ifDescr"] = GetStringColumn(params, old_row, "2")
			new_row["ifType"] = GetInt32Column(params, old_row, "3", -1)
			new_row["ifMtu"] = GetInt32Column(params, old_row, "4", -1)
			new_row["ifSpeed"] = GetUint64Column(params, old_row, "5", 0)
			new_row["ifPhysAddress"] = GetHardwareAddressColumn(params, old_row, "6")
			new_row["ifAdminStatus"] = GetInt32Column(params, old_row, "7", -1)
			new_row["ifOpStatus"] = GetInt32Column(params, old_row, "8", -1)
			new_row["ifLastChange"] = GetInt32Column(params, old_row, "9", -1)
			new_row["ifInOctets"] = GetUint64Column(params, old_row, "10", 0)
			new_row["ifInUcastPkts"] = GetUint64Column(params, old_row, "11", 0)
			new_row["ifInNUcastPkts"] = GetUint64Column(params, old_row, "12", 0)
			new_row["ifInDiscards"] = GetUint64Column(params, old_row, "13", 0)
			new_row["ifInErrors"] = GetUint64Column(params, old_row, "14", 0)
			new_row["ifInUnknownProtos"] = GetUint64Column(params, old_row, "15", 0)
			new_row["ifOutOctets"] = GetUint64Column(params, old_row, "16", 0)
			new_row["ifOutUcastPkts"] = GetUint64Column(params, old_row, "17", 0)
			new_row["ifOutNUcastPkts"] = GetUint64Column(params, old_row, "18", 0)
			new_row["ifOutDiscards"] = GetUint64Column(params, old_row, "19", 0)
			new_row["ifOutErrors"] = GetUint64Column(params, old_row, "20", 0)
			new_row["ifOutQLen"] = GetUint64Column(params, old_row, "21", 0)
			table[key] = new_row
			return nil
		})
}

type SystemType struct {
	Base
	device2id map[string]int
}

func ErrorIsRestric(msg string, restric bool, log *commons.Logger) commons.RuntimeError {
	if !restric {
		log.DEBUG.Print(msg)
		return nil
	}
	return commons.NewRuntimeError(commons.InternalErrorCode, msg)
}

func (self *SystemType) Init(params map[string]interface{}, drvName string) commons.RuntimeError {
	e := self.Base.Init(params, drvName)
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

func (self *SystemType) Get(params map[string]string) (
	map[string]interface{}, commons.RuntimeError) {
	oid, e := self.GetStringMetric(params, "sys.oid")
	if nil == e {
		if dt, ok := self.device2id[oid]; ok {
			return commons.Return(dt), nil
		}
	}

	//return nil, commons.NotImplemented

	t := 0
	_, dt, e := self.GetInt32Value(params, "1.3.6.1.2.1.4.1.0", -1)
	if nil != e {
		goto SERVICES
	}

	if 1 == dt {
		t += 4
	}
	_, dt, e = self.GetInt32Value(params, "1.3.6.1.2.1.17.1.2.0", -1)
	if nil != e {
		goto SERVICES
	}
	if dt > 0 {
		t += 2
	}

	if 0 != t {
		return commons.Return(t >> 1), nil
	}
SERVICES:
	services, e := self.GetInt32Metric(params, "sys.services", 0)
	if nil != e {
		return nil, e
	}
	return commons.Return((services & 0x7) >> 1), nil
}

func init() {
	commons.METRIC_DRVS["sys.oid"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &SystemOid{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.descr"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &SystemDescr{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.name"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &SystemName{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.services"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &SystemServices{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.upTime"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &SystemUpTime{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys.type"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &SystemType{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["sys"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &SystemInfo{}
		return drv, drv.Init(params, "snmp")
	}
	commons.METRIC_DRVS["interface"] = func(params map[string]interface{}) (commons.Driver, commons.RuntimeError) {
		drv := &Interface{}
		return drv, drv.Init(params, "snmp")
	}
}
