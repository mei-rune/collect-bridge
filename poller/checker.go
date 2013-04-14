package poller

type Checker interface {
	Run(value interface{}, res map[string]interface{}) bool
}
