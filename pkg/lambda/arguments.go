package lambda

import (
	"encoding/json"

	"github.com/Netflix/go-env"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/errors"
	"github.com/m-mizutani/retrospector/pkg/service"
)

// Arguments are passed to Handler. It includes environment variables, received event and factories, etc.
type Arguments struct {
	iocTopicARN     string `env:"IOC_TOPIC_ARN"`
	recordTableName string `env:"RECORD_TABLE_NAME"`
	awsRegion       string `env:"AWS_REGION"`

	Repository adaptor.Repository
	NewS3      adaptor.S3ClientFactory
	NewSNS     adaptor.SNSClientFactory

	Event interface{}
}

func newArguments(event interface{}) (*Arguments, error) {
	args := &Arguments{
		Event: event,
	}
	if err := args.bindEnv(); err != nil {
		return nil, err
	}
	repo, err := adaptor.NewDynamoRepository(args.awsRegion, args.recordTableName)
	if err != nil {
		return nil, err
	}
	args.Repository = repo
	args.NewS3 = adaptor.NewS3Client
	args.NewSNS = adaptor.NewSNSClient

	return args, nil
}

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
		return errors.Wrap(err, "Unmarhal lambda event")
	}
	return nil
}

// RepositoryService returns *service.RepositoryService created from Arguments.Repository
func (x *Arguments) RepositoryService() *service.RepositoryService {
	return service.NewRepositoryService(x.Repository)
}
