package poller

import (
	"commons"
	"errors"
	"fmt"
	"time"
)

type Job interface {
	commons.Startable
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

func (self *metricJob) Run(t time.Time) {
	res := self.client.Invoke("GET", self.client.Url, nil, 200)
	if res.HasError() {
		self.WARN.Printf("read metric '%s' failed, %v", self.metric, res.ErrorMessage())
		return
	}
	self.callActions(t, res)
}

func createMetricJob(attributes, ctx map[string]interface{}) (Job, error) {
	metric, e := commons.GetString(attributes, "metric")
	if nil != e {
		return nil, errors.New("'metric' is required, " + e.Error())
	}
	parentType, e := commons.GetString(attributes, "parent_type")
	if nil != e {
		return nil, errors.New("'parent_type' is required, " + e.Error())
	}
	parentId, e := commons.GetString(attributes, "parent_id")
	if nil != e {
		return nil, errors.New("'parent_id' is required, " + e.Error())
	}
	url, e := commons.GetString(ctx, "metrics.url")
	if nil != e {
		return nil, errors.New("'metrics.url' is required, " + e.Error())
	}
	if 0 == len(url) {
		return nil, errors.New("'metrics.url' is required.")
	}

	job := &metricJob{metric: metric,
		params: map[string]string{"managed_type": parentType, "managed_id": parentId, "metric": metric},
		client: commons.HttpClient{Url: commons.NewUrlBuilder(url).Concat(parentType, parentId, metric).ToUrl()}}

	job.trigger, e = newTrigger(attributes, func(t time.Time) { job.Run(t) }, ctx)
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
