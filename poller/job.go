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

func NewJob(attributes, ctx map[string]interface{}) (Job, error) {
	t := attributes["type"]
	switch t {
	case "metric_rule":
		return createMetricJob(attributes, ctx)
	}
	return nil, errors.New("unsupport job type - " + fmt.Sprint(t))
}

type metricJob struct {
	*Trigger
	metric string
	params map[string]string
	drv    commons.Driver
}

func (self *metricJob) Run(t time.Time) {
	res, e := self.drv.Get(self.params)
	if nil != e {
		self.WARN.Printf("read metric '%s' failed, %v", self.metric, e)
		return
	}
	self.CallActions(t, commons.GetReturn(res))
}

func createMetricJob(attributes, ctx map[string]interface{}) (commons.Startable, error) {
	metric, e := commons.TryGetString(attributes, "metric")
	if nil != e {
		return nil, errors.New("'metric' is required, " + e.Error())
	}
	parentType, e := commons.TryGetString(attributes, "parent_type")
	if nil != e {
		return nil, errors.New("'parent_type' is required, " + e.Error())
	}
	parentId, e := commons.TryGetString(attributes, "parent_id")
	if nil != e {
		return nil, errors.New("'parent_id' is required, " + e.Error())
	}
	drvMgr, ok := ctx["drvMgr"].(commons.DriverManager)
	if !ok {
		return nil, errors.New("'drvMgr' is required, " + e.Error())
	}
	drv, _ := drvMgr.Connect("kpi")
	if nil == drv {
		return nil, errors.New("driver 'kpi' is required.")
	}

	job := &metricJob{metric: metric,
		params: map[string]string{"device_type": parentType, "device_id": parentId, "metric": metric},
		drv:    drv}

	job.Trigger, e = NewTrigger(attributes, func(t time.Time) { job.Run(t) }, ctx)
	return job, e
}

// func createRequest(nm string, attributes, ctx map[string]interface{}) (string, bytes.Buffer, error) {
// 	url, e := commons.TryGetString(ctx, "metric_url")
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
