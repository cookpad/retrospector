package main_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/m-mizutani/retrospector/lambda/iocRecord"
)

func TestIOCRecord(t *testing.T) {
	// Setup test event
	iocSet := retrospector.IOCChunk{
		{
			Value: retrospector.Value{
				Data: "blue",
				Type: retrospector.ValueDomainName,
			},
			Source:      "one",
			UpdatedAt:   12345,
			Reason:      "timeless",
			Description: "five",
		},
		{
			Value: retrospector.Value{
				Data: "orange",
				Type: retrospector.ValueDomainName,
			},
			Source:      "two",
			UpdatedAt:   12345,
			Reason:      "timeless",
			Description: "six",
		},
		{
			Value: retrospector.Value{
				Data: "red",
				Type: retrospector.ValueDomainName,
			},
			Source:      "three",
			UpdatedAt:   12345,
			Reason:      "timeless",
			Description: "seven",
		},
	}

	keyEntities := []*retrospector.Entity{
		{
			Value: retrospector.Value{
				Data: "blue",
				Type: retrospector.ValueDomainName,
			},
		},
		{
			Value: retrospector.Value{
				Data: "orange",
				Type: retrospector.ValueDomainName,
			},
		},
		{
			Value: retrospector.Value{
				Data: "red",
				Type: retrospector.ValueDomainName,
			},
		},
	}

	t.Run("handle multiple ioc in SQS", func(t *testing.T) {
		rawEvent, err := json.Marshal(iocSet)
		require.NoError(t, err)

		snsEntity := events.SNSEntity{
			Message: string(rawEvent),
		}
		rawSNSEntity, err := json.Marshal(snsEntity)
		require.NoError(t, err)

		sqsEvent := events.SQSEvent{
			Records: []events.SQSMessage{
				{
					Body: string(rawSNSEntity),
				},
			},
		}

		repo := mock.NewRepository()
		resp0, err := repo.GetIOCSet(keyEntities)
		assert.NoError(t, err)
		assert.Empty(t, resp0)

		args := &arguments.Arguments{
			Repository: repo,
		}
		event := golambda.Event{
			Origin: sqsEvent,
		}
		_, err = main.Handler(args, event)
		require.NoError(t, err)

		resp, err := repo.GetIOCSet(keyEntities)
		require.NoError(t, err)
		assert.Equal(t, 3, len(resp))
		assert.Contains(t, iocSet, resp[0])
		assert.Contains(t, iocSet, resp[1])
		assert.Contains(t, iocSet, resp[2])
	})

	t.Run("handle multiple SQS messages", func(t *testing.T) {
		sqsEvent := events.SQSEvent{}

		for _, ioc := range iocSet {
			rawEvent, err := json.Marshal(retrospector.IOCChunk{ioc})
			require.NoError(t, err)

			snsEntity := events.SNSEntity{
				Message: string(rawEvent),
			}
			rawSNSEntity, err := json.Marshal(snsEntity)
			require.NoError(t, err)

			sqsEvent.Records = append(sqsEvent.Records, events.SQSMessage{
				Body: string(rawSNSEntity),
			})
		}

		repo := mock.NewRepository()
		resp0, err := repo.GetIOCSet(keyEntities)
		assert.NoError(t, err)
		assert.Empty(t, resp0)

		args := &arguments.Arguments{
			Repository: repo,
		}
		event := golambda.Event{Origin: sqsEvent}
		_, err = main.Handler(args, event)
		require.NoError(t, err)

		resp, err := repo.GetIOCSet(keyEntities)
		require.NoError(t, err)
		assert.Equal(t, 3, len(resp))
		assert.Contains(t, iocSet, resp[0])
		assert.Contains(t, iocSet, resp[1])
		assert.Contains(t, iocSet, resp[2])
	})

}
