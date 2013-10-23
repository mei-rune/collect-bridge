package poller

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type aggregator interface {
	aggregate(value interface{}, timestamp time.Time) (interface{}, bool, error)
}

type averager struct {
	current_all float64
	count       int
	sampled_at  time.Time
	interval    time.Duration
}

func (self *averager) aggregate(value interface{}, timestamp time.Time) (interface{}, bool, error) {
	switch v := value.(type) {
	case json.Number:
		f64, e := v.Float64()
		if nil != e {
			return nil, false, errors.New("crazy! it is not a number? - " + v.String())
		}
		value = f64
		self.current_all += f64
		break
	case int64:
		self.current_all += float64(v)
		break
	case int:
		self.current_all += float64(v)
		break
	case float64:
		self.current_all += v
		break
	case float32:
		self.current_all += float64(v)
		break
	case string:
		currentValue, e := strconv.ParseFloat(v, 10)
		if nil != e {
			return nil, false, errors.New("crazy! it is not a number? - " + v)
		}
		self.current_all += currentValue
		break
	case int32:
		self.current_all += float64(v)
		break
	case int8:
		self.current_all += float64(v)
		break
	case int16:
		self.current_all += float64(v)
		break
	case uint:
		self.current_all += float64(v)
		break
	case uint8:
		self.current_all += float64(v)
		break
	case uint16:
		self.current_all += float64(v)
		break
	case uint32:
		self.current_all += float64(v)
		break
	case uint64:
		self.current_all += float64(v)
		break
	case uintptr:
		self.current_all += float64(v)
		break
	default:
		return nil, false, fmt.Errorf("value is not a number- %T", v)
	}

	if timestamp.Before(self.sampled_at) {
		self.sampled_at = timestamp
		self.count = 0
		self.current_all = 0
		return value, true, nil
	}

	self.count += 1
	if timestamp.Sub(self.sampled_at) > self.interval {
		value = self.current_all / float64(self.count)
		self.sampled_at = timestamp
		self.count = 0
		self.current_all = 0
		return value, true, nil
	}
	return nil, false, nil
}

type last_merger struct {
	sampled_at time.Time
	interval   time.Duration
}

func (self *last_merger) aggregate(value interface{}, timestamp time.Time) (interface{}, bool, error) {
	if timestamp.After(self.sampled_at) && timestamp.Sub(self.sampled_at) < self.interval {
		//fmt.Println(self.sampled_at, timestamp, timestamp.Sub(self.sampled_at))
		return nil, false, nil
	}

	var e error
	switch v := value.(type) {
	case json.Number:
		value, e = v.Float64()
		if nil != e {
			return nil, false, errors.New("crazy! it is not a number? - " + v.String())
		}
	case int64:
		break
	case int:
		break
	case float64:
		break
	case float32:
		break
	case string:
		value, e = strconv.ParseFloat(v, 10)
		if nil != e {
			return nil, false, errors.New("crazy! it is not a number? - " + v)
		}
	case bool:
		if v {
			value = 1
		} else {
			value = 0
		}
	case int32:
		break
	case int8:
		break
	case int16:
		break
	case uint:
		break
	case uint8:
		break
	case uint16:
		break
	case uint32:
		break
	case uint64:
		break
	case uintptr:
		break
	default:
		return nil, false, fmt.Errorf("value is not a number- %T", v)
	}

	self.sampled_at = timestamp
	return value, true, nil
}
