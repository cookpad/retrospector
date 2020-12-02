package mock

import (
	"fmt"
	"time"

	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
)

// Repository is mock of adaptor.Repository
type Repository struct {
	data map[string]map[string]interface{}
}

// NewRepository is constructor of mock.Repository
func NewRepository() adaptor.Repository {
	return &Repository{
		data: make(map[string]map[string]interface{}),
	}
}

// PutEntities puts entity set to memory
func (x *Repository) PutEntities(entities []*retrospector.Entity) error {
	for _, entity := range entities {
		ts := time.Unix(entity.RecordedAt, 0)
		pk := fmt.Sprintf("entity/%s/%s", entity.Type, entity.Data)
		sk := ts.Format("20060102_150405")

		smap, ok := x.data[pk]
		if !ok {
			smap = make(map[string]interface{})
			x.data[pk] = smap
		}
		smap[sk] = entity
	}

	return nil
}

// GetEntities fetches entity set from memory by IOC set
func (x *Repository) GetEntities(iocSet []*retrospector.IOC) ([]*retrospector.Entity, error) {
	var results []*retrospector.Entity
	for _, ioc := range iocSet {
		pk := fmt.Sprintf("entity/%s/%s", ioc.Type, ioc.Data)
		for _, v := range x.data[pk] {
			entity, ok := v.(*retrospector.Entity)
			if !ok {
				continue
			}
			results = append(results, entity)
		}
	}

	return results, nil
}

// PutIOCSet puts IOC set to memory
func (x *Repository) PutIOCSet(iocSet []*retrospector.IOC) error {
	for _, ioc := range iocSet {
		pk := fmt.Sprintf("ioc/%s/%s", ioc.Type, ioc.Data)
		sk := ioc.Source

		smap, ok := x.data[pk]
		if !ok {
			smap = make(map[string]interface{})
			x.data[pk] = smap
		}
		smap[sk] = ioc
	}

	return nil
}

// GetIOCSet fetches IOC set from memory by entity set
func (x *Repository) GetIOCSet(entities []*retrospector.Entity) ([]*retrospector.IOC, error) {
	var results []*retrospector.IOC
	for _, entity := range entities {
		pk := fmt.Sprintf("ioc/%s/%s", entity.Type, entity.Data)
		for _, v := range x.data[pk] {
			ioc, ok := v.(*retrospector.IOC)
			if !ok {
				continue
			}
			results = append(results, ioc)
		}
	}

	return results, nil
}
