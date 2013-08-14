package poller

import (
	"commons"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Job interface {
	commons.Startable
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
	*trigger
	metric string
	params map[string]string
	client commons.HttpClient
}

func (self *metricJob) Stats() map[string]interface{} {
	res := self.trigger.Stats()
	if nil != self.params {
		for k, v := range self.params {
			res[k] = v
		}
	}
	return res
}

func (self *metricJob) run(t time.Time) error {
	res := self.client.InvokeWith("GET", self.client.Url, nil, 200)
	self.callActions(t, res)
	if res.HasError() {
		return errors.New("sampling failed, " + res.ErrorMessage())
	}
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

	url, e := commons.GetString(ctx, "sampling.url")
	if nil != e {
		return nil, errors.New("'sampling.url' is required, " + e.Error())
	}
	if 0 == len(url) {
		return nil, errors.New("'sampling.url' is required.")
	}

	client_url := ""
	if is_test {
		client_url = commons.NewUrlBuilder(url).Concat("metrics", "managed_object", parentId, metric).ToUrl()
	} else {
		client_url = commons.NewUrlBuilder(url).Concat("managed_object", parentId, metric).ToUrl()
	}

	job := &metricJob{metric: metric,
		params: map[string]string{"managed_type": "managed_object", "managed_id": parentId, "metric": metric, "trigger_id": id},
		client: commons.HttpClient{Url: client_url}}

	job.trigger, e = newTrigger(attributes,
		map[string]interface{}{"managed_type": "managed_object", "managed_id": parentId_int64, "metric": metric, "trigger_id": id},
		ctx,
		job.run)
	return job, e
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
