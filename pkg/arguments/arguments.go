package arguments

import (
	"net/http"

	"github.com/Netflix/go-env"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/service"
)

// Arguments are passed to Handler. It includes environment variables, received event and factories, etc.
type Arguments struct {
	IOCTopicARN     string `env:"IOC_TOPIC_ARN"`
	RecordTableName string `env:"RECORD_TABLE_NAME"`
	SlackWebhookURL string `env:"SLACK_WEBHOOK_URL"`
	AwsRegion       string `env:"AWS_REGION"`
	SecretsARN      string `env:"SECRETS_ARN"`

	// Do not change them in each lambda Function. They must be accessed in only pkg/lambda
	Repository adaptor.Repository             `env:"-"`
	NewS3      adaptor.S3ClientFactory        `env:"-"`
	NewSNS     adaptor.SNSClientFactory       `env:"-"`
	NewSM      golambda.SecretsManagerFactory `env:"-"`
	HTTP       adaptor.HTTPClient             `env:"-"`
}

type Secrets struct {
	OTXToken string `json:"otx_token"`
}

// -----------------------
// Data binding

// New is constructor of Arguments
func New() *Arguments {
	args := &Arguments{}

	if _, err := env.UnmarshalFromEnviron(args); err != nil {
		golambda.Logger.With("err", err).Error("Failed env.UnmarshalFromEnviron")
		panic(err)
	}

	repo, err := adaptor.NewDynamoRepository(args.AwsRegion, args.RecordTableName)
	if err != nil {
		golambda.Logger.With("err", err).Error("Failed NewDynamoRepository")
		panic(err)
	}

	args.Repository = repo
	args.NewS3 = adaptor.NewS3Client
	args.NewSNS = adaptor.NewSNSClient

	return args
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

func (x *Arguments) GetSecrets() (*Secrets, error) {
	var secrets Secrets
	if err := golambda.GetSecretValuesWithFactory(x.SecretsARN, &secrets, x.NewSM); err != nil {
		return nil, golambda.WrapError(err).With("ARN", x.SecretsARN)
	}

	return &secrets, nil
}
