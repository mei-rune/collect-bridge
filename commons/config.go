package commons

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

func LoadConfig(flagSet *flag.FlagSet) error {
	files := []string{"conf/app.properties",
		"etc/app.properties",
		"../conf/app.properties",
		"../etc/app.properties"}
	found := false
	app_file := ""
	for _, file := range files {
		if FileExists(file) {
			app_file = file
			found = true
			break
		}
	}

	if !found {
		var buffer bytes.Buffer
		buffer.WriteString("[warn] properties file is not exists, search path is:\r\n")
		for _, file := range files {
			buffer.WriteString("    ")
			buffer.WriteString(file)
			buffer.WriteString("\r\n")
		}
		fmt.Println(buffer.String())
		return nil
	}
	cfg, e := ReadProperties(app_file)
	if nil != e {
		return errors.New("read properties '" + app_file + "' failed," + e.Error())
	}

	actual := map[string]string{}
	fn := func(f *flag.Flag) {
		actual[f.Name] = f.Name
	}
	if nil == flagSet {
		flag.Visit(fn)
	} else {
		flagSet.Visit(fn)
	}

	formul := map[string]string{}
	fn = func(f *flag.Flag) {
		actual[f.Name] = f.Name
	}
	if nil == flagSet {
		flag.VisitAll(fn)
	} else {
		flagSet.VisitAll(fn)
	}

	defaults := map[string]string{"redis.host": "127.0.0.1",
		"redis.port":       "36379",
		"db.type":          "postgresql",
		"db.address":       "127.0.0.1",
		"db.port":          "35432",
		"data.db.schema":   "tpt_data",
		"models.db.schema": "tpt",
		"db.username":      "tpt",
		"db.password":      "extreme"}

	for k, _ := range formul {
		if _, ok := actual[k]; ok {
			continue
		}

		switch k {
		case "redis_address":
			redis_address := stringWith(cfg, "redis.host", defaults["redis.host"])
			redis_port := stringWith(cfg, "redis.port", defaults["redis.port"])
			if nil == flagSet {
				flag.Set(k, redis_address+":"+redis_port)
			} else {
				flagSet.Set(k, redis_address+":"+redis_port)
			}
		case "data_db.url":
			drv, url, e := CreateDBUrl("data.", cfg, defaults)
			if nil != e {
				return e
			}
			if nil == flagSet {
				flag.Set("data_db.url", url)
				flag.Set("data_db.driver", drv)
			} else {
				flagSet.Set("data_db.url", url)
				flagSet.Set("data_db.driver", drv)
			}
		case "db.url", "db_url":
			drv, url, e := CreateDBUrl("models.", cfg, defaults)
			if nil != e {
				return e
			}
			if nil == flagSet {
				flag.Set("db.url", url)
				flag.Set("db.driver", drv)
			} else {
				flagSet.Set("db.url", url)
				flagSet.Set("db.driver", drv)
			}
		}

		if v, ok := cfg[k]; ok {
			if nil == flagSet {
				flag.Set(k, v)
			} else {
				flagSet.Set(k, v)
			}
		}
	}
	return nil
}

func IsSetFlagVar(name string) (ret bool) {
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			ret = true
		}
	})
	return ret
}

func stringWith(props map[string]string, nm, defaultValue string) string {
	if v, ok := props[nm]; ok && 0 != len(v) {
		return v
	}
	return defaultValue
}

func CreateDBUrl(prefix string, props, defaultValues map[string]string) (string, string, error) {
	db_type := stringWith(props, prefix+"db.type", stringWith(defaultValues, prefix+"db.type", stringWith(defaultValues, "db.type", "")))
	db_address := stringWith(props, prefix+"db.address", stringWith(defaultValues, prefix+"db.address", stringWith(defaultValues, "db.address", "")))
	db_port := stringWith(props, prefix+"db.port", stringWith(defaultValues, prefix+"db.port", stringWith(defaultValues, "db.port", "")))
	db_schema := stringWith(props, prefix+"db.schema", stringWith(defaultValues, prefix+"db.schema", stringWith(defaultValues, "db.schema", "")))
	db_username := stringWith(props, prefix+"db.username", stringWith(defaultValues, prefix+"db.username", stringWith(defaultValues, "db.username", "")))
	db_password := stringWith(props, prefix+"db.password", stringWith(defaultValues, prefix+"db.password", stringWith(defaultValues, "db.password", "")))
	switch db_type {
	case "postgresql":
		return "postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			db_address, db_port, db_schema, db_username, db_password), nil
	default:
		return "", "", errors.New("unknown db type - " + db_type)
	}
}

func LoadConfigFromJsonFile(nm string, flagSet *flag.FlagSet, isOverride bool) error {
	f, e := os.Open(nm)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}

	var res map[string]interface{}
	e = json.NewDecoder(f).Decode(&res)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}

	actual := map[string]string{}
	fn := func(f *flag.Flag) {
		actual[f.Name] = f.Name
	}
	if nil == flagSet {
		flag.Visit(fn)
	} else {
		flagSet.Visit(fn)
	}

	e = assignFlagSet("", res, flagSet, actual, isOverride)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}
	return nil
}

func assignFlagSet(prefix string, res map[string]interface{}, flagSet *flag.FlagSet, actual map[string]string, isOverride bool) error {
	for k, v := range res {
		switch value := v.(type) {
		case map[string]interface{}:
			e := assignFlagSet(combineName(prefix, k), value, flagSet, actual, isOverride)
			if nil != e {
				return e
			}
			continue
		case []interface{}:
		case string:
		case float64:
		case bool:
		case nil:
			continue
		default:
			return fmt.Errorf("unsupported type for %s - %T", combineName(prefix, k), v)
		}
		nm := combineName(prefix, k)

		if !isOverride {
			if _, ok := actual[nm]; ok {
				log.Printf("load flag '%s' from config is skipped.\n", nm)
				continue
			}
		}

		var g *flag.Flag
		if nil == flagSet {
			g = flag.Lookup(nm)
		} else {
			g = flagSet.Lookup(nm)
		}
		if nil == g {
			log.Printf("flag '%s' is not defined.\n", nm)
			continue
		}

		err := g.Value.Set(fmt.Sprint(v))
		if nil != err {
			return err
		}
	}
	return nil
}

func combineName(prefix, nm string) string {
	if "" == prefix {
		return nm
	}
	return prefix + "." + nm
}

func SetFlags(cfg map[string]string, flagSet *flag.FlagSet, isOverride bool) {
	actual := map[string]string{}
	flags := make([]*flag.Flag, 0, 10)
	if nil == flagSet {
		if !isOverride {
			flag.Visit(func(g *flag.Flag) {
				actual[g.Name] = g.Name
			})
		}
		flag.VisitAll(func(g *flag.Flag) {
			if isOverride {
				flags = append(flags, g)
			} else if _, ok := actual[g.Name]; !ok {
				flags = append(flags, g)
			}
		})
		for _, g := range flags {
			if v, ok := cfg[g.Name]; ok {
				flag.Set(g.Name, v)
			}
		}
	} else {
		if !isOverride {
			flagSet.Visit(func(g *flag.Flag) {
				actual[g.Name] = g.Name
			})
		}
		flagSet.VisitAll(func(g *flag.Flag) {
			if isOverride {
				flags = append(flags, g)
			} else if _, ok := actual[g.Name]; !ok {
				flags = append(flags, g)
			}
		})
		for _, g := range flags {
			if v, ok := cfg[g.Name]; ok {
				flagSet.Set(g.Name, v)
			}
		}
	}
}
