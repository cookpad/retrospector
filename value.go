package retrospector

type Value struct {
	Data string    `json:"value" dynamo:"value"`
	Type ValueType `json:"type" dynamo:"type"`
}

type ValueType string

const (
	ValueIPAddr         ValueType = "ipaddr"
	ValueDomainName     ValueType = "domain"
	ValueURL            ValueType = "url"
	ValueFileHashSha256 ValueType = "filehash.sha256"
)
