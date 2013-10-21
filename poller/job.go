package poller

import (
	"bytes"
	"commons"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sampling"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var only_thread_safe = true

func newJob(attributes, ctx map[string]interface{}) (Job, error) {
	t := attributes["type"]
	switch t {
	case "metric_trigger":
		return createMetricJob(attributes, ctx)
	}
	return nil, errors.New("unsupport job type - " + fmt.Sprint(t))
}

type metricJob struct {
	*baseJob
	c              chan interface{}
	client         sampling.ChannelClient
	triggerBuilder TriggerBuilder
	wait           sync.WaitGroup
	closed         int32

	metric            string
	max_used_duration int64
	begin_fired_at    int64
	end_fired_at      int64

	timeout_histories commons.Int64Buffer
	recv_histories    samplingBuffer
}

func (self *metricJob) Interupt() {
	if 1 == atomic.LoadInt32(&self.closed) {
		return
	}

	self.c <- 0
}

func (self *metricJob) Close(reason int) {
	if 1 == atomic.LoadInt32(&self.closed) {
		return
	}
	if nil != self.client {
		self.client.Close()
	}
	close(self.c)
	self.wait.Wait()
	self.reset(reason)
}

func toTimes(ts []int64) []time.Time {
	if nil == ts {
		return nil
	}

	ret := make([]time.Time, len(ts))
	for i := 0; i < len(ret); i++ {
		ret[i] = time.Unix(ts[i], 0)
	}
	return ret
}

func toTimesWithResult(ts []samplingResult) []interface{} {
	if nil == ts {
		return nil
	}

	ret := make([]interface{}, len(ts))
	for i := 0; i < len(ret); i++ {
		var rm json.RawMessage
		if ts[i].is_ok {
			rm = json.RawMessage("[\"" + time.Unix(ts[i].sampled_at, 0).Format(time.RFC3339) + "\", true]")
		} else {
			rm = json.RawMessage("[\"" + time.Unix(ts[i].sampled_at, 0).Format(time.RFC3339) + "\", false]")
		}
		ret[i] = &rm
	}
	return ret
}

func (self *metricJob) Stats() map[string]interface{} {
	m := self.baseJob.Stats()
	m["max_used_duration"] = strconv.FormatInt(atomic.LoadInt64(&self.max_used_duration), 10) + "s"
	m["begin_fired_at"] = time.Unix(atomic.LoadInt64(&self.begin_fired_at), 0)
	m["end_fired_at"] = time.Unix(atomic.LoadInt64(&self.end_fired_at), 0)
	// NOTE: it is not thread safely
	if !only_thread_safe {
		m["timeout_histories"] = toTimes(self.timeout_histories.All())
		m["recv_histories"] = toTimesWithResult(self.recv_histories.All())
	}

	return m
}

func (self *metricJob) init(delay time.Duration) error {
	sinit := func() {
		if 1 == atomic.LoadInt32(&self.closed) {
			return
		}

		go self.run()
		self.wait.Add(1)
	}

	if 0 == delay {
		sinit()
	} else {
		time.AfterFunc(delay, sinit)
	}

	return nil
}

func (self *metricJob) run() {
	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			msg := buffer.String()
			log.Println(msg)
			self.set_last_error(msg)
		}

		atomic.StoreInt32(&self.closed, 1)
		self.wait.Done()
	}()

	trigger, e := self.triggerBuilder.New()
	if nil != e {
		self.set_last_error(e.Error())
		return
	}
	defer trigger.Close()

	is_running := true
	for is_running {
		select {
		case o, ok := <-self.c:
			if !ok || 0 == o {
				is_running = false
				break
			}

			res, ok := o.(ValueResult)
			if !ok {
				self.set_last_error(fmt.Sprintf("sampling failed, unsupported type - %T", o))
				break
			}

			self.callActions(res.CreatedAt(), res)
			if res.HasError() {
				self.set_last_error("sampling failed, " + res.ErrorMessage())
			} else {
				self.set_last_error("")
			}

			now := time.Now().Unix()
			atomic.StoreInt64(&self.end_fired_at, now)
			used_duration := now - atomic.LoadInt64(&self.begin_fired_at)
			if self.max_used_duration < used_duration {
				atomic.StoreInt64(&self.max_used_duration, used_duration)
			}
			self.recv_histories.Push(samplingResult{sampled_at: now, is_ok: !res.HasError()})
		case <-trigger.Channel():
			self.timeout(time.Now())
		}
	}
}

func (self *metricJob) timeout(t time.Time) {
	startAt := t.Unix()
	if 60 > startAt-atomic.LoadInt64(&self.begin_fired_at) {
		return
	}

	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			msg := buffer.String()
			self.set_last_error(msg)
			log.Println(msg)

			endAt := time.Now().Unix()
			atomic.StoreInt64(&self.end_fired_at, endAt)
			if self.max_used_duration < endAt-startAt {
				atomic.StoreInt64(&self.max_used_duration, endAt-startAt)
			}
		}
	}()

	self.timeout_histories.Push(startAt)
	atomic.StoreInt64(&self.begin_fired_at, startAt)
	self.client.Send()
}

func createMetricJob(attributes, ctx map[string]interface{}) (Job, error) {
	id := commons.GetStringWithDefault(attributes, "id", "")
	if "" == id {
		return nil, IdIsRequired
	}

	metric, e := commons.GetString(attributes, "metric")
	if nil != e {
		return nil, errors.New("'metric' is required, " + e.Error())
	}
	parentId, e := commons.GetString(attributes, "managed_object_id")
	if nil != e {
		return nil, errors.New("'managed_object_id' is required, " + e.Error())
	}
	parentId_int64, e := strconv.ParseInt(parentId, 10, 64)
	if nil != e {
		return nil, errors.New("'managed_object_id' is not a int64, " + e.Error())
	}

	var broker *sampling.SamplingBroker
	var ok bool

	sb := ctx["sampling_broker"]
	if nil == sb {
		return nil, errors.New("'sampling_broker' is required.")
	}
	if broker, ok = sb.(*sampling.SamplingBroker); !ok {
		return nil, errors.New("'sampling_broker' is not a SamplingBroker in the ctx.")
	}

	triggerBuilder, e := newTrigger(attributes, ctx)
	if nil != e {
		return nil, e
	}

	options := map[string]interface{}{"managed_type": "managed_object",
		"managed_id": parentId_int64, "metric": metric, "trigger_id": id, "interval": triggerBuilder.Interval()}
	base, e := newBase(attributes, options, ctx)
	if nil != e {
		return nil, e
	}

	c := make(chan interface{}, 10)
	cname := sampling.MakeChannelName(metric, "managed_object", parentId, "", nil)
	client, e := broker.SubscribeClient(cname, c, "GET", metric, "managed_object", parentId, "", nil, nil, 8*time.Second)
	if nil != e {
		return nil, e
	}

	job := &metricJob{baseJob: base,
		c:              c,
		client:         client,
		closed:         0,
		metric:         metric,
		triggerBuilder: triggerBuilder}

	job.timeout_histories.Init(make([]int64, 10))
	job.recv_histories.Init(make([]samplingResult, 10))

	e = job.init(delay_interval())
	if nil != e {
		job.Close(CLOSE_REASON_NORMAL)
		return nil, e
	}

	return job, nil
}

type errorJob struct {
	clazz, id, name, e string

	updated_at time.Time
}

func (self *errorJob) Interupt() {
}

func (self *errorJob) Close(reason int) {
}

func (self *errorJob) Id() string {
	return self.id
}

func (self *errorJob) Name() string {
	return self.name
}

func (self *errorJob) Stats() map[string]interface{} {
	return map[string]interface{}{
		"type":  self.clazz,
		"id":    self.id,
		"name":  self.name,
		"error": self.e}
}

func (self *errorJob) Version() time.Time {
	return self.updated_at
}
