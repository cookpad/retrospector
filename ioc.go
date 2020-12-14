package retrospector

type IOC struct {
	Value
	Source      string `json:"source" dynamo:"source"`
	UpdatedAt   int64  `json:"updated_at" dynamo:"updated_at"`
	Reason      string `json:"reason" dynamo:"reason"`
	Description string `json:"description" dynamo:"description"`
	Detected    bool   `json:"detected" dynamo:"detected"`
}

type IOCChunk []*IOC
