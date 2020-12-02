package retrospector

type IOC struct {
	Value
	Source    string `json:"source" dynamo:"source"`
	UpdatedAt int64  `json:"updated_at" dynamo:"updated_at"`
}
