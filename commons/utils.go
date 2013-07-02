package commons

import (
	"errors"
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
	TypeError = errors.New("it is not []interface{} or map[string]interface{}")
)

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
		return nil, TypeError
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
	return paths, nil
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
