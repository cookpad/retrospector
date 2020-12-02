package service

import (
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/adaptor"
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
	return x.repo.PutIOCSet(iocSet)
}

func (x *RepositoryService) GetIOCSet(entities []*retrospector.Entity) ([]*retrospector.IOC, error) {
	return x.repo.GetIOCSet(entities)
}
