package lambda

import (
	"encoding/json"
	"net/http"

	"github.com/Netflix/go-env"
	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/errors"
	"github.com/m-mizutani/retrospector/pkg/service"
)

// Arguments are passed to Handler. It includes environment variables, received event and factories, etc.
type Arguments struct {
	IOCTopicARN     string `env:"IOC_TOPIC_ARN"`
	RecordTableName string `env:"RECORD_TABLE_NAME"`
	SlackWebhookURL string `env:"SLACK_WEBHOOK_URL"`
	AwsRegion       string `env:"AWS_REGION"`

	// Do not change them in each lambda Function. They must be accessed in only pkg/lambda
	Repository adaptor.Repository       `env:"-"`
	NewS3      adaptor.S3ClientFactory  `env:"-"`
	NewSNS     adaptor.SNSClientFactory `env:"-"`
	HTTP       adaptor.HTTPClient       `env:"-"`

	Event interface{} `env:"-"`
}

func newArguments(event interface{}) (*Arguments, error) {
	args := &Arguments{
		Event: event,
	}
	if err := args.bindEnv(); err != nil {
		return nil, err
	}
	repo, err := adaptor.NewDynamoRepository(args.AwsRegion, args.RecordTableName)
	if err != nil {
		return nil, err
	}
	args.Repository = repo
	args.NewS3 = adaptor.NewS3Client
	args.NewSNS = adaptor.NewSNSClient

	return args, nil
}

// -----------------------
// Services

// RepositoryService returns *service.RepositoryService created from Arguments.Repository
func (x *Arguments) RepositoryService() *service.RepositoryService {
	return service.NewRepositoryService(x.Repository)
}

// SNSService returns a new *service.SNSService based on Arguments.IOCTopicARN
func (x *Arguments) SNSService() *service.SNSService {
	factory := x.NewSNS
	if factory == nil {
		factory = adaptor.NewSNSClient
	}
	return service.NewSNSService(factory)
}

func (x *Arguments) HTTPClient() adaptor.HTTPClient {
	client := x.HTTP
	if client == nil {
		client = &http.Client{}
	}
	return client
}

func (x *Arguments) EntityService() *service.EntityService {
	newS3 := x.NewS3
	if newS3 == nil {
		newS3 = adaptor.NewS3Client
	}
	return service.NewEntityService(newS3)
}

func (x *Arguments) AlertService() *service.AlertService {
	httpClient := x.HTTPClient()
	return service.NewAlertService(&service.AlertServiceArguments{
		HTTPClient:              httpClient,
		SlackIncomingWebhookURL: x.SlackWebhookURL,
	})
}

// -----------------------
// Data binding

func (x *Arguments) bindEnv() error {
	if _, err := env.UnmarshalFromEnviron(x); err != nil {
		return errors.Wrap(err, "Unmarshal environ vars")
	}
	return nil
}

// BindEvent convert event that Lambda Function received to v via json marshal/unmarshal
func (x *Arguments) BindEvent(v interface{}) error {
	raw, err := json.Marshal(x.Event)
	if err != nil {
		return errors.Wrap(err, "Marshal lambda event")
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return errors.Wrap(err, "Unmarshal lambda event")
	}
	return nil
}

// EventRecord is decapsulate event data (e.g. Body of SQS event)
type EventRecord []byte

// Bind unmarshal event record to object
func (x EventRecord) Bind(ev interface{}) error {
	if err := json.Unmarshal(x, ev); err != nil {
		return errors.Wrap(err, "Failed json.Unmarshal in DecodeEvent").With("raw", string(x))
	}
	return nil
}

// DecapSQSEvent decapsulate wrapped body data in SQSEvent
func (x *Arguments) DecapSQSEvent() ([]EventRecord, error) {
	var sqsEvent events.SQSEvent
	if err := x.BindEvent(&sqsEvent); err != nil {
		return nil, err
	}

	var output []EventRecord
	for _, record := range sqsEvent.Records {
		output = append(output, EventRecord(record.Body))
	}

	return output, nil
}

// DecapSNSoverSQSEvent decapsulate wrapped body data in SQSEvent
func (x *Arguments) DecapSNSoverSQSEvent() ([]EventRecord, error) {
	var sqsEvent events.SQSEvent
	if err := x.BindEvent(&sqsEvent); err != nil {
		return nil, err
	}

	var output []EventRecord
	for _, record := range sqsEvent.Records {
		var snsEntity events.SNSEntity
		if err := json.Unmarshal([]byte(record.Body), &snsEntity); err != nil {
			return nil, errors.Wrap(err, "Failed to unmarshal SNS entity in SQS msg").With("body", record.Body)
		}

		output = append(output, EventRecord(snsEntity.Message))
	}

	return output, nil
}

// DecapSNSEvent decapsulate wrapped body data in SNSEvent
func (x *Arguments) DecapSNSEvent() ([]EventRecord, error) {
	var snsEvent events.SNSEvent
	if err := x.BindEvent(&snsEvent); err != nil {
		return nil, err
	}

	var output []EventRecord
	for _, record := range snsEvent.Records {
		output = append(output, EventRecord(record.SNS.Message))
	}

	return output, nil
}
