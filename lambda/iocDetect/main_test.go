package main_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/m-mizutani/retrospector/lambda/iocDetect"
)

func TestIOCDetect(t *testing.T) {
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

	t.Run("detect one entity", func(t *testing.T) {
		httpClient := &mock.HTTPClient{
			RespCode: http.StatusOK,
			RespBody: ioutil.NopCloser(strings.NewReader("")),
		}

		repo := mock.NewRepository()
		require.NoError(t, repo.PutEntities([]*retrospector.Entity{
			{
				Value: retrospector.Value{
					Data: "blue",
					Type: retrospector.ValueDomainName,
				},
			},
		}))

		args := &arguments.Arguments{
			Repository:      repo,
			HTTP:            httpClient,
			SlackWebhookURL: "https://test.example.com/slack",
		}
		event := golambda.Event{Origin: sqsEvent}
		_, err := main.Handler(args, event)
		require.NoError(t, err)
		require.Equal(t, 1, len(httpClient.Requests))
		assert.Equal(t, "test.example.com", httpClient.Requests[0].URL.Host)
		assert.Equal(t, "/slack", httpClient.Requests[0].URL.Path)
	})

	t.Run("not detect any entity by data", func(t *testing.T) {
		httpClient := &mock.HTTPClient{
			RespCode: http.StatusOK,
			RespBody: ioutil.NopCloser(strings.NewReader("")),
		}

		repo := mock.NewRepository()
		require.NoError(t, repo.PutEntities([]*retrospector.Entity{
			{
				Value: retrospector.Value{
					Data: "green",
					Type: retrospector.ValueDomainName,
				},
			},
		}))

		args := &arguments.Arguments{
			Repository:      repo,
			HTTP:            httpClient,
			SlackWebhookURL: "https://test.example.com/slack",
		}
		event := golambda.Event{Origin: sqsEvent}
		_, err := main.Handler(args, event)
		require.NoError(t, err)

		require.Equal(t, 0, len(httpClient.Requests))
	})

	t.Run("not detect any entity by data", func(t *testing.T) {
		httpClient := &mock.HTTPClient{
			RespCode: http.StatusOK,
			RespBody: ioutil.NopCloser(strings.NewReader("")),
		}

		repo := mock.NewRepository()
		require.NoError(t, repo.PutEntities([]*retrospector.Entity{
			{
				Value: retrospector.Value{
					Data: "blue",
					Type: retrospector.ValueIPAddr,
				},
			},
		}))

		event := golambda.Event{Origin: sqsEvent}
		args := &arguments.Arguments{
			Repository:      repo,
			HTTP:            httpClient,
			SlackWebhookURL: "https://test.example.com/slack",
		}
		_, err := main.Handler(args, event)
		require.NoError(t, err)
		require.Equal(t, 0, len(httpClient.Requests))
	})

}
