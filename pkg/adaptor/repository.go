package adaptor

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/retrospector"

	"github.com/m-mizutani/golambda"
)

type Repository interface {
	PutEntities(entities []*retrospector.Entity) error
	GetEntities(iocSet []*retrospector.IOC) ([]*retrospector.Entity, error)
	UpdateEntityDetected(entity *retrospector.Entity) error
	PutIOCSet(iocSet []*retrospector.IOC) error
	GetIOCSet(entities []*retrospector.Entity) ([]*retrospector.IOC, error)
	UpdateIOCDetected(ioc *retrospector.IOC) error
}

type RepositoryFactory func(region, tableName string) (Repository, error)

func NewDynamoRepository(region, tableName string) (Repository, error) {
	ssn, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return &DynamoRepository{
		table: dynamo.New(ssn).Table(tableName),
	}, nil
}

type DynamoRepository struct {
	table dynamo.Table
}

const (
	dynamoHashKey    = "pk"
	dynamoRangeKey   = "sk"
	entityTimeToLive = time.Hour * 24 * 30
	iocTimeToLive    = time.Hour * 24 * 30
)

type dynamoItem struct {
	PK        string `dynamo:"pk"`
	SK        string `dynamo:"sk"`
	ExpiresAt int64  `dynamo:"expires_at"`
}

func (x *dynamoItem) HashKey() interface{}  { return x.PK }
func (x *dynamoItem) RangeKey() interface{} { return x.SK }

type entityItem struct {
	dynamoItem
	retrospector.Entity
}

type iocItem struct {
	dynamoItem
	retrospector.IOC
}

func makeEntityPKey(value *retrospector.Value) string {
	return fmt.Sprintf("entity/%s/%s", value.Type, value.Data)
}

func makeEntitySKey(entity *retrospector.Entity) string {
	sk := entity.Subject
	if sk == "" {
		sk = time.Unix(entity.RecordedAt, 0).Format("20060102_150405")
	}
	return sk
}

func (x *DynamoRepository) PutEntities(entities []*retrospector.Entity) error {
	var items []interface{}
	for _, entity := range entities {
		items = append(items, &entityItem{
			dynamoItem: dynamoItem{
				PK:        makeEntityPKey(&entity.Value),
				SK:        makeEntitySKey(entity),
				ExpiresAt: time.Unix(entity.RecordedAt, 0).Add(entityTimeToLive).Unix(),
			},
			Entity: *entity,
		})
	}

	if n, err := x.table.Batch().Write().Put(items...).Run(); err != nil {
		return golambda.WrapError(err, "PutEntities").With("items", items)
	} else if n != len(items) {
		return golambda.NewError("A number of wrote items is mismatched").With("n", n).With("items", items)
	}

	return nil
}

func (x *DynamoRepository) GetEntities(iocSet []*retrospector.IOC) ([]*retrospector.Entity, error) {
	var entities []*retrospector.Entity

	for _, ioc := range iocSet {
		pk := makeEntityPKey(&ioc.Value)
		var entityItems []*entityItem
		if err := x.table.Get(dynamoHashKey, pk).All(&entityItems); err != nil {
			return nil, golambda.WrapError(err, "Batch get entities").With("pk", pk).With("ioc", ioc)
		}

		for _, item := range entityItems {
			entities = append(entities, &item.Entity)
		}
	}

	return entities, nil
}

func (x *DynamoRepository) UpdateEntityDetected(entity *retrospector.Entity) error {
	pk := makeEntityPKey(&entity.Value)
	sk := makeEntitySKey(entity)

	q := x.table.Update(dynamoHashKey, pk).
		Range(dynamoRangeKey, sk).
		Set("detected", true)

	if err := q.Run(); err != nil {
		return golambda.WrapError(err, "Failed to update entity to detected").
			With("entity", entity).With("pk", pk).With("sk", sk)
	}

	return nil
}

func makeIOCPKey(value *retrospector.Value) string {
	return fmt.Sprintf("ioc/%s/%s", value.Type, value.Data)
}

func makeIOCSKey(ioc *retrospector.IOC) string {
	return ioc.Source
}

func (x *DynamoRepository) PutIOCSet(iocSet []*retrospector.IOC) error {
	var items []interface{}
	for _, ioc := range iocSet {
		ts := time.Unix(ioc.UpdatedAt, 0)
		items = append(items, &iocItem{
			dynamoItem: dynamoItem{
				PK:        makeIOCPKey(&ioc.Value),
				SK:        makeIOCSKey(ioc),
				ExpiresAt: ts.Add(iocTimeToLive).Unix(),
			},
			IOC: *ioc,
		})
	}

	if n, err := x.table.Batch().Write().Put(items...).Run(); err != nil {
		return golambda.WrapError(err, "PutIOCSet").With("items", items)
	} else if n != len(items) {
		return golambda.NewError("A number of wrote items is mismatched").With("n", n).With("items", items)
	}

	return nil
}

func (x *DynamoRepository) GetIOCSet(entities []*retrospector.Entity) ([]*retrospector.IOC, error) {

	var iocSet []*retrospector.IOC
	for _, entity := range entities {
		pk := makeIOCPKey(&entity.Value)

		var iocItems []*iocItem
		if err := x.table.Get(dynamoHashKey, pk).All(&iocItems); err != nil {
			return nil, golambda.WrapError(err, "Batch get entities").With("pk", pk).With("entity", entity)
		}

		for _, item := range iocItems {
			iocSet = append(iocSet, &item.IOC)
		}
	}

	return iocSet, nil
}

func (x *DynamoRepository) UpdateIOCDetected(ioc *retrospector.IOC) error {
	pk := makeIOCPKey(&ioc.Value)
	sk := makeIOCSKey(ioc)

	q := x.table.Update(dynamoHashKey, pk).
		Range(dynamoRangeKey, sk).
		Set("detected", true)

	if err := q.Run(); err != nil {
		return golambda.WrapError(err, "Failed to update IOC to detected").
			With("ioc", ioc).With("pk", pk).With("sk", sk)
	}

	return nil
}
