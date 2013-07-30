package sampling

import (
	"sync"
	"time"
)

type baseWorker struct {
	timestamp int64
	l         sync.Mutex

	dispatcher  *dispatcher
	metric_name string
	ctx         *context
	invoke      invoke_func
}

func (self *baseWorker) isTimeout(now int64, default_interval time.Duration) bool {
	self.l.Lock()
	defer self.l.Unlock()

	return (self.timestamp + int64(default_interval.Seconds())) <= now
}

// func (self *baseWorker) get() bool {
// 	res := self.invoke(self.dispatcher, self.metric_name, self.ctx)
//   return
// }

// type flux_worker struct {
// 	baseWorker
// }

// func (self *flux_worker) get() (interface{}, error) {
// }
