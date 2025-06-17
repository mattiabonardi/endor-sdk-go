package sdk

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type AbstractResourceRepository struct {
	resourceDefinition ResourceDefinition
	repositories       map[string]any // es. {"mongodb": *MongoAbstractResourceRepository}
}

func NewAbstractResourceRepository(def ResourceDefinition, client *mongo.Client, dbName string) (*AbstractResourceRepository, error) {
	repos := make(map[string]any)

	for _, ds := range def.DataSources {
		switch ds.GetType() {
		case "mongodb":
			mongoRepo, err := NewMongoAbstractResourceRepository(client, dbName, def)
			if err != nil {
				return nil, err
			}
			repos["mongodb"] = mongoRepo
		default:
			return nil, fmt.Errorf("unsupported data source type: %s", ds.GetType())
		}
	}

	return &AbstractResourceRepository{
		resourceDefinition: def,
		repositories:       repos,
	}, nil
}

func (r *AbstractResourceRepository) Create(dto CreateDTO[any]) error {
	for _, ds := range r.resourceDefinition.DataSources {
		switch ds.GetType() {
		case "mongodb":
			mongoRepo := r.repositories["mongodb"].(*MongoAbstractResourceRepository)
			if err := mongoRepo.Create(dto.Data); err != nil {
				return fmt.Errorf("mongo create failed: %w", err)
			}
		// in futuro: case "postgres": ...
		default:
			continue
		}
	}
	return nil
}
