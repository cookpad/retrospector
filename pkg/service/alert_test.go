package service_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/service"
	"github.com/stretchr/testify/require"
)

func TestAlertServiceSlackIntegration(t *testing.T) {
	url, ok := os.LookupEnv("TEST_SLACK_WEBHOOK_URL")
	if !ok {
		t.Skip("TEST_SLACK_WEBHOOK_URL is not set")
	}

	alertSvc := service.NewAlertService(&service.AlertServiceArguments{
		SlackIncomingWebhookURL: url,
		HTTPClient:              &http.Client{},
	})

	err := alertSvc.EmitToSlack(&service.Alert{
		Cause: service.AlertCauseIOC,
		Target: &retrospector.Value{
			Data: "example.com",
			Type: retrospector.ValueDomainName,
		},
		Entities: []*retrospector.Entity{
			{
				Source:      "google drive",
				RecordedAt:  time.Now().Unix(),
				Description: "open hogehoge file",
			},
		},
		IOCChunk: retrospector.IOCChunk{
			{
				Source:      "URLHaus",
				UpdatedAt:   time.Now().Add(time.Hour * -10).Unix(),
				Reason:      "bad smell",
				Description: "something wrong",
			},
		},
	})

	require.NoError(t, err)
}
