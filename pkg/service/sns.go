package service

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/errors"
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

func extractSNSRegion(topicARN string) (string, *errors.Error) {
	// topicARN sample: arn:aws:sns:us-east-1:111122223333:my-topic
	arnParts := strings.Split(topicARN, ":")

	if len(arnParts) != 6 {
		return "", errors.New("Invalid SNS topic ARN").With("ARN", topicARN)
	}

	return arnParts[3], nil
}

// Publish is wrapper of sns:Publish of AWS
func (x *SNSService) Publish(topicARN string, msg interface{}) error {
	region, daErr := extractSNSRegion(topicARN)
	if daErr != nil {
		return daErr
	}

	client, err := x.newSNS(region)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal message").With("msg", msg)
	}

	input := sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(string(raw)),
	}
	resp, err := client.Publish(&input)

	if err != nil {
		return errors.Wrap(err, "Fail to send SQS message").With("input", input)
	}

	logger.Trace().Interface("resp", resp).Msg("Sent SQS message")

	return nil
}