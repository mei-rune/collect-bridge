package sampling

import (
	"commons"
	"errors"
	"strconv"
)

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
//  e := self.snmpBase.Init(params, drvName)
//  if nil != e {
//    return e
//  }
//  log, ok := params["log"].(*commons.Logger)
//  if !ok {
//    log = commons.Log
//  }

//  restric := false
//  v, ok := params["restric"]
//  if ok {
//    restric = commons.AsBoolWithDefaultValue(v, restric)
//  }

//  dt := commons.SearchFile("etc/device_types.json")
//  if "" == dt {
//    return ErrorIsRestric("'etc/device_types.json' is not exists.", restric, log)
//  }

//  f, err := ioutil.ReadFile(dt)
//  if nil != err {
//    return ErrorIsRestric(fmt.Sprintf("read file '%s' failed, %s", dt, err.Error()), restric, log)
//  }

//  self.device2id = make(map[string]int)
//  err = json.Unmarshal(f, &self.device2id)
//  if nil != err {
//    return ErrorIsRestric(fmt.Sprintf("unmarshal json '%s' failed, %s", dt, err.Error()), restric, log)
//  }

//  return nil
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

func init() {

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
}
