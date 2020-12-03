package service

import (
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
	"github.com/m-mizutani/retrospector/pkg/errors"
)

type RepositoryService struct {
	repo adaptor.Repository
}

func NewRepositoryService(repo adaptor.Repository) *RepositoryService {
	return &RepositoryService{
		repo: repo,
	}
}

func (x *RepositoryService) PutEntities(entities []*retrospector.Entity) error {
	return x.repo.PutEntities(entities)
}

func (x *RepositoryService) GetEntities(iocSet []*retrospector.IOC) ([]*retrospector.Entity, error) {
	return x.repo.GetEntities(iocSet)
}

func (x *RepositoryService) PutIOCSet(iocSet []*retrospector.IOC) error {
	step := 10
	for i := 0; i < len(iocSet); i += step {
		ep := i + step
		if len(iocSet) < ep {
			ep = len(iocSet)
		}
		target := iocSet[i:ep]
		if err := x.repo.PutIOCSet(target); err != nil {
			return errors.With(err, "i", i)
		}
	}
	return nil
}

func (x *RepositoryService) GetIOCSet(entities []*retrospector.Entity) ([]*retrospector.IOC, error) {
	return x.repo.GetIOCSet(entities)
}
