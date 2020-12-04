package main_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/lambda"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/m-mizutani/retrospector/pkg/service"
	"github.com/stretchr/testify/require"

	main "github.com/m-mizutani/retrospector/lambda/entityRecord"
)

func TestEntityRecord(t *testing.T) {
	// Setup test event
	sqsEvent, err := createSQSEvent("us-east-5", events.S3Entity{
		Bucket: events.S3Bucket{
			Name: "blue",
		},
		Object: events.S3Object{
			Key: "my/entity-object",
		},
	})
	require.NoError(t, err)

	// Setup mock
	newS3, _ := mock.NewS3Mock()
	s3Svc := service.NewEntityService(newS3)
	wq := s3Svc.NewWriteQueue("us-east-5", "blue", "my/entity-object")
	wq.Write(&retrospector.Entity{
		Value: retrospector.Value{
			Data: "five",
			Type: retrospector.ValueDomainName,
		},
		Source:     "timeless",
		RecordedAt: time.Now().Unix(),
	})
	require.NoError(t, wq.Close())

	repo := mock.NewRepository()

	args := &lambda.Arguments{
		Repository: repo,
		Event:      sqsEvent,
		NewS3:      newS3,
	}
	require.NoError(t, main.Handler(args))
}

func createSQSEvent(region string, s3entities events.S3Entity) (*events.SQSEvent, error) {
	event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				AWSRegion: region,
				S3:        s3entities,
			},
		},
	}
	rawEvent, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	snsEntity := events.SNSEntity{
		Message: string(rawEvent),
	}
	rawSNSEntity, err := json.Marshal(snsEntity)
	if err != nil {
		return nil, err
	}

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				Body: string(rawSNSEntity),
			},
		},
	}

	return &sqsEvent, nil
}
