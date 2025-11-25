package sdk

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoResourceInstanceRepository[T ResourceInstanceInterface] struct {
	collectionName string
	options        ResourceInstanceRepositoryOptions
}

func NewMongoResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options ResourceInstanceRepositoryOptions) *MongoResourceInstanceRepository[T] {
	return &MongoResourceInstanceRepository[T]{
		collectionName: resourceId,
		options:        options,
	}
}

func (r *MongoResourceInstanceRepository[T]) Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error) {
	var rawResult bson.M

	// Preparazione del filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return nil, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Esegui la query
	err := r.getCollection().FindOne(ctx, filter).Decode(&rawResult)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
		return nil, NewInternalServerError(fmt.Errorf("failed to find resource instance: %w", err))
	}

	// Estrai e converti i metadata in modo robusto
	metadata := make(map[string]interface{})
	if rawMeta, ok := rawResult["metadata"]; ok && rawMeta != nil {
		metaBytes, err := bson.Marshal(rawMeta)
		if err == nil {
			_ = bson.Unmarshal(metaBytes, &metadata)
		}
	}
	delete(rawResult, "metadata")

	// Estrai e converti _id in stringa
	var idStr string
	if *r.options.AutoGenerateID {
		if oid, ok := rawResult["_id"].(primitive.ObjectID); ok {
			idStr = oid.Hex()
		} else {
			return nil, NewInternalServerError(fmt.Errorf("invalid _id type in database"))
		}
	} else {
		if s, ok := rawResult["_id"].(string); ok {
			idStr = s
		} else {
			return nil, NewInternalServerError(fmt.Errorf("invalid _id type in database"))
		}
	}
	delete(rawResult, "_id")

	// Mappa i campi rimanenti nel modello T
	var thisModel T
	resourceBytes, err := bson.Marshal(rawResult)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to marshal raw resource: %w", err))
	}
	if err := bson.Unmarshal(resourceBytes, &thisModel); err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to unmarshal to model: %w", err))
	}

	// Imposta l'ID string nel modello
	thisModel.SetID(idStr)

	return &ResourceInstance[T]{
		This:     thisModel,
		Metadata: metadata,
	}, nil
}

func (r *MongoResourceInstanceRepository[T]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error) {
	mongoFilter, err := prepareFilter[T](dto.Filter)
	if err != nil {
		return nil, err
	}
	opts := options.Find().SetProjection(prepareProjection[T](dto.Projection))
	cursor, err := r.getCollection().Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to list resources: %w", err))
	}
	defer cursor.Close(ctx)

	var rawResults []bson.M
	if err := cursor.All(ctx, &rawResults); err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to decode resources: %w", err))
	}

	results := make([]ResourceInstance[T], 0, len(rawResults))
	for _, raw := range rawResults {
		// Estrai e converti metadata in modo robusto
		metadata := make(map[string]interface{})
		if rawMeta, ok := raw["metadata"]; ok && rawMeta != nil {
			metaBytes, err := bson.Marshal(rawMeta)
			if err == nil {
				_ = bson.Unmarshal(metaBytes, &metadata)
			}
		}
		delete(raw, "metadata")

		// Estrai e converti _id in stringa
		var idStr string
		if *r.options.AutoGenerateID {
			if oid, ok := raw["_id"].(primitive.ObjectID); ok {
				idStr = oid.Hex()
			} else {
				return nil, NewInternalServerError(fmt.Errorf("invalid _id type in database"))
			}
		} else {
			if s, ok := raw["_id"].(string); ok {
				idStr = s
			} else {
				return nil, NewInternalServerError(fmt.Errorf("invalid _id type in database"))
			}
		}
		delete(raw, "_id")

		// Mappa i campi rimanenti nel modello T
		var thisModel T
		resourceBytes, err := bson.Marshal(raw)
		if err != nil {
			return nil, NewInternalServerError(fmt.Errorf("failed to marshal raw resource: %w", err))
		}
		if err := bson.Unmarshal(resourceBytes, &thisModel); err != nil {
			return nil, NewInternalServerError(fmt.Errorf("failed to unmarshal to model: %w", err))
		}

		// Imposta l'ID string nel modello
		thisModel.SetID(idStr)

		results = append(results, ResourceInstance[T]{
			This:     thisModel,
			Metadata: metadata,
		})
	}

	return results, nil
}

func (r *MongoResourceInstanceRepository[T]) Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	idPtr := dto.Data.This.GetID()
	var idStr string

	if *r.options.AutoGenerateID {
		oid := primitive.NewObjectID()
		idStr = oid.Hex()
		dto.Data.This.SetID(idStr)
	} else {
		if idPtr == nil || *idPtr == "" {
			return nil, NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}
		idStr = *idPtr

		_, err := r.Instance(ctx, ReadInstanceDTO{Id: idStr})
		if err == nil {
			return nil, NewConflictError(fmt.Errorf("resource instance with id %v already exists", idStr))
		}
		var endorErr *EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return nil, err
		}
	}

	resourceBytes, err := bson.Marshal(dto.Data.This)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to marshal resource: %w", err))
	}
	var resourceMap bson.M
	if err := bson.Unmarshal(resourceBytes, &resourceMap); err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to unmarshal resource: %w", err))
	}

	// _id
	if *r.options.AutoGenerateID {
		oid, _ := primitive.ObjectIDFromHex(idStr)
		resourceMap["_id"] = oid
	} else {
		resourceMap["_id"] = idStr
	}

	// Aggiungi metadata in oggetto separato
	resourceMap["metadata"] = dto.Data.Metadata

	_, err = r.getCollection().InsertOne(ctx, resourceMap)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, NewConflictError(fmt.Errorf("resource instance already exists: %w", err))
		}
		return nil, NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
	}

	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[T]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return nil, err
	}

	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return nil, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	resourceBytes, err := bson.Marshal(dto.Data.This)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to marshal resource: %w", err))
	}

	var resourceMap bson.M
	if err := bson.Unmarshal(resourceBytes, &resourceMap); err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to unmarshal resource: %w", err))
	}

	if *r.options.AutoGenerateID {
		objectID, _ := primitive.ObjectIDFromHex(dto.Id)
		resourceMap["_id"] = objectID
	} else {
		resourceMap["_id"] = dto.Id
	}

	resourceMap["metadata"] = dto.Data.Metadata

	result, err := r.getCollection().ReplaceOne(ctx, filter, resourceMap)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to update resource instance: %w", err))
	}
	if result.MatchedCount == 0 {
		return nil, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	dto.Data.This.SetID(dto.Id)
	return &dto.Data, nil
}

func (r *MongoResourceInstanceRepository[T]) Delete(ctx context.Context, dto ReadInstanceDTO) error {
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return err
	}

	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	result, err := r.getCollection().DeleteOne(ctx, filter)
	if err != nil {
		return NewInternalServerError(fmt.Errorf("failed to delete resource instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	return nil
}

func prepareProjection[T ResourceInstanceInterface](projection map[string]interface{}) bson.M {
	result := bson.M{}

	// Creiamo un'istanza vuota di T per capire quali campi sono "normali"
	var thisModel T
	modelFields := map[string]struct{}{}
	b, _ := bson.Marshal(thisModel)
	_ = bson.Unmarshal(b, &modelFields) // otteniamo tutti i campi di T

	for k, v := range projection {
		if _, ok := modelFields[k]; ok {
			// Campo normale, rimane a livello root
			result[k] = v
		} else {
			// Campo extra → va dentro metadata
			result["metadata."+k] = v
		}
	}

	return result
}

func prepareFilter[T ResourceInstanceInterface](filter map[string]interface{}) (bson.M, error) {
	result := bson.M{}

	// Creiamo un'istanza vuota di T per capire quali campi sono "normali"
	var thisModel T
	modelFields := map[string]struct{}{}
	b, _ := bson.Marshal(thisModel)
	_ = bson.Unmarshal(b, &modelFields) // otteniamo tutti i campi di T

	for k, v := range filter {
		if _, ok := modelFields[k]; ok {
			// Campo normale, resta a livello root
			result[k] = v
		} else {
			// Campo extra → va dentro metadata
			result["metadata."+k] = v
		}
	}

	return result, nil
}

func (r *MongoResourceInstanceRepository[T]) getCollection() *mongo.Collection {
	client, _ := GetMongoClient()
	return client.Database(GetConfig().DynamicResourceDocumentDBName).Collection(r.collectionName)
}

// ==== SPECIALIZED REPOSITORY IMPLEMENTATION ====

type MongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C any] struct {
	collectionName string
	options        ResourceInstanceRepositoryOptions
	categoryInfo   *CategorySpecialized[C]
}

func NewMongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C any](
	resourceId string,
	options ResourceInstanceRepositoryOptions,
	categoryInfo *CategorySpecialized[C],
) *MongoResourceInstanceSpecializedRepository[T, C] {
	return &MongoResourceInstanceSpecializedRepository[T, C]{
		collectionName: resourceId,
		options:        options,
		categoryInfo:   categoryInfo,
	}
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstanceSpecialized[T, C], error) {
	var rawResult bson.M

	// Preparazione del filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return nil, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Esegui la query
	err := r.getCollection().FindOne(ctx, filter).Decode(&rawResult)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
		return nil, NewInternalServerError(fmt.Errorf("failed to find resource instance: %w", err))
	}

	// Converti in ResourceInstanceSpecialized
	return r.convertToSpecialized(rawResult)
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstanceSpecialized[T, C], error) {
	// Usa la stessa logica del repository base ma converte in specialized
	cursor, err := r.getCollection().Find(ctx, bson.M{})
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to find resource instances: %w", err))
	}
	defer cursor.Close(ctx)

	var rawResults []bson.M
	if err = cursor.All(ctx, &rawResults); err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to decode resource instances: %w", err))
	}

	specializedResults := make([]ResourceInstanceSpecialized[T, C], len(rawResults))
	for i, rawResult := range rawResults {
		converted, err := r.convertToSpecialized(rawResult)
		if err != nil {
			return nil, err
		}
		specializedResults[i] = *converted
	}

	return specializedResults, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Create(ctx context.Context, dto CreateDTO[ResourceInstanceSpecialized[T, C]]) (*ResourceInstanceSpecialized[T, C], error) {
	// Converte ResourceInstanceSpecialized in documento BSON
	doc, err := r.convertFromSpecialized(dto.Data)
	if err != nil {
		return nil, err
	}

	// Genera ID se necessario
	if *r.options.AutoGenerateID {
		idValue := dto.Data.This.GetID()
		if idValue == nil || *idValue == "" {
			objectID := primitive.NewObjectID()
			dto.Data.This.SetID(objectID.Hex())
			doc["_id"] = objectID
		}
	}

	// Inserisci in MongoDB
	_, err = r.getCollection().InsertOne(ctx, doc)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
	}

	return &dto.Data, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Delete(ctx context.Context, dto ReadInstanceDTO) error {
	// Usa la stessa logica del repository base
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	result, err := r.getCollection().DeleteOne(ctx, filter)
	if err != nil {
		return NewInternalServerError(fmt.Errorf("failed to delete resource instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	return nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstanceSpecialized[T, C]]) (*ResourceInstanceSpecialized[T, C], error) {
	// Converte in documento BSON
	doc, err := r.convertFromSpecialized(dto.Data)
	if err != nil {
		return nil, err
	}

	// Preparazione del filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return nil, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Update in MongoDB
	update := bson.M{"$set": doc}
	_, err = r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to update resource instance: %w", err))
	}

	return &dto.Data, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) getCollection() *mongo.Collection {
	client, _ := GetMongoClient()
	return client.Database(GetConfig().DynamicResourceDocumentDBName).Collection(r.collectionName)
}

// convertToSpecialized converte documento BSON in ResourceInstanceSpecialized
func (r *MongoResourceInstanceSpecializedRepository[T, C]) convertToSpecialized(rawResult bson.M) (*ResourceInstanceSpecialized[T, C], error) {
	// Estrai base metadata
	baseMetadata := extractMetadataFromBSON(rawResult, "metadata")

	// Estrai specialization
	var specialization *ResourceInstanceSpecialization[C]
	if specData, exists := rawResult["specialization"]; exists {
		if specMap, ok := specData.(bson.M); ok {
			// Estrai This (modello statico categoria)
			var categoryThis C
			if thisData, exists := specMap["this"]; exists {
				bsonBytes, _ := bson.Marshal(thisData)
				bson.Unmarshal(bsonBytes, &categoryThis)
			}

			// Estrai metadata categoria
			categoryMetadata := extractMetadataFromBSON(specMap, "metadata")

			specialization = &ResourceInstanceSpecialization[C]{
				This:     categoryThis,
				Metadata: categoryMetadata,
			}
		}
	}

	// Estrai base This
	var baseThis T
	if thisData, exists := rawResult["this"]; exists {
		bsonBytes, _ := bson.Marshal(thisData)
		bson.Unmarshal(bsonBytes, &baseThis)
	}

	return &ResourceInstanceSpecialized[T, C]{
		This:           baseThis,
		Metadata:       baseMetadata,
		Specialization: specialization,
	}, nil
}

// convertFromSpecialized converte ResourceInstanceSpecialized in documento BSON
func (r *MongoResourceInstanceSpecializedRepository[T, C]) convertFromSpecialized(data ResourceInstanceSpecialized[T, C]) (bson.M, error) {
	doc := bson.M{
		"this":     data.This,
		"metadata": data.Metadata,
	}

	// Aggiungi specialization se presente
	if data.Specialization != nil {
		doc["specialization"] = bson.M{
			"this":     data.Specialization.This,
			"metadata": data.Specialization.Metadata,
		}
	}

	return doc, nil
}

// ==== COMMON UTILITIES ====

// extractMetadataFromBSON estrae metadata da un documento BSON
func extractMetadataFromBSON(rawResult bson.M, metadataField string) map[string]interface{} {
	metadata := make(map[string]interface{})
	if rawMeta, ok := rawResult[metadataField]; ok && rawMeta != nil {
		if metaBytes, err := bson.Marshal(rawMeta); err == nil {
			_ = bson.Unmarshal(metaBytes, &metadata)
		}
	}
	return metadata
}
