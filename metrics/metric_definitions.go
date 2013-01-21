package metrics

type MetricDefinition struct {
	Level      []string          `json:"level"`
	Name       string            `json:"name"`
	Method     string            `json:"method"`
	File       string            `json:"file"`
	Action     map[string]string `json:"action"`
	Match      []Filter          `json:"match"`
	Categories []string          `json:"categories"`
}

type Filter struct {
	Method    string   `json:"method"`
	Arguments []string `json:"arguments"`
	//Value     string   `json:"value"`
}
