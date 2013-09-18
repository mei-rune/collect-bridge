package poller

import (
	"commons"
	"errors"
	"fmt"
	"sampling"
	"strconv"
	"time"
)

const (
	CLOSE_REASON_NORMAL   = -1
	CLOSE_REASON_UNKNOW   = 0
	CLOSE_REASON_DISABLED = 1
	CLOSE_REASON_DELETED  = 2
	CLOSE_REASON_MAX      = 2
)

type Job interface {
	Interupt()
	Close(reason int)

	Id() string
	Name() string
	Stats() map[string]interface{}
	Version() time.Time
}

func newJob(attributes, ctx map[string]interface{}) (Job, error) {
	t := attributes["type"]
	switch t {
	case "metric_trigger":
		return createMetricJob(attributes, ctx)
	}
	return nil, errors.New("unsupport job type - " + fmt.Sprint(t))
}

type metricJob struct {
	Trigger
	metric string
	params map[string]string
	client sampling.ChannelClient
}

func (self *metricJob) Stats() map[string]interface{} {
	res := self.Trigger.Stats()
	if nil != self.params {
		for k, v := range self.params {
			res[k] = v
		}
	}
	return res
}

func (self *metricJob) run(t time.Time) error {
	self.client.Send()
	return nil
}

func createMetricJob(attributes, ctx map[string]interface{}) (Job, error) {
	id, e := commons.GetString(attributes, "id")
	if nil != e {
		return nil, errors.New("'id' is required, " + e.Error())
	}
	if 0 == len(id) {
		return nil, errors.New("'id' is empty")
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

	//client_url := commons.NewUrlBuilder(url).Concat("managed_object", parentId, metric).ToUrl()

	job := &metricJob{metric: metric,
		params: map[string]string{"managed_type": "managed_object", "managed_id": parentId, "metric": metric, "trigger_id": id}}

	//	client: commons.HttpClient{Url: client_url}}

	job.Trigger, e = newTrigger(attributes,
		map[string]interface{}{"managed_type": "managed_object", "managed_id": parentId_int64, "metric": metric, "trigger_id": id},
		ctx,
		job.run)
	if nil != e {
		return nil, e
	}

	cname := sampling.MakeChannelName(metric, "managed_object", parentId, "", nil)
	client, e := broker.SubscribeClient(cname, job.Trigger.GetChannel(), "GET", metric, "managed_object", parentId, "", nil, nil, 8*time.Second)
	if nil != e {
		job.Close(CLOSE_REASON_NORMAL)
		return nil, e
	}
	job.client = client
	return job, nil
}

// func createRequest(nm string, attributes, ctx map[string]interface{}) (string, bytes.Buffer, error) {
// 	url, e := commons.GetString(ctx, "metric_url")
// 	if nil != e {
// 		return nil, errors.New("'metric_url' is required, " + e.Error())
// 	}
// 	params := attributes["$parent"]

// 	var urlBuffer bytes.Buffer
// 	urlBuffer.WriteString(url)
// 	urlBuffer.WriteString("/")
// 	urlBuffer.WriteString(nm)
// 	urlBuffer.WriteString("/")
// 	urlBuffer.WriteString(nm)
// }

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
