package poller

type server struct {
	jobs map[string]Job
}

func (s *server) register(job Job) {
	s.jobs[job.Id()] = job
}
