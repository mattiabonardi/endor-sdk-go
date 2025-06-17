package sdk

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAbstractResourceRepository struct {
	Collection     *mongo.Collection
	ResourceSchema ResourceDefinition
}

func NewMongoAbstractResourceRepository(client *mongo.Client, dbName string, def ResourceDefinition) (*MongoAbstractResourceRepository, error) {
	// Trova il primo datasource Mongo (per ora ne supportiamo solo uno)
	var mongoDS *MongoDataSource
	for _, ds := range def.DataSources {
		if ds.GetType() == "mongodb" {
			mongoDS = ds.(*MongoDataSource)
			break
		}
	}
	if mongoDS == nil {
		return nil, fmt.Errorf("no MongoDB data source found in definition")
	}

	coll := client.Database(dbName).Collection(mongoDS.Collection)
	return &MongoAbstractResourceRepository{
		Collection:     coll,
		ResourceSchema: def,
	}, nil
}

func (r *MongoAbstractResourceRepository) Create(resource any) error {
	// TODO: create real document based on path name, not property name
	// for now skip. Now document has the name fields as schema properties
	//doc := make(map[string]any)
	//for logicalName, fieldMap := range r.getMappings() {
	//}

	//TODO: handle ID field

	_, err := r.Collection.InsertOne(context.Background(), resource)
	return err
}

// Recupera i mappings Mongo dal ResourceDefinition
/*func (r *MongoAbstractResourceRepository) getMappings() map[string]MongoFieldMapping {
	for _, ds := range r.ResourceSchema.DataSources {
		if ds.GetType() == "mongodb" {
			return ds.(*MongoDataSource).Mappings
		}
	}
	return nil
}*/
