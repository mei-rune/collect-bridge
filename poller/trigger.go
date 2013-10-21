package poller

import (
	"commons"
	"errors"
	"strings"
	"time"
)

type Trigger interface {
	Channel() <-chan time.Time
	Interval() time.Duration
	Close()
}

type TriggerBuilder interface {
	Interval() time.Duration
	New() (Trigger, error)
}

type triggerBuilder struct {
	interval time.Duration
}

func (self *triggerBuilder) Interval() time.Duration {
	return self.interval
}

func (self *triggerBuilder) New() (Trigger, error) {
	return &intervalTrigger{
		ticker:   time.NewTicker(self.interval),
		interval: self.interval}, nil
}

var (
	delay_start           = 0
	ExpressionSyntexError = errors.New("'expression' is error syntex")
)

const every = "@every "

func delay_interval() time.Duration {
	if is_test {
		return 0
	}

	// delay start trigger after the specific interval
	delay := delay_start % (5 * 60 * 1000)
	delay_start += 113
	return time.Duration(delay) * time.Millisecond
}

func newTrigger(attributes, ctx map[string]interface{}) (TriggerBuilder, error) {
	expression := commons.GetStringWithDefault(attributes, "expression", "")
	if "" == expression {
		return nil, commons.IsRequired("expression")
	}

	if strings.HasPrefix(expression, every) {
		interval, err := time.ParseDuration(expression[len(every):])
		if nil != err {
			return nil, errors.New(ExpressionSyntexError.Error() + ", " + err.Error())
		}

		return &triggerBuilder{interval: interval}, nil
	}
	return nil, ExpressionSyntexError
}

type intervalTrigger struct {
	ticker   *time.Ticker
	interval time.Duration
}

func (self *intervalTrigger) Channel() <-chan time.Time {
	return self.ticker.C
}

func (self *intervalTrigger) Interval() time.Duration {
	return self.interval
}

func (self *intervalTrigger) Close() {
	self.ticker.Stop()
}
