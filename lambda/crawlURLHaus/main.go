package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
)

const (
	urlhausURL     = "https://urlhaus.abuse.ch/downloads/csv_recent/"
	chunkSizeLimit = 32
)

func isIPaddress(v string) bool {
	return net.ParseIP(v) != nil
}

// Handler is main function and exposed for test
func Handler(args *arguments.Arguments, event golambda.Event) (interface{}, error) {
	snsSvc := args.SNSService()

	req, err := http.NewRequest("GET", urlhausURL, nil)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to create new URLhaus HTTP request").With("url", urlhausURL)
	}

	client := args.HTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to send HTTP request").With("url", urlhausURL)
	}
	if resp.StatusCode != 200 {
		return nil, golambda.WrapError(err, "Unexpected status code").With("code", resp.StatusCode).With("url", urlhausURL)
	}

	reader := csv.NewReader(resp.Body)
	reader.Comment = []rune("#")[0]

	iocMap := make(map[retrospector.Value]*retrospector.IOC)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, golambda.WrapError(err, "Fail to read CSV of URLhaus")
		}

		if len(row) != 8 {
			continue
		}

		url, err := url.Parse(row[2])
		if err != nil {
			return nil, golambda.WrapError(err, "Fail to parse URL in URLhaus CSV")
		}

		ts, err := time.Parse("2006-01-02 15:04:05", row[1])
		if err != nil {
			return nil, golambda.WrapError(err, "Fail to parse tiemstamp in URLhaus CSV")
		}

		value := retrospector.Value{
			Data: url.Hostname(),
			Type: retrospector.ValueDomainName,
		}
		if isIPaddress(value.Data) {
			value.Type = retrospector.ValueIPAddr
		}

		ioc, ok := iocMap[value]
		if !ok {
			ioc = &retrospector.IOC{
				Value:       value,
				Source:      "URLhaus",
				UpdatedAt:   ts.Unix(),
				Reason:      row[4],
				Description: fmt.Sprintf("%s: %s", row[0], row[2]),
			}
			iocMap[value] = ioc
		} else if len(ioc.Description) < 1024 {
			ioc.Description += fmt.Sprintf(", %s: %s", row[0], row[2])
		}
	}

	var iocChunk retrospector.IOCChunk
	for _, ioc := range iocMap {
		iocChunk = append(iocChunk, ioc)
	}
	if err := snsSvc.PublishIOC(args.IOCTopicARN, iocChunk); err != nil {
		return nil, golambda.WrapError(err).With("topic", args.IOCTopicARN)
	}

	return nil, nil
}

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		return Handler(arguments.New(), event)
	})
}
