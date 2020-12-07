package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/errors"

	"github.com/slack-go/slack"
)

type AlertServiceArguments struct {
	SlackIncomingWebhookURL string
	HTTPClient              adaptor.HTTPClient
}

type AlertService struct {
	args *AlertServiceArguments
}

func NewAlertService(args *AlertServiceArguments) *AlertService {
	return &AlertService{
		args: args,
	}
}

// Alert contains entity set and IOC set of detection
type Alert struct {
	Cause    AlertCause
	Target   *retrospector.Value
	Entities []*retrospector.Entity
	IOCChunk retrospector.IOCChunk
}

// AlertCause shows type of alert
type AlertCause int

const (
	// AlertCauseEntity indicates that a new entity is in existing IOC data set
	AlertCauseEntity AlertCause = iota
	// AlertCauseIOC indicates that a new IOC is in existing entity data set
	AlertCauseIOC
)

// Up to 3 IOC/entity items in slack message
const maxItemDisplaySlack = 3

func (x *AlertService) EmitToSlack(alert *Alert) error {
	if x.args.HTTPClient == nil {
		return errors.New("HTTPClient is required in AlertServiceArguments to emit Slack, but not set")
	}
	if x.args.SlackIncomingWebhookURL == "" {
		return errors.New("SlackIncomingWebhookURL is required in AlertServiceArguments to emit Slack, but not set")
	}

	newField := func(title, value string) *slack.TextBlockObject {
		return slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*\n%s", title, value), false, false)
	}

	title := fmt.Sprintf(":alert: New Alert: %s (%s)",
		strings.Replace(alert.Target.Data, ".", "[.]", -1),
		alert.Target.Type,
	)
	blocks := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", title, true, false)),
	}

	blocks = append(blocks, slack.NewDividerBlock())
	blocks = append(blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", "*Detected IOC*", false, false), nil, nil))

	for i, ioc := range alert.IOCChunk {
		if i >= 3 {
			break
		}

		blocks = append(blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*", ioc.Source), false, false),
			[]*slack.TextBlockObject{
				newField("Reason", ioc.Reason),
				newField("UpdatedAt", time.Unix(ioc.UpdatedAt, 0).Format("2006-01-02 15:04:05")),
				newField("Description", strings.Replace(ioc.Description, ".", "[.]", -1)),
			}, nil),
		)
	}

	blocks = append(blocks, slack.NewDividerBlock())
	blocks = append(blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", "*Affected Entity*", false, false), nil, nil))

	for i, entity := range alert.Entities {
		if i >= 3 {
			break
		}

		blocks = append(blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*", entity.Source), false, false),
			[]*slack.TextBlockObject{
				newField("Description", entity.Description),
				newField("RecordedAt", time.Unix(entity.RecordedAt, 0).Format("2006-01-02 15:04:05")),
			}, nil,
		))
	}

	msg := slack.NewBlockMessage(blocks...)
	raw, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal slack message").With("msg", msg)
	}

	req, err := http.NewRequest("POST", x.args.SlackIncomingWebhookURL, bytes.NewBuffer(raw))
	if err != nil {
		return errors.Wrap(err, "Failed to create a new HTTP request to Slack")
	}

	resp, err := x.args.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Failed to post message to slack in communication").With("msg", msg)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New("Failed to post message to slack in API").
			With("msg", msg).With("code", resp.StatusCode).With("body", string(body))
	}

	return nil
}
