package commons

import (
	"errors"
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
