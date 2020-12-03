package main

import (
	"encoding/csv"
	"fmt"
	"io"
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

	var iocChunk retrospector.IOCChunk
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

		iocChunk = append(iocChunk, &retrospector.IOC{
			Value: retrospector.Value{
				Data: url.Hostname(),
				Type: retrospector.ValueDomainName,
			},
			Source:      "URLhaus",
			UpdatedAt:   ts.Unix(),
			Reason:      row[4],
			Description: fmt.Sprintf("%s: %s", row[0], row[2]),
		})

		if len(iocChunk) >= chunkSizeLimit {
			if err := snsSvc.Publish(args.IOCTopicARN, iocChunk); err != nil {
				return errors.With(err, "ioc", iocChunk).With("topic", args.IOCTopicARN)
			}
			iocChunk = retrospector.IOCChunk{}
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

/*

func downloadURLhasu(csvURL string, ch chan *badman.EntityQueue) {
	defer close(ch)
	bufferSize := 128
	buffer := []*badman.BadEntity{}

	defer func() {
		if len(buffer) > 0 {
			ch <- &badman.EntityQueue{Entities: buffer}
		}
	}()

	body := getHTTPBody(csvURL, ch)
	if body == nil {
		return
	}

	reader := csv.NewReader(body)
	reader.Comment = []rune("#")[0]

	for {
		row, err := reader.Read()
		if err == io.EOF {
			return
		} else if err != nil {
			ch <- &badman.EntityQueue{
				Error: errors.Wrapf(err, "Fail to read CSV of URLhaus"),
			}
			return
		}

		if len(row) != 8 {
			continue
		}

		url, err := url.Parse(row[2])
		if err != nil {
			ch <- &badman.EntityQueue{
				Error: errors.Wrapf(err, "Fail to parse URL in URLhaus CSV"),
			}
			return
		}

		ts, err := time.Parse("2006-01-02 15:04:05", row[1])
		if err != nil {
			ch <- &badman.EntityQueue{
				Error: errors.Wrapf(err, "Fail to parse tiemstamp in URLhaus CSV"),
			}
			return
		}

		buffer = append(buffer, &badman.BadEntity{
			Name:    url.Hostname(),
			SavedAt: ts,
			Src:     "URLhaus",
			Reason:  row[4],
		})

		if len(buffer) >= bufferSize {
			ch <- &badman.EntityQueue{Entities: buffer}
			buffer = []*badman.BadEntity{}
		}
	}
}

// URLhausRecent downloads blacklist from https://urlhaus.abuse.ch/downloads/csv_recent/
// The blacklist has only URLs in recent 30 days.
type URLhausRecent struct {
	URL string
}

// NewURLhausRecent is constructor of URLhausRecent
func NewURLhausRecent() *URLhausRecent {
	return &URLhausRecent{
		URL: "https://urlhaus.abuse.ch/downloads/csv_recent/",
	}
}

// Download of URLhausRecent downloads domains.txt and parses to extract domain names.
func (x *URLhausRecent) Download() chan *badman.EntityQueue {
	ch := make(chan *badman.EntityQueue, defaultSourceChanSize)
	go downloadURLhasu(x.URL, ch)
	return ch
}

// URLhausOnline downloads blacklist from https://urlhaus.abuse.ch/downloads/csv_recent/
// The blacklist has only online URLs.
type URLhausOnline struct {
	URL string
}

// NewURLhausOnline is constructor of URLhausOnline
func NewURLhausOnline() *URLhausOnline {
	return &URLhausOnline{
		URL: "https://urlhaus.abuse.ch/downloads/csv_online/",
	}
}

// Download of URLhausOnline downloads domains.txt and parses to extract domain names.
func (x *URLhausOnline) Download() chan *badman.EntityQueue {
	ch := make(chan *badman.EntityQueue, defaultSourceChanSize)
	go downloadURLhasu(x.URL, ch)
	return ch


*/
