package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
)

var logger = golambda.Logger

type otxContent struct {
	Content     string `json:"content"`
	Description string `json:"description"`
	ID          int64  `json:"id"`
	Indicator   string `json:"indicator"`
	Title       string `json:"title"`
	Type        string `json:"type"`
}

type otxResponse struct {
	Count    int64         `json:"count"`
	Next     *string       `json:"next"`
	Previous *string       `json:"previous"`
	Results  []*otxContent `json:"results"`
}

// Handler is exported for test
func Handler(args *arguments.Arguments, event golambda.Event) (interface{}, error) {
	baseURL := "https://otx.alienvault.com/api/v1/indicators/export"

	secrets, err := args.GetSecrets()
	if err != nil {
		return nil, err
	}

	iocMap := make(map[retrospector.Value]*retrospector.IOC)

	apiURL := &baseURL
	duration := time.Hour * 24
	now := time.Now().UTC()

	for apiURL != nil {
		req, err := http.NewRequest("GET", *apiURL, nil)
		if err != nil {
			return nil, golambda.WrapError(err, "Creating OTX API request").With("url", apiURL)
		}
		if apiURL == &baseURL {
			since := now.Add(-duration)
			q := url.Values{}
			q.Add("modified_since", since.Format("2006-01-02T15:04:05+00:00"))
			req.URL.RawQuery = q.Encode()
		}

		if secrets.OTXToken == "" {
			return nil, golambda.NewError("otx_token is not set in secrets").With("secretARN", args.SecretsARN)
		}

		req.Header.Add("X-OTX-API-KEY", secrets.OTXToken)

		logger.With("url", req.URL.String()).Trace("API access")
		resp, err := args.HTTPClient().Do(req)
		if err != nil {
			return nil, golambda.WrapError(err).With("url", req.URL.String())
		}
		if resp.StatusCode != http.StatusOK {
			raw, err := ioutil.ReadAll(resp.Body)
			body := string(raw)
			if err != nil {
				body = err.Error()
			}
			return nil, golambda.NewError("OTX server error").With("body", body).With("code", resp.StatusCode).With("URL", req.URL.String())
		}

		var otxResp otxResponse
		if err := json.NewDecoder(resp.Body).Decode(&otxResp); err != io.EOF && err != nil {
			return nil, golambda.WrapError(err, "Decoding JSON response").With("URL", req.URL.String())
		}

		for _, content := range otxResp.Results {
			var value *retrospector.Value
			switch content.Type {
			case "hostname", "domain":
				value = &retrospector.Value{
					Data: content.Indicator,
					Type: retrospector.ValueDomainName,
				}
			case "IPv4":
				value = &retrospector.Value{
					Data: content.Indicator,
					Type: retrospector.ValueIPAddr,
				}
			}

			if value == nil {
				continue
			}

			if ioc, ok := iocMap[*value]; !ok {
				ioc = &retrospector.IOC{
					Value:       *value,
					Source:      "otx",
					Reason:      content.Title,
					UpdatedAt:   now.Unix(),
					Description: fmt.Sprintf("id:%d", content.ID),
				}
				iocMap[*value] = ioc
			} else if len(ioc.Description) < 1024 {
				ioc.Description += fmt.Sprintf(", id:%d", content.ID)
			}
		}

		apiURL = otxResp.Next
	}

	var iocChunk retrospector.IOCChunk
	for _, ioc := range iocMap {
		iocChunk = append(iocChunk, ioc)
	}

	snsSvc := args.SNSService()
	if err := snsSvc.PublishIOC(args.IOCTopicARN, iocChunk); err != nil {
		return nil, golambda.WrapError(err).With("topic", args.IOCTopicARN)
	}

	logger.With("ioc_count", len(iocChunk)).Info("Published IOC")

	return nil, nil
}

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		return Handler(arguments.New(), event)
	})
}
