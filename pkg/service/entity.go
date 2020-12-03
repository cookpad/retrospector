package service

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/errors"
)

type EntityService struct {
	newS3 adaptor.S3ClientFactory
}

func NewEntityService(newS3 adaptor.S3ClientFactory) *EntityService {
	return &EntityService{
		newS3: newS3,
	}
}

type entityQueueMsg struct {
	Error  error
	Entity *retrospector.Entity
}

type ReadQueue struct {
	queue  chan *entityQueueMsg
	err    error
	closed bool
}

func (x *ReadQueue) Read() *retrospector.Entity {
	if x.closed {
		return nil
	}

	msg := <-x.queue
	if msg == nil {
		x.closed = true
		return nil
	}
	if msg.Error != nil {
		x.closed = true
		x.err = msg.Error
		return nil
	}

	return msg.Entity
}

func (x *ReadQueue) Error() error {
	return x.err
}

// NewReadQueue is constructor of ReadQueue
func (x *EntityService) NewReadQueue(region, bucket, key string) *ReadQueue {
	queue := make(chan *entityQueueMsg, 256)
	go func() {
		defer close(queue)
		s3Client, err := x.newS3(region)
		if err != nil {
			queue <- &entityQueueMsg{Error: err}
			return
		}

		input := &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}
		output, err := s3Client.GetObject(input)
		if err != nil {
			queue <- &entityQueueMsg{
				Error: errors.Wrap(err, "Failed GetObject").With("input", input),
			}
			return
		}

		scanner := bufio.NewScanner(output.Body)
		for scanner.Scan() {
			buf := scanner.Bytes()
			entity := &retrospector.Entity{}
			if err := json.Unmarshal(buf, entity); err != nil {
				queue <- &entityQueueMsg{
					Error: errors.Wrap(err, "Failed json.Marshal for scanned data").With("buf", string(buf)),
				}
				return
			}

			queue <- &entityQueueMsg{
				Entity: entity,
			}
		}
	}()

	return &ReadQueue{
		queue: queue,
	}
}

type WriteQueue struct {
	queue  chan *retrospector.Entity
	wg     sync.WaitGroup
	err    error
	closed bool
}

func (x *WriteQueue) Write(entity *retrospector.Entity) {
	if !x.closed {
		x.queue <- entity
	}
}

func (x *WriteQueue) Close() error {
	close(x.queue)
	x.wg.Wait()
	return x.err
}

// NewWriteQueue is constructor of WriteQueue
func (x *EntityService) NewWriteQueue(region, bucket, key string) *WriteQueue {
	queue := make(chan *retrospector.Entity, 256)
	wq := &WriteQueue{
		queue: queue,
	}

	wq.wg.Add(1)
	go func() {
		defer func() {
			wq.closed = true
			wq.wg.Done()
		}()

		s3Client, err := x.newS3(region)
		if err != nil {
			wq.err = errors.Wrap(err, "Failed to create S3Client").With("region", region)
			return
		}

		// TODO: async upload to S3
		buf := &bytes.Buffer{}
		gz := gzip.NewWriter(buf)
		rc := []byte("\n")
		for entity := range queue {
			raw, err := json.Marshal(entity)
			if err != nil {
				wq.err = errors.Wrap(err, "Failed to marshal entity").With("entity", entity)
				return
			}

			if _, err := gz.Write(append(raw, rc...)); err != nil {
				wq.err = errors.Wrap(err, "Failed to write line of entity").With("raw", string(raw))
				return
			}
		}
		if err := gz.Close(); err != nil {
			wq.err = errors.Wrap(err, "Failed to close gzip stream")
			return
		}

		input := &s3.PutObjectInput{
			Bucket:          aws.String(bucket),
			Key:             aws.String(key),
			Body:            bytes.NewReader(buf.Bytes()),
			ContentEncoding: aws.String("gzip"),
			ContentType:     aws.String("application/x-gzip"),
		}
		if _, err := s3Client.PutObject(input); err != nil {
			wq.err = errors.Wrap(err, "Failed to put object").With("input", input)
			return
		}
	}()

	return wq
}
