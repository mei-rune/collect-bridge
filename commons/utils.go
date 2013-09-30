package commons

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	//"time"
)

var (
	ThrowPanic = func(instance interface{}) {
		panic(fmt.Sprintf("it is not []interface{} or map[string]interface{}, actual is [%T]%v", instance, instance))
	}
	typeError = TypeError("it is not []interface{} or map[string]interface{}")
)

type Closeable interface {
	Close()
}

func Close(close_list []Closeable) {
	for _, c := range close_list {
		c.Close()
	}
}

func Iterator(instance interface{}) ([][2]interface{}, error) {
	results := make([][2]interface{}, 0, 10)
	switch values := instance.(type) {
	case map[string]interface{}:
		for ck, r := range values {
			results = append(results, [2]interface{}{ck, r})
		}
	case []interface{}:
		for ck, r := range values {
			results = append(results, [2]interface{}{ck, r})
		}
	default:
		return nil, typeError
	}
	return results, nil
}

func Each(instance interface{}, cb func(k interface{}, v interface{}), default_cb func(instance interface{})) {
	switch values := instance.(type) {
	case map[string]interface{}:
		for ck, r := range values {
			cb(ck, r)
		}
	case []interface{}:
		for ck, r := range values {
			cb(ck, r)
		}
	default:
		if nil != default_cb {
			default_cb(instance)
		}
	}
}

func ConvertToIntList(value, sep string) ([]int, error) {
	ss := strings.Split(value, sep)
	results := make([]int, 0, len(ss))
	for _, s := range ss {
		i, e := strconv.ParseInt(s, 10, 32)
		if nil != e {
			return nil, errors.New("'" + value + "' contains nonnumber.")
		}
		results = append(results, int(i))
	}
	return results, nil
}
func GetIntList(params map[string]string, key string) ([]int, error) {
	v, ok := params[key]
	if !ok {
		return nil, errors.New("'" + key + "' is not exists in the params.")
	}

	ss := strings.Split(v, ",")
	results := make([]int, 0, len(ss))
	for _, s := range ss {
		i, e := strconv.ParseInt(s, 10, 32)
		if nil != e {
			return nil, errors.New("'" + key + "' contains nonnumber - " + v + ".")
		}
		results = append(results, int(i))
	}
	return results, nil
}

func DirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}

	return true
}

func FileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func EnumerateFiles(pa string) ([]string, error) {
	if "" == pa {
		return nil, errors.New("path is empty.")
	}

	dir, serr := os.Stat(pa)
	if serr != nil {
		return nil, serr
	}

	if !dir.IsDir() {
		return nil, errors.New(pa + " is not a directory")
	}

	fd, err := os.Open(pa)
	if nil != err {
		return nil, err
	}
	defer fd.Close()

	paths := make([]string, 0, 30)
	for {
		dirs, err := fd.Readdir(10)
		if nil != err {
			if io.EOF == err {
				return paths, nil
			} else {
				return nil, err
			}
		}
		for _, dir := range dirs {
			if dir.IsDir() {
				sub_paths, err := EnumerateFiles(path.Join(pa, dir.Name()))
				if nil != err {
					return nil, err
				}
				for _, sp := range sub_paths {
					paths = append(paths, sp)
				}
			} else {
				paths = append(paths, path.Join(pa, dir.Name()))
			}
		}
	}
}

func SearchFile(pattern string) string {
	ed, e := filepath.Abs(os.Args[0])
	if nil == e {
		ed = strings.Replace(filepath.Dir(ed), "\\", "/", -1)
		pa := path.Join(ed, pattern)
		if FileExists(pa) {
			return pa
		}
		pa = path.Join(ed, "..", pattern)
		if FileExists(pa) {
			return pa
		}
	}

	wd, e := os.Getwd()
	if nil == e {
		wd, e := filepath.Abs(wd)
		if nil == e {
			wd = strings.Replace(wd, "\\", "/", -1)
			pa := path.Join(wd, pattern)
			if FileExists(pa) {
				return pa
			}
			pa = path.Join(wd, "..", pattern)
			if FileExists(pa) {
				return pa
			}
		}
	}
	return ""
}

// func EscapeString(s string, escapeChar, endChar rune, charset string) (string, string) {
// 	var buffer bytes.Buffer
// 	is_escape := false
// 	for i, c := range s {
// 		if is_escape {
// 			if !strings.IndexRune(charset, c) {
// 				buffer.WriteRune('\\')
// 				if endChar == c {
// 					return buffer.String(), s[i+1:]
// 				}
// 			}
// 			buffer.WriteRune(rn)
// 			is_escape = false
// 		} else if endChar == c {
// 			return buffer.String(), s[i+1:]
// 		} else if escapeChar == c {
// 			is_escape = true
// 		} else {
// 			buffer.WriteRune(rn)
// 		}
// 	}
// 	return buffer.String(), ""
// }

func ReadProperties(nm string) (map[string]string, error) {
	f, e := os.Open(nm)
	if nil != e {
		return nil, e
	}
	defer f.Close()

	cfg := map[string]string{}
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
		cfg[key] = os.ExpandEnv(value)
	}
	return cfg, nil
}

func LoadDefaultProperties(db_prefix, drv_name, url_name, redis_name string, defaults map[string]string) error {
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

	if "" != url_name && !IsSetFlagVar(url_name) {
		drv, url, e := CreateDBUrl(db_prefix, cfg, defaults)
		if nil != e {
			return e
		}
		flag.Set(url_name, url)
		flag.Set(drv_name, drv)
	}

	if "" != redis_name && !IsSetFlagVar(redis_name) {
		redis_address := stringWith(cfg, "redis.host", defaults["redis.host"])
		redis_port := stringWith(cfg, "redis.port", defaults["redis.port"])
		flag.Set(redis_name, redis_address+":"+redis_port)
	}
	SetFlags(cfg, nil, false)
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
	db_type := stringWith(props, prefix+"db.type", defaultValues["db.type"])
	db_address := stringWith(props, prefix+"db.address", defaultValues["db.address"])
	db_port := stringWith(props, prefix+"db.port", defaultValues["db.port"])
	db_schema := stringWith(props, prefix+"db.schema", defaultValues["db.schema"])
	db_username := stringWith(props, prefix+"db.username", defaultValues["db.username"])
	db_password := stringWith(props, prefix+"db.password", defaultValues["db.password"])
	switch db_type {
	case "postgresql":
		return "postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			db_address, db_port, db_schema, db_username, db_password), nil
	default:
		return "", "", errors.New("unknown db type - " + db_type)
	}
}

//const time_format = "time format error with valuw is '%s', excepted format is 'xxx[unit]', xxx is a number, unit must is in (ms, s, m)."

// func ParseDuration(s string) (time.Duration, error) {
// 	return time.ParseDuration(s)

// 	// idx := strings.IndexFunc(s, func(r rune) bool {
// 	// 	switch r {
// 	// 	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
// 	// 		return false
// 	// 	}
// 	// 	return true
// 	// })

// 	// if idx == 0 {
// 	// 	return 0, fmt.Errorf(time_format, s)
// 	// }

// 	// unit := time.Second
// 	// if -1 != idx {
// 	// 	switch s[idx:] {
// 	// 	case "ms", "MS":
// 	// 		unit = time.Millisecond
// 	// 	case "s", "S":
// 	// 		unit = time.Second
// 	// 	case "m", "M":
// 	// 		unit = time.Minute
// 	// 	default:
// 	// 		return 0, fmt.Errorf(time_format, s)
// 	// 	}
// 	// 	s = s[:idx]
// 	// }

// 	// i, err := strconv.ParseInt(s, 10, 0)
// 	// if nil != err {
// 	// 	return 0, fmt.Errorf(time_format, s, err.Error())
// 	// }
// 	// return time.Duration(i) * unit, nil
// }
