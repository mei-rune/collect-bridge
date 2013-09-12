package poller

import (
	"bytes"
	"commons"
	ds "data_store"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var loadThreshold = 20

type server struct {
	commons.SimpleServer

	jobs       map[string]Job
	client     *ds.Client
	ctx        map[string]interface{}
	firedAt    time.Time
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
	clazz := commons.GetStringWithDefault(attributes, "type", "unknow_type")
	name := commons.GetStringWithDefault(attributes, "name", "unknow_name")
	id := commons.GetStringWithDefault(attributes, "id", "unknow_id")

	job, e := newJob(attributes, s.ctx)
	if nil != e {
		updated_at, _ := commons.GetTime(attributes, "updated_at")
		msg := fmt.Sprintf("create '%v:%v' failed, %v\n", id, name, e)
		job = &errorJob{clazz: clazz, id: id, name: name, e: msg, updated_at: updated_at}
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

func (s *server) stopJob(id string, reason int) {
	job, ok := s.jobs[id]
	if !ok {
		return
	}
	job.Close(reason)
	delete(s.jobs, id)
	log.Println("stop trigger with id was '" + id + "'")
}

func (s *server) onStart() error {
	results, err := s.client.FindByWithIncludes("trigger", map[string]string{}, "action")
	if nil != err {
		return errors.New("load triggers from db failed," + err.Error())
	}

	id2results := map[int]map[string]interface{}{}
	for _, attributes := range results {
		id := commons.GetIntWithDefault(attributes, "id", 0)
		if 0 == id {
			return errors.New("'id' of trigger is 0?")
		}

		id2results[id] = attributes
	}

	var client *commons.Client
	var all_cookies map[int64]map[string]interface{}
	if *load_cookies {
		client = commons.NewClient(*foreignUrl, "alert_cookies")
		loader := &cookiesLoaderImpl{client: client}

		if e := loader.init(); nil != e {
			return e
		}
		all_cookies = loader.id2cookies
		s.ctx["cookies_loader"] = loader
		defer delete(s.ctx, "cookies_loader")
	} else {
		delete(s.ctx, "cookies_loader")
	}

	for _, attributes := range results {
		s.startJob(attributes)
	}

	if nil != all_cookies {
		for action_id, _ := range all_cookies {
			action_id_str := strconv.FormatInt(int64(action_id), 10)
			log.Println("load alert cookies with id was " + action_id_str + " is failed, action is deleted.")
			dres := client.Delete(map[string]string{"id": "@" + action_id_str})
			if dres.HasError() {
				log.Println("delete alert cookies with id was " + action_id_str + " is failed, " + dres.ErrorMessage())
			}
		}
	}

	return nil
}

func (s *server) onStop() {
	for _, job := range s.jobs {
		job.Close(CLOSE_REASON_NORMAL)
	}
	s.jobs = make(map[string]Job)

	log.Println("server is stopped.")
}

func (s *server) onIdle() {
	s.firedAt = time.Now()
	new_snapshots, e := s.client.Snapshot("trigger", map[string]string{})
	if nil != e {
		s.last_error = e
		log.Println("[srv] poll failed,", e)
		return
	}
	s.last_error = nil

	old_snapshots := map[string]*ds.RecordVersion{}

	for id, job := range s.jobs {
		old_snapshots[id] = &ds.RecordVersion{UpdatedAt: job.Version()}
	}

	newed, updated, deleted := ds.Diff(new_snapshots, old_snapshots)

	if *load_cookies {
		var should_load []string = nil
		if nil != newed {
			should_load = append(should_load, newed...)
		}
		if nil != updated {
			should_load = append(should_load, updated...)
		}

		if nil != should_load && 0 != len(should_load) {

			started_at := time.Now()
			loader := &cookiesLoaderImpl{client: commons.NewClient(*foreignUrl, "alert_cookies")}
			if loadThreshold < len(should_load) {
				if e := loader.init(); nil != e {
					s.last_error = e
					log.Println("[srv] load cookies failed,", e)
					return
				}
			} else {
				loader.loadFromWebWhileNotFound = true
				loader.isPersist = true
			}

			s.ctx["cookies_loader"] = loader
			defer delete(s.ctx, "cookies_loader")

			log.Println("[srv] load ", len(loader.id2cookies), " cookies of ", len(should_load), " trigger and ", time.Now().Sub(started_at), "is elapsed")
		}
	} else {
		delete(s.ctx, "cookies_loader")
	}

	if nil != newed {
		for _, id := range newed {
			s.loadJob(id)
		}
		log.Println("[srv] new triggers with count is", len(newed), "is started.")
	}

	if nil != updated {
		for _, id := range updated {
			s.stopJob(id, CLOSE_REASON_NORMAL)
			s.loadJob(id)
		}
		log.Println("[srv] updated triggers with count is", len(updated), "is started.")
	}

	if nil != deleted {
		for _, id := range deleted {
			s.stopJob(id, CLOSE_REASON_DELETED)
		}
		log.Println("[srv] deleted triggers with count is", len(deleted), "is started.")
	}

	log.Println("[srv] poll is ok.")
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
				"firedAt":    s.firedAt.Format(time.RFC3339Nano),
				"status":     s.StatusString(),
				"last_error": s.last_error.Error()})
		} else {
			messages = append(messages, map[string]string{
				"name":    "self",
				"firedAt": s.firedAt.Format(time.RFC3339Nano),
				"status":  s.StatusString()})
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

		results, e := s.client.FindByWithIncludes("network_device", map[string]string{"address": address}, "metric_trigger")
		if nil != e {
			resp.WriteErrorString(http.StatusInternalServerError, e.Error())
			return
		}

		id_list := make([]string, 0, 10)

		for _, result := range results {
			triggers, e := commons.GetObjects(result, "$metric_trigger")
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
