package sampling

import (
	"sync"
)

type runningRequests struct {
	l        sync.Mutex
	requests map[string]int64
}

func (self *runningRequests) get(nm string) int64 {
	self.l.Lock()
	defer self.l.Unlock()
	return self.requests[nm]
}

func (self *runningRequests) put(nm string, t int64) {
	self.l.Lock()
	defer self.l.Unlock()
	self.requests[nm] = t
}
func (self *runningRequests) remove(nm string) {
	self.l.Lock()
	defer self.l.Unlock()
	delete(self.requests, nm)
}

func (self *runningRequests) clearExpired(now int64) {
	self.l.Lock()
	defer self.l.Unlock()
	expired := make([]string, 0, 10)
	for k, v := range self.requests {
		if now-v > 10*60 {
			expired = append(expired, k)
		}
	}

	for _, k := range expired {
		delete(self.requests, k)
	}
}
