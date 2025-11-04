package sdk

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoAbstractResourceRepository struct {
	collection        *mongo.Collection
	def               ResourceDefinition
	currentDatasource MongoDataSource
	context           context.Context
}

func NewMongoAbstractResourceRepository(dbName string, def ResourceDefinition, currentDatasource MongoDataSource) (*MongoAbstractResourceRepository, error) {
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

	client, _ := GetMongoClient()

	coll := client.Database(dbName).Collection(mongoDS.Collection)
	return &MongoAbstractResourceRepository{
		collection:        coll,
		def:               def,
		currentDatasource: currentDatasource,
		context:           context.TODO(),
	}, nil
}

func (r *MongoAbstractResourceRepository) Instance(dto ReadInstanceDTO) (any, error) {
	instance := bson.M{}
	idMapping, err := r.getIdMapping()
	if err != nil {
		return nil, err
	}
	filter := bson.M{idMapping.Path: dto.Id}
	var opts *options.FindOneOptions
	if !r.has_IdPath() {
		opts = options.FindOne().SetProjection(bson.M{"_id": 0})
	}
	err = r.collection.FindOne(r.context, filter, opts).Decode(&instance)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewNotFoundError(err)
		} else {
			return nil, NewInternalServerError(err)
		}
	}
	return &instance, nil
}

func (r *MongoAbstractResourceRepository) List(dto ReadDTO) ([]any, error) {
	var opts *options.FindOptions
	if !r.has_IdPath() {
		opts = options.Find().SetProjection(dto.Projection)
	}
	cursor, err := r.collection.Find(r.context, dto.Filter, opts)
	if err != nil {
		return nil, NewInternalServerError(err)
	}
	var storedResources []bson.M
	if err := cursor.All(r.context, &storedResources); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []any{}, nil
		} else {
			return nil, NewInternalServerError(err)
		}
	}
	result := make([]any, len(storedResources))
	for i, doc := range storedResources {
		result[i] = doc
	}
	return result, nil
}

func (r *MongoAbstractResourceRepository) Create(dto CreateDTO[any]) error {
	// get document
	doc, err := r.getDocument(dto.Data)
	if err != nil {
		return err
	}
	// check if already exist
	idValue, ok := doc[r.def.Id].(string)
	if !ok {
		return NewBadRequestError(fmt.Errorf("document id is not a string"))
	}
	_, err = r.Instance(ReadInstanceDTO{Id: idValue})
	if err != nil {
		var endorError *EndorError
		if errors.As(err, &endorError) && endorError.StatusCode == 404 {
			// TODO: create real document based on path name, not property name
			// for now skip. Now document has the name fields as schema properties
			_, err = r.collection.InsertOne(context.Background(), doc)
			return err
		} else {
			return err
		}
	}
	return NewConflictError(fmt.Errorf("resource already exist"))
}

func (r *MongoAbstractResourceRepository) Delete(dto DeleteByIdDTO) error {
	// check if already exist
	_, err := r.Instance(ReadInstanceDTO(dto))
	if err != nil {
		return err
	}
	idMapping, err := r.getIdMapping()
	if err != nil {
		return err
	}
	filter := bson.M{idMapping.Path: dto.Id}
	_, err = r.collection.DeleteOne(context.Background(), filter)
	return err
}

func (r *MongoAbstractResourceRepository) Update(dto UpdateByIdDTO[any]) (any, error) {
	// Check if the resource exists
	_, err := r.Instance(ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return nil, err
	}

	// Convert input to bson.M
	doc, err := r.getDocument(dto.Data)
	if err != nil {
		return nil, err
	}

	// Get the ID field mapping
	idMapping, err := r.getIdMapping()
	if err != nil {
		return nil, err
	}

	// Build the filter and update
	filter := bson.M{idMapping.Path: dto.Id}
	update := bson.M{"$set": doc}

	// Perform the update
	_, err = r.collection.UpdateOne(r.context, filter, update)
	if err != nil {
		return nil, err
	}

	// Return the updated instance
	return dto.Data, nil
}

func (f *MongoAbstractResourceRepository) getDocument(input any) (bson.M, error) {
	data, err := bson.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}
	var result bson.M
	if err := bson.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to bson.M: %w", err)
	}
	return result, nil
}

func (r *MongoAbstractResourceRepository) getIdMapping() (*MongoFieldMapping, error) {
	// create id filter
	idMapping := r.getMapping(r.def.Id)
	if idMapping == nil {
		return nil, NewNotFoundError(fmt.Errorf("mapping not found"))
	}
	return idMapping, nil
}

func (r *MongoAbstractResourceRepository) getMapping(propertyName string) *MongoFieldMapping {
	for name, mapping := range r.currentDatasource.Mappings {
		if name == propertyName {
			return &mapping
		}
	}
	return nil
}

func (r *MongoAbstractResourceRepository) has_IdPath() bool {
	for _, v := range r.currentDatasource.Mappings {
		if v.Path == "_id" {
			return true
		}
	}
	return false
}
