package sampling

import (
	"bufio"
	"commons"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type systemOid struct {
	snmpBase
}

func (self *systemOid) Call(params MContext) (interface{}, error) {
	return self.GetResult(params, "1.3.6.1.2.1.1.2.0", RES_OID)
}

type systemDescr struct {
	snmpBase
}

func (self *systemDescr) Call(params MContext) (interface{}, error) {
	return self.GetResult(params, "1.3.6.1.2.1.1.1.0", RES_STRING)
}

type systemName struct {
	snmpBase
}

func (self *systemName) Call(params MContext) (interface{}, error) {
	return self.GetResult(params, "1.3.6.1.2.1.1.5.0", RES_STRING)
}

type systemUpTime struct {
	snmpBase
}

func (self *systemUpTime) Call(params MContext) (interface{}, error) {
	t, e := self.GetUint64(params, "1.3.6.1.2.1.1.3.0")
	if nil != e {
		return nil, e
	}
	return (t / 100), nil
}

type systemLocation struct {
	snmpBase
}

func (self *systemLocation) Call(params MContext) (interface{}, error) {
	return self.GetResult(params, "1.3.6.1.2.1.1.6.0", RES_STRING)
}

type systemServices struct {
	snmpBase
}

func (self *systemServices) Call(params MContext) (interface{}, error) {
	return self.GetResult(params, "1.3.6.1.2.1.1.7.0", RES_INT64)
}

type systemInfo struct {
	snmpBase
}

func (self *systemInfo) Call(params MContext) (interface{}, error) {
	return self.GetOneResult(params, "1.3.6.1.2.1.1", "",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			oid := GetOidWithDefault(params, old_row, "2")
			services := GetUint32WithDefault(params, old_row, "7", 0)

			new_row := map[string]interface{}{}
			new_row["descr"] = GetStringWithDefault(params, old_row, "1")
			new_row["oid"] = oid
			new_row["upTime"] = GetUint32WithDefault(params, old_row, "3", 0) / 100
			new_row["contact"] = GetStringWithDefault(params, old_row, "4")
			new_row["name"] = GetStringWithDefault(params, old_row, "5")
			new_row["location"] = GetStringWithDefault(params, old_row, "6")
			new_row["services"] = services

			params.Set("$sys.oid", oid)
			params.Set("$sys.services", strconv.Itoa(int(services)))
			t, _ := params.Read().GetInt("sys.type", nil, params)
			new_row["type"] = t
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

func (self *systemType) Init(params map[string]interface{}) error {
	file := commons.GetStringWithDefault(params, "oid2type", "oid2type.dat")
	binDir := filepath.Dir(abs(os.Args[0]))
	files := []string{abs(file),
		abs(filepath.Join("conf", file)),
		abs(filepath.Join("etc", file)),
		abs(filepath.Join("lib", file)),
		abs(filepath.Join("..", "conf", file)),
		abs(filepath.Join("..", "etc", file)),
		abs(filepath.Join("..", "lib", file)),
		abs(filepath.Join(binDir, file)),
		abs(filepath.Join(binDir, "..", file)),
		abs(filepath.Join(binDir, "conf", file)),
		abs(filepath.Join(binDir, "etc", file)),
		abs(filepath.Join(binDir, "lib", file)),
		abs(filepath.Join(binDir, "..", "conf", file)),
		abs(filepath.Join(binDir, "..", "etc", file)),
		abs(filepath.Join(binDir, "..", "lib", file))}

	found := false
	for _, nm := range files {
		if st, e := os.Stat(nm); nil == e && nil != st && !st.IsDir() {
			file = nm
			found = true
			break
		}
	}

	if found {
		f, e := os.Open(file)
		if nil != e {
			log.Println("[warn] load oid2type config from '"+file+"' failed,", e)
		} else {
			defer f.Close()
			log.Println("[warn] load oid2type config from '" + file + "'.")
			self.device2id = map[string]int{}
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				ss := strings.SplitN(scanner.Text(), "#", 2)
				ss = strings.SplitN(ss[0], "//", 2)
				s := strings.TrimSpace(ss[0])
				if 0 == len(s) {
					continue
				}
				ss = strings.SplitN(s, "=", 2)
				key := strings.TrimLeft(strings.TrimSpace(ss[0]), ".")
				value := strings.TrimSpace(ss[1])
				if 0 == len(key) {
					continue
				}
				if 0 == len(value) {
					continue
				}

				t, e := strconv.ParseInt(value, 10, 0)
				if nil != e {
					continue
				}

				self.device2id[key] = int(t)
			}
		}
	} else {
		log.Println("[warn] load oid2type config failed, file is not founc:")
		for _, nm := range files {
			log.Println("\t\t", nm)
		}
	}
	return self.snmpBase.Init(params)
}

func (self *systemType) Call(params MContext) (interface{}, error) {
	if nil != self.device2id && 0 != len(self.device2id) {
		oid := params.GetStringWithDefault("$sys.oid", "")
		if 0 != len(oid) {
			if dt, ok := self.device2id[strings.TrimPrefix(oid, ".")]; ok {
				return dt, nil
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
		return (t >> 1), nil
	}
SERVICES:
	services, e := params.GetUint32("$sys.services")
	if nil != e {
		return nil, e
	}
	return ((services & 0x7) >> 1), nil
}

func abs(pa string) string {
	s, e := filepath.Abs(pa)
	if nil != e {
		panic(e.Error())
	}
	return s
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
