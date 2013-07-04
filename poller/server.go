package poller

import (
	"commons"
	"encoding/json"
	"fmt"
	"time"
)

type server struct {
	commons.SimpleServer

	jobs map[string]Job
}

func newServer(refresh time.Duration) *server {
	srv := &server{jobs: make(map[string]Job), SimpleServer: commons.SimpleServer{IdledInterval: refresh}}
	srv.OnIdle = srv.onIdle
	return srv
}

func (s *server) register(j Job) {
	s.jobs[j.Id()] = j
}

func (s *server) onIdle() {

}

func (s *server) String() string {
	return s.ReturnString(func() string {
		messages := make([]json.RawMessage, 0, len(s.jobs))
		for _, job := range s.jobs {
			messages = append(messages, json.RawMessage([]byte(job.Stats())))
		}

		s, e := json.MarshalIndent(messages, "", "  ")
		if nil != e {
			return e.Error()
		}

		return string(s)
	})
}
