package service

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/logging"
)

var logger = logging.Logger

// SNSService is accessor to SQS
type SNSService struct {
	newSNS adaptor.SNSClientFactory
}

// NewSNSService is constructor of
func NewSNSService(newSNS adaptor.SNSClientFactory) *SNSService {
	return &SNSService{
		newSNS: newSNS,
	}
}

func extractSNSRegion(topicARN string) (string, error) {
	// topicARN sample: arn:aws:sns:us-east-1:111122223333:my-topic
	arnParts := strings.Split(topicARN, ":")

	if len(arnParts) != 6 {
		return "", golambda.NewError("Invalid SNS topic ARN").With("ARN", topicARN)
	}

	return arnParts[3], nil
}

func publishSNS(client adaptor.SNSClient, topicARN string, msg interface{}) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return golambda.WrapError(err, "Fail to marshal message").With("msg", msg)
	}

	input := sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(string(raw)),
	}
	resp, err := client.Publish(&input)

	if err != nil {
		return golambda.WrapError(err, "Fail to send SQS message").With("input", input)
	}

	logger.Trace().Interface("resp", resp).Msg("Sent SQS message")

	return nil
}

// PublishIOC is wrapper of sns:Publish of AWS for IOCChunk
func (x *SNSService) PublishIOC(topicARN string, chunk retrospector.IOCChunk) error {
	region, err := extractSNSRegion(topicARN)
	if err != nil {
		return err
	}

	client, err := x.newSNS(region)
	if err != nil {
		return err
	}

	const step = 32

	for i := 0; i < len(chunk); i += step {
		e := i + step
		if len(chunk) < e {
			e = len(chunk)
		}
		c := chunk[i:e]
		if err := publishSNS(client, topicARN, c); err != nil {
			return golambda.WrapError(err).With("chunk", c)
		}
	}

	return nil
}
