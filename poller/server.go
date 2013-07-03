package poller

import (
	"commons"
)

type server struct {
	commons.SimpleServer

	jobs map[string]Job
}

func (s *server) register(j Job) {
	s.jobs[j.Id()] = j
}

func (s *server) onIdle() {

}
