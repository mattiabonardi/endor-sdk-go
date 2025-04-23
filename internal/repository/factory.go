package repository

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/models"
)

type RepositoryFactory struct {
	adapters models.RepositoryAdapters
}

func (h *RepositoryFactory) Create(resource models.Resource) (models.Repository[any], error) {
	if resource.Persistence.Type == "mongodb" {
		r := NewMongoRepository[any](h.adapters.MongoClient, resource)
		return r, nil
	} else {
		return nil, fmt.Errorf("persistence class %s not existent", resource.Persistence.Type)
	}
}
