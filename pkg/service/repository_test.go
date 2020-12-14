package service_test

import (
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

func TestDynamoRepositoryService(t *testing.T) {
	tableName, ok := os.LookupEnv("TEST_TABLE_NAME")
	if !ok {
		t.Skip("Skip test because TEST_TABLE_NAME is not set")
	}
	region, ok := os.LookupEnv("AWS_REGION")
	if !ok {
		t.Skip("Skip test because AWS_REGION is not set")
	}

	repo, err := adaptor.NewDynamoRepository(region, tableName)
	require.NoError(t, err)
	svc := service.NewRepositoryService(repo)
	testRepositoryService(t, svc)
}

func TestMockRepositoryService(t *testing.T) {
	svc := service.NewRepositoryService(mock.NewRepository())
	testRepositoryService(t, svc)
}

func testRepositoryService(t *testing.T, svc *service.RepositoryService) {
	t.Run("EntityTest", func(t *testing.T) {
		now := time.Now()
		v1 := uuid.New().String()
		v2 := uuid.New().String()

		data := []*retrospector.Entity{
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueDomainName,
				},
				Source:     "blue",
				RecordedAt: now.Unix(),
			},
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueIPAddr,
				},
				Source:     "blue_ipaddr",
				RecordedAt: now.Add(time.Second).Unix(),
			},
			{
				Value: retrospector.Value{
					Data: v2,
					Type: retrospector.ValueDomainName,
				},
				Source:     "orange1",
				RecordedAt: now.Unix(),
			},
			{
				Value: retrospector.Value{
					Data: v2,
					Type: retrospector.ValueDomainName,
				},
				Source:     "orange2",
				RecordedAt: now.Add(time.Second).Unix(),
			},
		}

		err := svc.PutEntities(data)
		require.NoError(t, err)

		t.Run("found one entity by one ioc", func(t *testing.T) {
			resp, err := svc.GetEntities([]*retrospector.IOC{
				{
					Value: retrospector.Value{
						Data: v1,
						Type: retrospector.ValueDomainName,
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, 1, len(resp))
			assert.Equal(t, data[0], resp[0])
		})

		t.Run("found 2 entities by one ioc", func(t *testing.T) {
			resp, err := svc.GetEntities([]*retrospector.IOC{
				{
					Value: retrospector.Value{
						Data: v2,
						Type: retrospector.ValueDomainName,
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, 2, len(resp))
			assert.Contains(t, resp, data[2])
			assert.Contains(t, resp, data[3])
		})

		t.Run("found different entity by different value type", func(t *testing.T) {
			resp, err := svc.GetEntities([]*retrospector.IOC{
				{
					Value: retrospector.Value{
						Data: v1,
						Type: retrospector.ValueIPAddr,
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, 1, len(resp))
			assert.Contains(t, resp, data[1])
		})
	})

	t.Run("IOCTest", func(t *testing.T) {
		now := time.Now()
		v1 := uuid.New().String()
		v2 := uuid.New().String()

		data := []*retrospector.IOC{
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueFileHashSha256,
				},

				Source:    "blue",
				UpdatedAt: now.Unix(),
			},
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueFileHashSha256,
				},

				Source:    "orange",
				UpdatedAt: now.Unix(),
			},
			{
				Value: retrospector.Value{
					Data: v2,
					Type: retrospector.ValueDomainName,
				},
				Source:    "blue",
				UpdatedAt: now.Add(time.Second).Unix(),
			},
		}
		require.NoError(t, svc.PutIOCSet(data))

		t.Run("found 1 ioc by 1 entity", func(t *testing.T) {
			resp, err := svc.GetIOCSet([]*retrospector.Entity{
				{
					Value: retrospector.Value{
						Data: v2,
						Type: retrospector.ValueDomainName,
					},
				},
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(resp))
			assert.Equal(t, data[2], resp[0])
		})

		t.Run("found 2 ioc by 1 entity", func(t *testing.T) {
			resp, err := svc.GetIOCSet([]*retrospector.Entity{
				{
					Value: retrospector.Value{
						Data: v1,
						Type: retrospector.ValueFileHashSha256,
					},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, 2, len(resp))
			assert.Contains(t, resp, data[0])
			assert.Contains(t, resp, data[1])
		})
	})

	t.Run("update detection status of entity", func(t *testing.T) {
		v1 := uuid.New().String()

		data := []*retrospector.Entity{
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueDomainName,
				},
				Subject: "tester",
			},
		}
		require.NoError(t, svc.PutEntities(data))

		entities := []*retrospector.IOC{
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueDomainName,
				},
			},
		}

		resp, err := svc.GetEntities(entities)
		require.NoError(t, err)
		require.Equal(t, 1, len(resp))
		assert.False(t, resp[0].Detected)

		require.NoError(t, svc.UpdateEntityDetected(data[0]))

		resp, err = svc.GetEntities(entities)
		require.NoError(t, err)
		require.Equal(t, 1, len(resp))
		assert.True(t, resp[0].Detected)
	})

	t.Run("update detection status of IOC", func(t *testing.T) {
		now := time.Now()
		v1 := uuid.New().String()

		data := []*retrospector.IOC{
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueDomainName,
				},

				Source:    "blue",
				UpdatedAt: now.Unix(),
			},
		}
		require.NoError(t, svc.PutIOCSet(data))

		entities := []*retrospector.Entity{
			{
				Value: retrospector.Value{
					Data: v1,
					Type: retrospector.ValueDomainName,
				},
			},
		}

		resp, err := svc.GetIOCSet(entities)
		require.NoError(t, err)
		require.Equal(t, 1, len(resp))
		assert.False(t, resp[0].Detected)

		require.NoError(t, svc.UpdateIOCDetected(data[0]))

		resp, err = svc.GetIOCSet(entities)
		require.NoError(t, err)
		require.Equal(t, 1, len(resp))
		assert.True(t, resp[0].Detected)
	})
}
