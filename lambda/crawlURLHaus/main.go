package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/errors"
	"github.com/m-mizutani/retrospector/pkg/lambda"
)

const (
	urlhausURL     = "https://urlhaus.abuse.ch/downloads/csv_recent/"
	chunkSizeLimit = 32
)

func isIPaddress(v string) bool {
	return net.ParseIP(v) != nil
}

// Handler is main function and exposed for test
func Handler(args *lambda.Arguments) error {
	snsSvc := args.SNSService()

	req, err := http.NewRequest("GET", urlhausURL, nil)
	if err != nil {
		return errors.Wrap(err, "Fail to create new URLhaus HTTP request").With("url", urlhausURL)
	}

	client := args.HTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Fail to send HTTP request").With("url", urlhausURL)
	}
	if resp.StatusCode != 200 {
		return errors.Wrap(err, "Unexpected status code").With("code", resp.StatusCode).With("url", urlhausURL)
	}

	reader := csv.NewReader(resp.Body)
	reader.Comment = []rune("#")[0]

	iocMap := make(map[retrospector.Value]*retrospector.IOC)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "Fail to read CSV of URLhaus")
		}

		if len(row) != 8 {
			continue
		}

		url, err := url.Parse(row[2])
		if err != nil {
			return errors.Wrap(err, "Fail to parse URL in URLhaus CSV")
		}

		ts, err := time.Parse("2006-01-02 15:04:05", row[1])
		if err != nil {
			return errors.Wrap(err, "Fail to parse tiemstamp in URLhaus CSV")
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

		if len(iocChunk) >= chunkSizeLimit {
			if err := snsSvc.Publish(args.IOCTopicARN, iocChunk); err != nil {
				return errors.With(err, "ioc", iocChunk).With("topic", args.IOCTopicARN)
			}
			iocChunk = nil
		}
	}

	if len(iocChunk) > 0 {
		if err := snsSvc.Publish(args.IOCTopicARN, iocChunk); err != nil {
			return errors.With(err, "ioc", iocChunk).With("topic", args.IOCTopicARN)
		}
	}

	return nil
}

func main() {
	lambda.Run(Handler)
}
