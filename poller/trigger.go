package poller

import (
	"commons"
	"errors"
	"strings"
	"time"
)

type Trigger interface {
	GetChannel() <-chan time.Time
	Close()
}

type TriggerBuilder interface {
	New() (Trigger, error)
}

type TriggerBuilderFunc func() (Trigger, error)

func (tb TriggerBuilderFunc) New() (Trigger, error) {
	return tb()
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

		return TriggerBuilderFunc(func() (Trigger, error) {
			return &intervalTrigger{
				ticker:   time.NewTicker(interval),
				interval: interval}, nil
		}), nil
	}
	return nil, ExpressionSyntexError
}

type intervalTrigger struct {
	ticker   *time.Ticker
	interval time.Duration
}

func (self *intervalTrigger) GetChannel() <-chan time.Time {
	return self.ticker.C
}

func (self *intervalTrigger) Close() {
	self.ticker.Stop()
}
