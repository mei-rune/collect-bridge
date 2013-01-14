package commons

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

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
