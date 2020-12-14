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

func makeEntityPKey(value *retrospector.Value) string {
	return fmt.Sprintf("entity/%s/%s", value.Type, value.Data)
}

func makeIOCPKey(value *retrospector.Value) string {
	return fmt.Sprintf("ioc/%s/%s", value.Type, value.Data)
}

func makeIOCSKey(ioc *retrospector.IOC) string {
	return ioc.Source
}

func makeEntitySKey(entity *retrospector.Entity) string {
	sk := entity.Subject
	if sk == "" {
		sk = time.Unix(entity.RecordedAt, 0).Format("20060102_150405")
	}
	return sk
}

// PutEntities puts entity set to memory
func (x *Repository) PutEntities(entities []*retrospector.Entity) error {
	for _, entity := range entities {
		pk := makeEntityPKey(&entity.Value)
		sk := makeEntitySKey(entity)

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
		pk := makeEntityPKey(&ioc.Value)

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

func (x *Repository) UpdateEntityDetected(target *retrospector.Entity) error {
	pk := makeEntityPKey(&target.Value)
	sk := makeEntitySKey(target)

	if p, ok := x.data[pk]; ok {
		if s, ok := p[sk]; ok {
			if entity, ok := s.(*retrospector.Entity); ok {
				entity.Detected = true
			}
		}
	}

	return nil
}

// PutIOCSet puts IOC set to memory
func (x *Repository) PutIOCSet(iocSet []*retrospector.IOC) error {
	for _, ioc := range iocSet {
		pk := makeIOCPKey(&ioc.Value)
		sk := makeIOCSKey(ioc)

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
		pk := makeIOCPKey(&entity.Value)

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

func (x *Repository) UpdateIOCDetected(target *retrospector.IOC) error {
	pk := makeIOCPKey(&target.Value)
	sk := makeIOCSKey(target)

	if p, ok := x.data[pk]; ok {
		if s, ok := p[sk]; ok {
			if ioc, ok := s.(*retrospector.IOC); ok {
				ioc.Detected = true
			}
		}
	}

	return nil
}
