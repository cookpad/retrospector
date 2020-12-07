package mock_test

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockS3(t *testing.T) {
	s3client, err := mock.NewS3Client("eu-east-0")
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	wr := gzip.NewWriter(buf)
	wr.Write([]byte("five"))
	require.NoError(t, wr.Close())

	putInput := &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("blue"),
		Body:   bytes.NewReader(buf.Bytes()),
	}

	_, err = s3client.PutObject(putInput)
	require.NoError(t, err)

	t.Run("Get exiting object", func(t *testing.T) {
		input := &s3.GetObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("blue"),
		}
		output, err := s3client.GetObject(input)
		require.NoError(t, err)
		data, err := ioutil.ReadAll(output.Body)
		require.NoError(t, err)
		assert.Equal(t, "five", string(data))
	})

	t.Run("Access non-existing object and get error", func(t *testing.T) {
		input := &s3.GetObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("orange"),
		}
		_, err := s3client.GetObject(input)
		assert.Error(t, err)
		assert.Equal(t, s3.ErrCodeNoSuchKey, err.Error())
	})
}
