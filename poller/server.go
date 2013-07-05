package poller

import (
	"bytes"
	"commons"
	"ds"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/runner-mei/go-restful"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type errorJob struct {
	id, name, e string

	updated_at time.Time
}

func (self *errorJob) Start() error {
	return nil
}

func (self *errorJob) Stop() {
}

func (self *errorJob) Id() string {
	return self.id
}

func (self *errorJob) Name() string {
	return self.name
}

func (self *errorJob) Stats() map[string]interface{} {
	return map[string]interface{}{
		"id":    self.id,
		"name":  self.name,
		"error": self.e}
}

func (self *errorJob) Version() time.Time {
	return self.updated_at
}

type server struct {
	commons.SimpleServer

	jobs       map[string]Job
	client     *ds.Client
	ctx        map[string]interface{}
	last_error error
}

func newServer(refresh time.Duration, client *ds.Client, ctx map[string]interface{}) *server {
	srv := &server{SimpleServer: commons.SimpleServer{Interval: refresh},
		jobs:   make(map[string]Job),
		client: client,
		ctx:    ctx}

	srv.OnTimeout = srv.onIdle
	srv.OnStart = srv.onStart
	srv.OnStop = srv.onStop
	return srv
}

func (s *server) startJob(attributes map[string]interface{}) {
	name := commons.GetStringWithDefault(attributes, "name", "unknow_name")
	id := commons.GetStringWithDefault(attributes, "id", "unknow_id")

	job, e := newJob(attributes, s.ctx)
	if nil != e {
		updated_at, _ := commons.GetTime(attributes, "updated_at")
		msg := fmt.Sprintf("create '%v:%v' failed, %v\n", id, name, e)
		job = &errorJob{id: id, name: name, e: msg, updated_at: updated_at}
		log.Print(msg)
		goto end
	}

	e = job.Start()
	if nil != e {
		updated_at, _ := commons.GetTime(attributes, "updated_at")
		msg := fmt.Sprintf("start '%v:%v' failed, %v\n", id, name, e)
		job = &errorJob{id: id, name: name, e: msg, updated_at: updated_at}
		log.Print(msg)
		goto end
	}

	log.Printf("load '%v:%v' is ok\n", id, name)
end:
	s.jobs[job.Id()] = job
}

func (s *server) loadJob(id string) {
	attributes, e := s.client.FindByIdWithIncludes("trigger", id, "action")
	if nil != e {
		msg := "load trigger '" + id + "' from db failed," + e.Error()
		job := &errorJob{id: id, e: msg}
		s.jobs[job.Id()] = job
		log.Print(msg)
		return
	}

	s.startJob(attributes)
}

func (s *server) stopJob(id string) {
	job, ok := s.jobs[id]
	if !ok {
		return
	}
	job.Stop()
	delete(s.jobs, id)
}

func (s *server) onStart() error {
	results, err := s.client.FindByWithIncludes("trigger", map[string]string{}, "action")
	if nil != err {
		return errors.New("load triggers from db failed," + err.Error())
	}

	for _, attributes := range results {
		s.startJob(attributes)
	}

	return nil
}

func (s *server) onStop() {
	for _, job := range s.jobs {
		job.Stop()
	}
	s.jobs = make(map[string]Job)
}

func (s *server) onIdle() {
	new_snapshots, e := s.client.Snapshot("trigger", map[string]string{})
	if nil == e {
		s.last_error = e
		return
	}
	s.last_error = nil

	old_snapshots := map[string]*ds.RecordVersion{}

	for id, job := range s.jobs {
		old_snapshots[id] = &ds.RecordVersion{UpdatedAt: job.Version()}
	}

	newed, updated, deleted := ds.Diff(new_snapshots, old_snapshots)
	if nil != newed {
		for _, id := range newed {
			s.loadJob(id)
		}
	}

	if nil != updated {
		for _, id := range updated {
			s.stopJob(id)
			s.loadJob(id)
		}
	}

	if nil != deleted {
		for _, id := range deleted {
			s.stopJob(id)
		}
	}
}

func (s *server) String() string {
	return s.ReturnString(func() string {
		messages := make([]interface{}, 0, len(s.jobs))
		for _, job := range s.jobs {
			messages = append(messages, job.Stats())
		}
		if nil != s.last_error {
			messages = append(messages, map[string]string{
				"name":       "self",
				"last_error": s.last_error.Error()})
		}

		s, e := json.MarshalIndent(messages, "", "  ")
		if nil != e {
			return e.Error()
		}

		return string(s)
	})
}

func (s *server) wrap(req *restful.Request, resp *restful.Response, cb func()) {
	s.NotReturn(func() {
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
				resp.WriteErrorString(http.StatusInternalServerError, buffer.String())
			}
		}()
		cb()
	})
}

func (s *server) Sync(req *restful.Request, resp *restful.Response) {
	s.wrap(req, resp, func() {
		s.onIdle()
		if nil == s.last_error {
			resp.Write([]byte("ok"))
		} else {
			resp.WriteErrorString(http.StatusInternalServerError, s.last_error.Error())
		}
	})
}

func (s *server) StatsAll(req *restful.Request, resp *restful.Response) {
	s.wrap(req, resp, func() {
		messages := make([]interface{}, 0, len(s.jobs))
		for _, job := range s.jobs {
			messages = append(messages, job.Stats())
		}

		if nil != s.last_error {
			messages = append(messages, map[string]string{
				"id":    "0",
				"name":  "self",
				"error": s.last_error.Error()})
		}

		resp.WriteAsJson(messages)
	})
}

func (s *server) StatsById(req *restful.Request, resp *restful.Response) {
	s.wrap(req, resp, func() {
		id := req.PathParameter("id")
		if 0 == len(id) {
			resp.WriteErrorString(http.StatusBadRequest, commons.IsRequired("id").Error())
			return
		}

		job, ok := s.jobs[id]
		if !ok {
			resp.WriteErrorString(http.StatusNotFound, "not found")
			return
		}
		resp.WriteAsJson(job.Stats())
	})
}

func (s *server) StatsByName(req *restful.Request, resp *restful.Response) {
	s.wrap(req, resp, func() {
		name := req.PathParameter("name")
		if 0 == len(name) {
			resp.WriteErrorString(http.StatusBadRequest, commons.IsRequired("name").Error())
			return
		}

		messages := make([]interface{}, 0, len(s.jobs))
		for _, job := range s.jobs {
			if strings.Contains(job.Name(), name) {
				messages = append(messages, job.Stats())
			}
		}
		resp.WriteAsJson(messages)
	})
}

func (s *server) StatsByAddress(req *restful.Request, resp *restful.Response) {
	s.wrap(req, resp, func() {
		address := req.PathParameter("address")
		if 0 == len(address) {
			resp.WriteErrorString(http.StatusBadRequest, commons.IsRequired("address").Error())
			return
		}

		results, e := s.client.FindByWithIncludes("managed_object", map[string]string{"address": address}, "trigger")
		if nil != e {
			resp.WriteErrorString(http.StatusInternalServerError, e.Error())
			return
		}

		id_list := make([]string, 0, 10)

		for _, result := range results {
			triggers, e := commons.GetObjects(result, "$trigger")
			if nil != e {
				continue
			}
			for _, trigger := range triggers {
				id_list = append(id_list, commons.GetStringWithDefault(trigger, "id", ""))
			}
		}

		messages := make([]interface{}, 0, len(id_list))
		for _, id := range id_list {
			if 0 == len(id) {
				continue
			}

			if job, ok := s.jobs[id]; ok {
				messages = append(messages, job.Stats())
			} else {
				messages = append(messages, map[string]string{"id": id, "name": "unknow", "status": "not found in the jobs"})
			}
		}
		resp.WriteAsJson(messages)
	})
}
