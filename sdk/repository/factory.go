package repository

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
	"github.com/mattiabonardi/endor-sdk-go/sdk/resource"
)

type RepositoryFactory struct {
	adapters sdk.RepositoryAdapters
}

func (h *RepositoryFactory) Create(resource resource.Resource) (sdk.Repository[any], error) {
	if resource.Persistence.Type == "mongodb" {
		r := NewMongoResourceRepository[any](h.adapters.MongoClient, resource)
		return r, nil
	} else {
		return nil, fmt.Errorf("persistence class %s not existent", resource.Persistence.Type)
	}
}
