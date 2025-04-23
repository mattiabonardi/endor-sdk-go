package repository

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal"
)

type RepositoryFactory struct {
	adapters internal.RepositoryAdapters
}

func (h *RepositoryFactory) Create(resource internal.Resource) (internal.Repository[any], error) {
	if resource.Persistence.Type == "mongodb" {
		r := NewMongoResourceRepository[any](h.adapters.MongoClient, resource)
		return r, nil
	} else {
		return nil, fmt.Errorf("persistence class %s not existent", resource.Persistence.Type)
	}
}
