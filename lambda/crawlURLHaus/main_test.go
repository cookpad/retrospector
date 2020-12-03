package main_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/lambda"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/m-mizutani/retrospector/lambda/crawlURLHaus"
)

func TestCrawlURLHausIntegration(t *testing.T) {
	if _, ok := os.LookupEnv("ENABLE_INTEGRATION_TEST"); !ok {
		t.Skip("ENABLE_INTEGRATION_TEST is not set")
	}

	newSNS, client := mock.NewSNSMock()
	args := &lambda.Arguments{
		IOCTopicARN: "arn:aws:sns:us-east-1:111122223333:my-topic",
		NewSNS:      newSNS,
	}

	require.NoError(t, main.Handler(args))
	assert.Greater(t, len(client.PublishInput), 1000)
	assert.Equal(t, "us-east-1", client.Region)
	assert.Equal(t, "arn:aws:sns:us-east-1:111122223333:my-topic", *client.PublishInput[0].TopicArn)
}

func TestCrawlURLHaus(t *testing.T) {
	sampleData := `
################################################################
# abuse.ch URLhaus Database Dump (CSV - recent URLs only)      #
# Last updated: 2020-12-03 06:06:09 (UTC)                      #
#                                                              #
# Terms Of Use: https://urlhaus.abuse.ch/api/                  #
# For questions please contact urlhaus [at] abuse.ch           #
################################################################
#
# id,dateadded,url,url_status,threat,tags,urlhaus_link,reporter
"884896","2020-12-03 06:06:09","http://94.122.77.235:32794/Mozi.m","online","malware_download","elf,Mozi","https://urlhaus.abuse.ch/url/884896/","lrz_urlhaus"
"884895","2020-12-03 06:06:08","http://61.52.236.225:34611/Mozi.m","online","malware_download","elf,Mozi","https://urlhaus.abuse.ch/url/884895/","lrz_urlhaus"
"884894","2020-12-03 06:06:06","http://182.121.210.95:37153/bin.sh","online","malware_download","32-bit,elf,mips","https://urlhaus.abuse.ch/url/884894/","geenensp"
"884893","2020-12-03 06:06:06","http://83.224.148.25:53456/Mozi.a","online","malware_download","elf,Mozi","https://urlhaus.abuse.ch/url/884893/","lrz_urlhaus"
`
	ts1, err := time.Parse("2006-01-02 15:04:05", "2020-12-03 06:06:09")
	require.NoError(t, err)

	newSNS, client := mock.NewSNSMock()
	httpClient := &mock.HTTPClient{
		RespCode: http.StatusOK,
		RespBody: ioutil.NopCloser(strings.NewReader(sampleData)),
	}

	args := &lambda.Arguments{
		IOCTopicARN: "arn:aws:sns:us-east-1:111122223333:my-topic",
		NewSNS:      newSNS,
		HTTP:        httpClient,
	}

	require.NoError(t, main.Handler(args))
	require.Equal(t, 1, len(client.PublishInput))
	assert.Equal(t, "us-east-1", client.Region)
	assert.Equal(t, "arn:aws:sns:us-east-1:111122223333:my-topic", *client.PublishInput[0].TopicArn)

	var iocChunk retrospector.IOCChunk
	require.NoError(t, json.Unmarshal([]byte(*client.PublishInput[0].Message), &iocChunk))
	require.Equal(t, 4, len(iocChunk))
	assert.Equal(t, "94.122.77.235", iocChunk[0].Data)
	assert.Equal(t, ts1.Unix(), iocChunk[0].UpdatedAt)
}
