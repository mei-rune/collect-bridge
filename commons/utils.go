package commons

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

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
	return paths, nil
}

const time_format = "time format error with valuw is '%s', excepted format is 'xxx[unit]', xxx is a number, unit must is in (ms, s, m)."

func ParseTime(s string) (time.Duration, error) {
	idx := strings.IndexFunc(s, func(r rune) bool {
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return false
		}
		return true
	})

	if idx == 0 {
		return 0, fmt.Errorf(time_format, s)
	}

	unit := time.Second
	if -1 != idx {
		switch s[idx:] {
		case "ms", "MS":
			unit = time.Millisecond
		case "s", "S":
			unit = time.Second
		case "m", "M":
			unit = time.Minute
		default:
			return 0, fmt.Errorf(time_format, s)
		}
		s = s[:idx]
	}

	i, err := strconv.ParseInt(s, 10, 0)
	if nil != err {
		return 0, fmt.Errorf(time_format, s, err.Error())
	}
	return time.Duration(i) * unit, nil
}

func IsReturnOk(params map[string]interface{}) bool {
	v, ok := params["value"]
	if ok {
		return v == "ok"
	}
	return false
}

func GetReturn(params map[string]interface{}) interface{} {
	v, ok := params["value"]
	if ok {
		return v
	}
	return nil
}

func Return(value interface{}) map[string]interface{} {
	return map[string]interface{}{"value": value}
}

func ReturnWithKV(params map[string]interface{}, key string, value interface{}) map[string]interface{} {
	params[key] = value
	return params
}

func ReturnWithValue(params map[string]interface{}, value interface{}) map[string]interface{} {
	params["value"] = value
	return params
}

func ReturnOK() map[string]interface{} {
	return map[string]interface{}{"value": "ok"}
}

func ReturnFailed() map[string]interface{} {
	return map[string]interface{}{"value": "ok"}
}
