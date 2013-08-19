package sampling

import (
	"commons"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type result_type int

const (
	RES_STRING result_type = iota
	RES_OID
	RES_INT32
	RES_INT64
	RES_UINT32
	RES_UINT64
)

var (
	metricNotExistsError = commons.IsRequired("metric")
	snmpNotExistsError   = commons.IsRequired("snmp")
)

type snmpBase struct {
	drv commons.Driver
}

func (self *snmpBase) CopyFrom(from *snmpBase) {
	self.drv = from.drv
}

func (self *snmpBase) Init(params map[string]interface{}) error {
	v := params["snmp"]
	if nil != v {
		if drv, ok := v.(commons.Driver); ok {
			self.drv = drv
			return nil
		}
	}

	v = params["drv_manager"]
	if nil == v {
		return commons.IsRequired("snmp' or 'drv_manager'")
	}

	drvMgr, ok := v.(commons.DriverManager)
	if !ok {
		return errors.New("'drv_manager' is not a driver manager.")
	}

	drv, _ := drvMgr.Connect("snmp")
	if nil == v {
		return errors.New("'snmp' is not exists in the driver manager")
	}
	self.drv = drv
	return nil
}

func (self *snmpBase) copyParameter(params MContext, snmp_params map[string]string, rw string) error {
	version := params.GetStringWithDefault("snmp.version", "")
	if 0 == len(version) {
		return snmpNotExistsError
	}

	snmp_params["snmp.version"] = version

	address := params.GetStringWithDefault("@address", "")
	if 0 == len(address) {
		return commons.IsRequired("@address")
	}
	snmp_params["snmp.address"] = address
	snmp_params["snmp.port"] = params.GetStringWithDefault("snmp.port", "")
	timeout := params.GetStringWithDefault("snmp.timeout", "")
	if 0 == len(timeout) {
		timeout = params.GetStringWithDefault("timeout", "")
		if 0 != len(timeout) {
			snmp_params["snmp.timeout"] = timeout
		}
	} else {
		snmp_params["snmp.timeout"] = timeout
	}

	switch version {
	case "v3", "V3", "3":
		snmp_params["snmp.secmodel"] = params.GetStringWithDefault("snmp.sec_model", "")
		snmp_params["snmp.auth_pass"] = params.GetStringWithDefault("snmp."+rw+"_auth_pass", "")
		snmp_params["snmp.priv_pass"] = params.GetStringWithDefault("snmp."+rw+"_priv_pass", "")
		snmp_params["snmp.max_msg_size"] = params.GetStringWithDefault("snmp.max_msg_size", "")
		snmp_params["snmp.context_name"] = params.GetStringWithDefault("snmp.context_name", "")
		snmp_params["snmp.identifier"] = params.GetStringWithDefault("snmp.identifier", "")
		snmp_params["snmp.engine_id"] = params.GetStringWithDefault("snmp.engine_id", "")
		break
	default:
		community := params.GetStringWithDefault("snmp."+rw+"_community", "")
		if 0 == len(community) {
			return commons.IsRequired("snmp." + rw + "_community")
		}

		snmp_params["snmp.community"] = community
	}
	return nil
}
func (self *snmpBase) GetNext(params MContext, oid string) (map[string]interface{}, error) {
	return self.Read(params, "next", oid)
}

func (self *snmpBase) Get(params MContext, oid string) (map[string]interface{}, error) {
	return self.Read(params, "get", oid)
}

func (self *snmpBase) GetBulk(params MContext, oid string) (map[string]interface{}, error) {
	return self.Read(params, "bulk", oid)
}

func (self *snmpBase) Write(params MContext, oid string, value interface{}) error {
	snmp_params := make(map[string]string)
	snmp_params["snmp.oid"] = oid
	snmp_params["snmp.action"] = "set"
	e := self.copyParameter(params, snmp_params, "write")
	if nil != e {
		return e
	}
	res := self.drv.Put(snmp_params, value)
	if res.HasError() {
		return res.Error()
	}

	return nil
}

func (self *snmpBase) Read(params MContext, action, oid string) (map[string]interface{}, error) {
	snmp_params := make(map[string]string)
	snmp_params["snmp.oid"] = oid
	snmp_params["snmp.action"] = action
	e := self.copyParameter(params, snmp_params, "read")
	if nil != e {
		return nil, e
	}
	res := self.drv.Get(snmp_params)
	if res.HasError() {
		return nil, res.Error()
	}

	rv := res.InterfaceValue()
	if nil == rv {
		return nil, commons.ValueIsNil
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("snmp result is not a map[string]interface{}, actual is [%T]%v.", rv, rv)
	}
	return values, nil
}

func (self *snmpBase) GetResult(params MContext, oid string, rt result_type) commons.Result {
	return self.ReadResult(params, "get", oid, rt)
}

func (self *snmpBase) ReadResult(params MContext, action, oid string, rt result_type) commons.Result {
	values, e := self.Read(params, action, oid)
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}

	switch rt {
	case RES_STRING:
		s, e := TryGetString(params, values, oid)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}
		return commons.Return(s)
	case RES_OID:
		s, e := TryGetOid(params, values, oid)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}
		return commons.Return(s)
	case RES_INT32:
		i32, e := TryGetInt32(params, values, oid, 0)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}
		return commons.Return(i32)
	case RES_INT64:
		i64, e := TryGetInt64(params, values, oid, 0)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}
		return commons.Return(i64)
	case RES_UINT32:
		u32, e := TryGetUint32(params, values, oid, 0)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}
		return commons.Return(u32)
	case RES_UINT64:
		u64, e := TryGetInt64(params, values, oid, 0)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}
		return commons.Return(u64)
	default:
		return commons.ReturnWithInternalError("unsupported type of snmp result - " + strconv.Itoa(int(rt)))
	}
}

func (self *snmpBase) GetString(params MContext, oid string) (string, error) {
	values, e := self.Get(params, oid)
	if nil != e {
		return "", e
	}
	return TryGetString(params, values, oid)
}

func (self *snmpBase) GetOid(params MContext, oid string) (string, error) {
	values, e := self.Get(params, oid)
	if nil != e {
		return "", e
	}
	return TryGetOid(params, values, oid)
}

func (self *snmpBase) GetInt32(params MContext, oid string) (int32, error) {
	values, e := self.Get(params, oid)
	if nil != e {
		return 0, e
	}
	return TryGetInt32(params, values, oid, 0)
}

func (self *snmpBase) GetInt64(params MContext, oid string) (int64, error) {
	values, e := self.Get(params, oid)
	if nil != e {
		return 0, e
	}
	return TryGetInt64(params, values, oid, 0)
}

func (self *snmpBase) GetUint32(params MContext, oid string) (uint32, error) {
	values, e := self.Get(params, oid)
	if nil != e {
		return 0, e
	}
	return TryGetUint32(params, values, oid, 0)
}

func (self *snmpBase) GetUint64(params MContext, oid string) (uint64, error) {
	values, e := self.Get(params, oid)
	if nil != e {
		return 0, e
	}
	return TryGetUint64(params, values, oid, 0)
}

func (self *snmpBase) GetTable(params MContext, oid, columns string,
	cb func(key string, row map[string]interface{}) error) (e error) {
	defer func() {
		if o := recover(); nil != o {
			e = errors.New(fmt.Sprint(o))
		}
	}()

	snmp_params := map[string]string{"snmp.oid": oid,
		"snmp.action":  "table",
		"snmp.columns": columns}

	e = self.copyParameter(params, snmp_params, "read")
	if nil != e {
		return e
	}

	res := self.drv.Get(snmp_params)
	if res.HasError() {
		return res.Error()
	}

	rv := res.InterfaceValue()
	if nil == rv {
		return commons.ValueIsNil
	}
	values, ok := rv.(map[string]interface{})
	if !ok {
		return fmt.Errorf("snmp result must is not a map[string]interface{} - [%T]%v.", rv, rv)
	}

	for key, r := range values {
		row, ok := r.(map[string]interface{})
		if !ok {
			return fmt.Errorf("row with key is '%s' process failed, it is not a map[string]interface{} - [%T]%v.", key, r, r)
		}

		e := cb(key, row)
		if nil != e {
			if commons.InterruptError == e {
				break
			}

			return errors.New("row with key is '" + key + "' process failed, " + e.Error())
		}
	}
	return nil
}

func (self *snmpBase) OneInTable(params MContext, oid, columns string,
	cb func(key string, row map[string]interface{}) error) error {
	return self.GetTable(params, oid, columns, func(key string, row map[string]interface{}) error {
		e := cb(key, row)
		if nil == e {
			return commons.InterruptError
		}

		if commons.ContinueError == e {
			return nil
		}

		return e
	})
}

func (self *snmpBase) EachInTable(params MContext, oid, columns string,
	cb func(key string, row map[string]interface{}) error) error {
	return self.GetTable(params, oid, columns, cb)
}

func (self *snmpBase) GetOneResult(params MContext, oid, columns string,
	cb func(key string, row map[string]interface{}) (map[string]interface{}, error)) commons.Result {
	var err error
	var result map[string]interface{} = nil
	err = self.GetTable(params, oid, columns, func(key string, row map[string]interface{}) error {
		var e error
		result, e = cb(key, row)
		if nil == e {
			return commons.InterruptError
		}

		if commons.ContinueError == e {
			return nil
		}

		return e
	})
	if nil != err {
		return commons.ReturnWithInternalError(err.Error())
	}
	return commons.Return(result)
}

func (self *snmpBase) GetAllResult(params MContext, oid, columns string,
	cb func(key string, row map[string]interface{}) (map[string]interface{}, error)) commons.Result {
	var err error
	var results []map[string]interface{} = nil
	err = self.GetTable(params, oid, columns, func(key string, row map[string]interface{}) error {
		result, e := cb(key, row)
		if nil != e {
			return e
		}
		results = append(results, result)
		return nil
	})
	if nil != err {
		return commons.ReturnWithInternalError(err.Error())
	}
	return commons.Return(results)
}

type systemOid struct {
	snmpBase
}

func (self *systemOid) Call(params MContext) commons.Result {
	return self.GetResult(params, "1.3.6.1.2.1.1.2.0", RES_OID)
}

type systemDescr struct {
	snmpBase
}

func (self *systemDescr) Call(params MContext) commons.Result {
	return self.GetResult(params, "1.3.6.1.2.1.1.1.0", RES_STRING)
}

type systemName struct {
	snmpBase
}

func (self *systemName) Call(params MContext) commons.Result {
	return self.GetResult(params, "1.3.6.1.2.1.1.5.0", RES_STRING)
}

type systemUpTime struct {
	snmpBase
}

func (self *systemUpTime) Call(params MContext) commons.Result {
	t, e := self.GetUint64(params, "1.3.6.1.2.1.1.3.0")
	if nil != e {
		return commons.Return(0).SetError(commons.InternalErrorCode, e.Error())
	}
	return commons.Return(t / 100)
}

type systemLocation struct {
	snmpBase
}

func (self *systemLocation) Call(params MContext) commons.Result {
	return self.GetResult(params, "1.3.6.1.2.1.1.6.0", RES_STRING)
}

type systemServices struct {
	snmpBase
}

func (self *systemServices) Call(params MContext) commons.Result {
	return self.GetResult(params, "1.3.6.1.2.1.1.7.0", RES_INT64)
}

type systemInfo struct {
	snmpBase
}

func (self *systemInfo) Call(params MContext) commons.Result {
	return self.GetOneResult(params, "1.3.6.1.2.1.1", "",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			oid := GetOid(params, old_row, "2")
			services := GetUint32(params, old_row, "7", 0)

			new_row := map[string]interface{}{}
			new_row["descr"] = GetString(params, old_row, "1")
			new_row["oid"] = oid
			new_row["upTime"] = GetUint32(params, old_row, "3", 0) / 100
			new_row["contact"] = GetString(params, old_row, "4")
			new_row["name"] = GetString(params, old_row, "5")
			new_row["location"] = GetString(params, old_row, "6")
			new_row["services"] = services

			params.Set("&sys.oid", oid)
			params.Set("&sys.services", strconv.Itoa(int(services)))
			new_row["type"] = params.GetUintWithDefault("!sys.type", 0)
			return new_row, nil
		})
}

type interfaceAll struct {
	snmpBase
}

func (self *interfaceAll) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
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
			new_row["ifSpecific"] = GetOid(params, old_row, "22")
			return new_row, nil
		})
}

type interfaceStatus struct {
	snmpBase
}

func (self *interfaceStatus) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,7,8",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32(params, old_row, "1", -1)
			new_row["ifAdminStatus"] = GetInt32(params, old_row, "7", -1)
			new_row["ifOpStatus"] = GetInt32(params, old_row, "8", -1)
			return new_row, nil
		})
}

type interfaceScalar struct {
	snmpBase
}

func (self *interfaceScalar) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,10,11,12,13,14,15,16,17,18,19,20",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32(params, old_row, "1", -1)
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
			return new_row, nil
		})
}

type interfaceDescr struct {
	snmpBase
}

func (self *interfaceDescr) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,2,3,4,5,6,22",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32(params, old_row, "1", -1)
			new_row["ifDescr"] = GetString(params, old_row, "2")
			new_row["ifType"] = GetInt32(params, old_row, "3", -1)
			new_row["ifMtu"] = GetInt32(params, old_row, "4", -1)
			new_row["ifSpeed"] = GetUint64(params, old_row, "5", 0)
			new_row["ifPhysAddress"] = GetHardwareAddress(params, old_row, "6")
			new_row["ifSpecific"] = GetOid(params, old_row, "22")
			return new_row, nil
		})
}

type tcpConnection struct {
	snmpBase
}

var tcpConnectionState = []string{
	"closed",
	"listen",
	"synSent",
	"synReceived",
	"established",
	"finWait1",
	"finWait2",
	"closeWait",
	"lastAck",
	"closing",
	"timeWait",
	"deleteTCB"}

func tcpConnectionStateString(i int32) string {
	if i < 1 || i > 12 {
		return "unknow status"
	}

	return tcpConnectionState[i-1]
}

func (self *tcpConnection) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.6.13.1", "1,2,3,4,5",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["tcpConnState"] = GetInt32(params, old_row, "1", -1)
			new_row["tcpConnStateString"] = tcpConnectionStateString(GetInt32(params, old_row, "1", -1))
			new_row["tcpConnLocalAddress"] = GetIPAddress(params, old_row, "2")
			new_row["tcpConnLocalPort"] = GetInt32(params, old_row, "3", 0)
			new_row["tcpConnRemAddress"] = GetIPAddress(params, old_row, "4")
			new_row["tcpConnRemPort"] = GetInt32(params, old_row, "5", 0)
			return new_row, nil
		})
}

type udpListen struct {
	snmpBase
}

func (self *udpListen) Call(params MContext) commons.Result {
	return self.GetAllResult(params, "1.3.6.1.2.1.6.13.1", "1,2,3,4,5",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["udpLocalAddress"] = GetIPAddress(params, old_row, "1")
			new_row["udpLocalPort"] = GetInt32(params, old_row, "2", 0)
			return new_row, nil
		})
}

type systemType struct {
	snmpBase
	device2id map[string]int
}

func ErrorIsRestric(msg string, restric bool, log *commons.Logger) error {
	if !restric {
		log.DEBUG.Print(msg)
		return nil
	}
	return errors.New(msg)
}

// func (self *systemType) Init(params map[string]interface{}, drvName string) error {
// 	e := self.snmpBase.Init(params, drvName)
// 	if nil != e {
// 		return e
// 	}
// 	log, ok := params["log"].(*commons.Logger)
// 	if !ok {
// 		log = commons.Log
// 	}

// 	restric := false
// 	v, ok := params["restric"]
// 	if ok {
// 		restric = commons.AsBoolWithDefaultValue(v, restric)
// 	}

// 	dt := commons.SearchFile("etc/device_types.json")
// 	if "" == dt {
// 		return ErrorIsRestric("'etc/device_types.json' is not exists.", restric, log)
// 	}

// 	f, err := ioutil.ReadFile(dt)
// 	if nil != err {
// 		return ErrorIsRestric(fmt.Sprintf("read file '%s' failed, %s", dt, err.Error()), restric, log)
// 	}

// 	self.device2id = make(map[string]int)
// 	err = json.Unmarshal(f, &self.device2id)
// 	if nil != err {
// 		return ErrorIsRestric(fmt.Sprintf("unmarshal json '%s' failed, %s", dt, err.Error()), restric, log)
// 	}

// 	return nil
// }

func (self *systemType) Call(params MContext) commons.Result {
	if nil != self.device2id {
		oid := params.GetStringWithDefault("&sys.oid", "")
		if 0 != len(oid) {
			if dt, ok := self.device2id[oid]; ok {
				return commons.Return(dt)
			}
		}
	}

	t := 0
	dt, e := self.GetInt32(params, "1.3.6.1.2.1.4.1.0")
	if nil != e {
		goto SERVICES
	}

	if 1 == dt {
		t += 4
	}
	dt, e = self.GetInt32(params, "1.3.6.1.2.1.17.1.2.0")
	if nil != e {
		goto SERVICES
	}
	if dt > 0 {
		t += 2
	}

	if 0 != t {
		return commons.Return(t >> 1)
	}
SERVICES:
	services, e := params.GetUint32("&sys.services")
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}
	return commons.Return((services & 0x7) >> 1)
}

type snmpRead struct {
	snmpBase
	action string
}

func (self *snmpRead) Call(params MContext) commons.Result {
	address := params.GetStringWithDefault("@address", "")
	if 0 == len(address) {
		return commons.ReturnWithIsRequired("@address")
	}
	idx := strings.IndexRune(address, ',')
	if -1 != idx {
		params.Set("@address", address[0:idx])
		params.Set("snmp.port", address[idx+1:])
	}

	oid, e := params.GetString("snmp.oid")
	if nil != e {
		return commons.ReturnWithIsRequired("snmp.oid")
	}
	if 0 == len(oid) {
		return commons.ReturnWithBadRequest("'snmp.oid' is empty.")
	}

	typ := params.GetStringWithDefault("snmp.type", "")
	switch typ {
	case "string":
		return self.ReadResult(params, self.action, oid, RES_STRING)
	case "oid":
		return self.ReadResult(params, self.action, oid, RES_OID)
	case "int32":
		return self.ReadResult(params, self.action, oid, RES_INT32)
	case "int64":
		return self.ReadResult(params, self.action, oid, RES_INT64)
	case "uint32":
		return self.ReadResult(params, self.action, oid, RES_UINT32)
	case "uint64":
		return self.ReadResult(params, self.action, oid, RES_UINT64)
	default:
		values, e := self.Read(params, self.action, oid)
		if nil != e {
			return commons.ReturnWithInternalError(e.Error())
		}
		return commons.Return(values)
	}
}

type snmpWrite struct {
	snmpBase
}

func (self *snmpWrite) Call(params MContext) commons.Result {
	address := params.GetStringWithDefault("@address", "")
	if 0 == len(address) {
		return commons.ReturnWithIsRequired("@address")
	}
	idx := strings.IndexRune(address, ',')
	if -1 != idx {
		params.Set("@address", address[0:idx])
		params.Set("snmp.port", address[idx+1:])
	}

	oid, e := params.GetString("snmp.oid")
	if nil != e {
		return commons.ReturnWithIsRequired("snmp.oid")
	}
	if 0 == len(oid) {
		return commons.ReturnWithBadRequest("'snmp.oid' is empty.")
	}

	body, e := params.Body()
	if nil != e {
		return commons.ReturnWithBadRequest("read body failed, " + e.Error())
	}
	if nil == body {
		return commons.ReturnWithBadRequest("'body' is nil.")
	}

	e = self.Write(params, oid, body)
	if nil != e {
		return commons.ReturnWithInternalError(e.Error())
	}
	return commons.Return(true)
}

type snmpTable struct {
	snmpBase
}

func (self *snmpTable) Call(params MContext) commons.Result {
	address := params.GetStringWithDefault("@address", "")
	if 0 == len(address) {
		return commons.ReturnWithIsRequired("@address")
	}
	idx := strings.IndexRune(address, ',')
	if -1 != idx {
		params.Set("@address", address[0:idx])
		params.Set("snmp.port", address[idx+1:])
	}

	oid, e := params.GetString("snmp.oid")
	if nil != e {
		return commons.ReturnWithIsRequired("snmp.oid")
	}
	if 0 == len(oid) {
		return commons.ReturnWithBadRequest("'snmp.oid' is empty.")
	}

	return self.GetAllResult(params, oid, params.GetStringWithDefault("snmp.columns", ""),
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			return old_row, nil
		})
}

func init() {

	Methods["snmp_get"] = newRouteSpec("get", "snmp_get", "get a snmp value", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &snmpRead{action: "get"}
			return drv, drv.Init(params)
		})
	Methods["snmp_next"] = newRouteSpec("get", "snmp_next", "get a next snmp value", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &snmpRead{action: "next"}
			return drv, drv.Init(params)
		})

	Methods["snmp_bulk"] = newRouteSpec("get", "snmp_bulk", "get some snmp value", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &snmpRead{action: "bulk"}
			return drv, drv.Init(params)
		})

	Methods["snmp_table"] = newRouteSpec("get", "snmp_table", "get a snmp table", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &snmpTable{}
			return drv, drv.Init(params)
		})

	Methods["snmp_set"] = newRouteSpec("put", "snmp_set", "set a snmp value", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &snmpWrite{}
			return drv, drv.Init(params)
		})

	Methods["sys_oid"] = newRouteSpec("get", "sys.oid", "the oid of system", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemOid{}
			return drv, drv.Init(params)
		})

	Methods["sys_descr"] = newRouteSpec("get", "sys.descr", "the oid of system", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemDescr{}
			return drv, drv.Init(params)
		})

	Methods["sys_name"] = newRouteSpec("get", "sys.name", "the name of system", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemName{}
			return drv, drv.Init(params)
		})

	Methods["sys_services"] = newRouteSpec("get", "sys.services", "the name of system", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemServices{}
			return drv, drv.Init(params)
		})

	Methods["sys_upTime"] = newRouteSpec("get", "sys.upTime", "the upTime of system", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemUpTime{}
			return drv, drv.Init(params)
		})

	Methods["sys_type"] = newRouteSpec("get", "sys.type", "the type of system", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemType{}
			return drv, drv.Init(params)
		})

	Methods["sys_location"] = newRouteSpec("get", "sys.location", "the location of system", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemLocation{}
			return drv, drv.Init(params)
		})

	Methods["sys"] = newRouteSpec("get", "sys", "the system info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &systemInfo{}
			return drv, drv.Init(params)
		})

	Methods["interface"] = newRouteSpec("get", "interface", "the interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceAll{}
			return drv, drv.Init(params)
		})

	Methods["interfaceDescr"] = newRouteSpec("get", "interfaceDescr", "the descr part of interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceDescr{}
			return drv, drv.Init(params)
		})

	Methods["interfaceStatus"] = newRouteSpec("get", "interfaceStatus", "the status part of interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceStatus{}
			return drv, drv.Init(params)
		})
	Methods["interfaceScalar"] = newRouteSpec("get", "interfaceScalar", "the scalar part of interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceScalar{}
			return drv, drv.Init(params)
		})

	Methods["tcpConnection"] = newRouteSpec("get", "tcpConnection", "the tcp connection table", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &tcpConnection{}
			return drv, drv.Init(params)
		})

	Methods["udpListen"] = newRouteSpec("get", "udpListen", "the udp listen table", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &udpListen{}
			return drv, drv.Init(params)
		})

}

// func splitsystemOid(oid string) (uint, string) {
// 	if !strings.HasPrefix(oid, "1.3.6.1.4.1.") {
// 		return 0, oid
// 	}
// 	oid = oid[12:]
// 	idx := strings.IndexRune(oid, '.')
// 	if -1 == idx {
// 		u, e := strconv.ParseUint(oid, 10, 0)
// 		if nil != e {
// 			panic(e.Error())
// 		}
// 		return uint(u), ""
// 	}

// 	u, e := strconv.ParseUint(oid[:idx], 10, 0)
// 	if nil != e {
// 		panic(e.Error())
// 	}
// 	return uint(u), oid[idx+1:]
// }

// // func (self *dispatcherBase) RegisterGetFunc(oids []string, get DispatchFunc) {
// // 	for _, oid := range oids {
// // 		main, sub := splitsystemOid(oid)
// // 		methods := self.get_methods[main]
// // 		if nil == methods {
// // 			methods = map[string]DispatchFunc{}
// // 			self.get_methods[main] = methods
// // 		}
// // 		methods[sub] = get
// // 	}
// // }

// func findFunc(oid string, funcs map[uint]map[string]DispatchFunc) DispatchFunc {
// 	main, sub := splitsystemOid(oid)
// 	methods := funcs[main]
// 	if nil == methods {
// 		return nil
// 	}
// 	get := methods[sub]
// 	if nil != get {
// 		return get
// 	}
// 	if "" == sub {
// 		return nil
// 	}
// 	return methods[""]
// }

// func findDefaultFunc(oid string, funcs map[uint]map[string]DispatchFunc) DispatchFunc {
// 	main, sub := splitsystemOid(oid)
// 	methods := funcs[main]
// 	if nil == methods {
// 		return nil
// 	}
// 	if "" == sub {
// 		return nil
// 	}
// 	return methods[""]
// }

// func (self *dispatcherBase) invoke(params MContext, funcs map[uint]map[string]DispatchFunc) commons.Result {
// 	oid, e := self.GetMetricAsString(params, "sys.oid")
// 	if nil != e {
// 		return commons.ReturnError(e.Code(), "get system oid failed, "+e.Error())
// 	}
// 	f := findFunc(oid, funcs)
// 	if nil != f {
// 		res := f(params)
// 		if !res.HasError() {
// 			return res
// 		}
// 		if commons.ContinueCode != res.ErrorCode() {
// 			return res
// 		}

// 		f = findDefaultFunc(oid, funcs)
// 		if nil != f {
// 			res := f(params)
// 			if !res.HasError() {
// 				return res
// 			}
// 			if commons.ContinueCode != res.ErrorCode() {
// 				return res
// 			}
// 		}
// 	}
// 	if nil != self.get {
// 		return self.get(params)
// 	}
// 	return commons.ReturnError(commons.NotAcceptableCode, "Unsupported device - "+oid)
// }
