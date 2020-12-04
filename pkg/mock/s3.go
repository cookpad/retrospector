package mock

import (
	"compress/gzip"
	"errors"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
)

var s3Objects map[string]map[string]io.Reader

func init() {
	s3Objects = make(map[string]map[string]io.Reader)
}

type S3Client struct {
	Region string
}

func NewS3Mock() (adaptor.S3ClientFactory, *S3Client) {
	client := &S3Client{}
	return func(region string) (adaptor.S3Client, error) {
		client.Region = region
		return client, nil
	}, client
}

func NewS3Client(region string) (adaptor.S3Client, error) {
	return &S3Client{}, nil
}

func (x *S3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	bucket, ok := s3Objects[*input.Bucket]
	if !ok {
		return nil, errors.New(s3.ErrCodeNoSuchBucket)
	}

	reader, ok := bucket[*input.Key]
	if !ok {
		return nil, errors.New(s3.ErrCodeNoSuchKey)
	}

	gz, err := gzip.NewReader(reader)
	if err != nil {
		log.Fatal("gzip error in GetObject: ", err)
	}

	return &s3.GetObjectOutput{
		Body: gz,
	}, nil
}

func (x *S3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	memBucket, ok := s3Objects[*input.Bucket]
	if !ok {
		memBucket = make(map[string]io.Reader)
		s3Objects[*input.Bucket] = memBucket
	}

	memBucket[*input.Key] = input.Body
	return &s3.PutObjectOutput{}, nil
}
