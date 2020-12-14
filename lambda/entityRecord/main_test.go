package main_test

import (
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/m-mizutani/retrospector/pkg/service"
	"github.com/stretchr/testify/require"

	main "github.com/m-mizutani/retrospector/lambda/entityRecord"
)

func TestEntityRecord(t *testing.T) {
	// Setup test event
	s3Event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				AWSRegion: "us-east-5",
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: "blue",
					},
					Object: events.S3Object{
						Key: "my/entity-object",
					},
				},
			},
		},
	}
	var event golambda.Event
	require.NoError(t, event.EncapSNSonSQSMessage(s3Event))

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

	args := &arguments.Arguments{
		Repository: repo,
		NewS3:      newS3,
	}
	_, err := main.Handler(args, event)
	require.NoError(t, err)
}
