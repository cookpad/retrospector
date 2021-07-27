package retrospector

type Entity struct {
	Value
	Source      string `json:"source" dynamo:"source"`
	Subject     string `json:"subject" dynamo:"subject"`
	RecordedAt  int64  `json:"recorded_at" dynamo:"recorded_at"`
	Description string `json:"description" dynamo:"description"`
	Detected    bool   `json:"detected" dynamo:"detected"`
}
