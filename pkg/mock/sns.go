package mock

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
)

// SNSClient is mock SNS client
type SNSClient struct {
	Region       string
	PublishInput []*sns.PublishInput
}

// Publish is mock of SNS.Publish
func (x *SNSClient) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	x.PublishInput = append(x.PublishInput, input)
	return &sns.PublishOutput{}, nil
}

// NewSNSMock returns SNSClientFactory and mock.SNSClient that SNSClientFactory returns
func NewSNSMock() (adaptor.SNSClientFactory, *SNSClient) {
	client := &SNSClient{}
	return func(region string) (adaptor.SNSClient, error) {
		client.Region = region
		return client, nil
	}, client
}
