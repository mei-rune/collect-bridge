package poller

import (
	"commons"
	"ds"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

func (self *errorJob) Stats() string {
	return self.e
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
			messages = append(messages, json.RawMessage([]byte(job.Stats())))
		}
		if nil != s.last_error {
			messages = append(messages, map[string]string{
				"name":       "server",
				"last_error": s.last_error.Error()})
		}

		s, e := json.MarshalIndent(messages, "", "  ")
		if nil != e {
			return e.Error()
		}

		return string(s)
	})
}

func (s *server) Sync() string {
	return s.ReturnString(func() string {
		s.onIdle()
		if nil == s.last_error {
			return ""
		} else {
			return s.last_error.Error()
		}
	})
}
