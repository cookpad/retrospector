package service_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/mock"
	"github.com/m-mizutani/retrospector/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3ServiceWithMock(t *testing.T) {
	testS3Service(t, mock.NewS3Client, "my-region", "my-bucket")
}

func TestS3ServiceWithAWS(t *testing.T) {
	testBucket, ok := os.LookupEnv("TEST_BUCKET_NAME")
	if !ok {
		t.Skip(t, "TEST_BUCKET_NAME is not set")
	}

	testS3Service(t, adaptor.NewS3Client, os.Getenv("AWS_REGION"), testBucket)
}

func testS3Service(t *testing.T, newS3 adaptor.S3ClientFactory, s3Region, s3Bucket string) {
	t.Run("Simple write and read", func(t *testing.T) {
		now := time.Now().UTC()
		entities := []*retrospector.Entity{
			{
				Value: retrospector.Value{
					Data: "10.1.2.3",
					Type: retrospector.ValueIPAddr,
				},
				Source:     "hoge:1",
				RecordedAt: now.Unix(),
			},
			{
				Value: retrospector.Value{
					Data: "example.com",
					Type: retrospector.ValueDomainName,
				},
				Source:     "moge:1",
				RecordedAt: now.Add(time.Second).Unix(),
			},
			{
				Value: retrospector.Value{
					Data: "https://example.org",
					Type: retrospector.ValueURL,
				},
			},
		}

		s3Key := fmt.Sprintf("retrospector-test/%s.json.gz", uuid.New().String())
		svc := service.NewEntityService(newS3)
		wq := svc.NewWriteQueue(s3Region, s3Bucket, s3Key)
		for _, entity := range entities {
			wq.Write(entity)
		}

		require.NoError(t, wq.Close())
		rq := svc.NewReadQueue(s3Region, s3Bucket, s3Key)

		e0 := rq.Read()
		require.NotNil(t, e0)
		assert.Equal(t, entities[0], e0)

		e1 := rq.Read()
		require.NotNil(t, e1)
		assert.Equal(t, entities[1], e1)

		e2 := rq.Read()
		require.NotNil(t, e2)
		assert.Equal(t, entities[2], e2)

		e3 := rq.Read()
		require.Nil(t, e3)

		require.NoError(t, rq.Error())
	})
}
