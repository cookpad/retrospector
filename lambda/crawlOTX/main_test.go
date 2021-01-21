package main_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/m-mizutani/retrospector/lambda/crawlOTX"
)

type dummySM struct {
	req []*secretsmanager.GetSecretValueInput
}

func (x *dummySM) GetSecretValue(req *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	x.req = append(x.req, req)
	return &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(`{"otx_token":"blue-token"}`),
	}, nil
}

func TestCrawlOTX(t *testing.T) {
	sampleData := `{
	"results": [
		{
		"id": 2711680570,
		"indicator": "example.org",
		"type": "hostname",
		"title": null,
		"description": null,
		"content": ""
		},
		{
		"id": 2724696685,
		"indicator": "10.1.2.3",
		"type": "IPv4",
		"title": null,
		"description": null,
		"content": ""
		}
	],
	"count": 2,
	"previous": "https://otx.alienvault.com/api/v1/indicators/export?modified_since=2020-12-10T00%3A00%3A00+00%3A00&page=2",
	"next": "https://otx.alienvault.com/api/v1/indicators/export?modified_since=2020-12-10T00%3A00%3A00+00%3A00&page=4"
}`

	newSNS, snsClient := mock.NewSNSMock()
	httpClient := &mock.HTTPClient{
		RespCode: http.StatusOK,
		RespBody: ioutil.NopCloser(strings.NewReader(sampleData)),
	}

	var smClient dummySM
	args := &arguments.Arguments{
		IOCTopicARN: "arn:aws:sns:us-east-1:111122223333:my-topic",
		NewSNS:      newSNS,
		HTTP:        httpClient,
		NewSM:       func(region string) (golambda.SecretsManagerClient, error) { return &smClient, nil },
		SecretsARN:  "arn:aws:secretsmanager:ap-northeast-1:111122223333:secret:orange",
	}

	_, err := main.Handler(args, golambda.Event{})
	require.NoError(t, err)

	require.Equal(t, 1, len(snsClient.PublishInput))
	assert.Equal(t, "us-east-1", snsClient.Region)
	assert.Equal(t, "arn:aws:sns:us-east-1:111122223333:my-topic", *snsClient.PublishInput[0].TopicArn)

	require.Equal(t, 1, len(smClient.req))
	assert.Equal(t, "arn:aws:secretsmanager:ap-northeast-1:111122223333:secret:orange", *smClient.req[0].SecretId)

	var iocChunk retrospector.IOCChunk
	require.NoError(t, json.Unmarshal([]byte(*snsClient.PublishInput[0].Message), &iocChunk))
	require.Equal(t, 2, len(iocChunk))
	iocValues := []string{"example.org", "10.1.2.3"}
	assert.Contains(t, iocValues, iocChunk[0].Data)
	assert.Contains(t, iocValues, iocChunk[1].Data)
}

func TestCrawlOTXIntegration(t *testing.T) {
	if _, ok := os.LookupEnv("ENABLE_INTEGRATION_TEST"); !ok {
		t.Skip("ENABLE_INTEGRATION_TEST is not set")
	}
	secretsARN, ok := os.LookupEnv("SECRETS_ARN")
	if !ok {
		t.Skip("SECRETS_ARN is not set")
	}

	newSNS, client := mock.NewSNSMock()
	args := &arguments.Arguments{
		IOCTopicARN: "arn:aws:sns:us-east-1:111122223333:my-topic",
		NewSNS:      newSNS,
		SecretsARN:  secretsARN,
	}

	_, err := main.Handler(args, golambda.Event{})
	require.NoError(t, err)
	assert.Greater(t, len(client.PublishInput), 0)
	assert.Equal(t, "us-east-1", client.Region)
	assert.Equal(t, "arn:aws:sns:us-east-1:111122223333:my-topic", *client.PublishInput[0].TopicArn)
}
