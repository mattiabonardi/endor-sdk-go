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

type MongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface] struct {
	collectionName string
	options        ResourceInstanceRepositoryOptions
}

func NewMongoResourceInstanceSpecializedRepository[T ResourceInstanceInterface, C ResourceInstanceSpecializedInterface](
	resourceId string,
	options ResourceInstanceRepositoryOptions,
) *MongoResourceInstanceSpecializedRepository[T, C] {
	return &MongoResourceInstanceSpecializedRepository[T, C]{
		collectionName: resourceId,
		options:        options,
	}
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstanceSpecialized[T, C], error) {
	var rawResult bson.M

	// Preparazione filtro
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

	// FindOne
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
	doc, err := r.convertFromSpecialized(dto.Data)
	if err != nil {
		return nil, err
	}

	// Genera _id se necessario
	if *r.options.AutoGenerateID {
		idValue := dto.Data.This.GetID()
		if idValue == nil || *idValue == "" {
			objectID := primitive.NewObjectID()
			dto.Data.This.SetID(objectID.Hex())
			doc["_id"] = objectID
		}
	}

	_, err = r.getCollection().InsertOne(ctx, doc)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
	}

	return &dto.Data, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) Delete(ctx context.Context, dto ReadInstanceDTO) error {
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
	doc, err := r.convertFromSpecialized(dto.Data)
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

func (r *MongoResourceInstanceSpecializedRepository[T, C]) convertToSpecialized(rawResult bson.M) (*ResourceInstanceSpecialized[T, C], error) {
	metadata := make(map[string]interface{})
	if rawMeta, ok := rawResult["metadata"]; ok && rawMeta != nil {
		switch m := rawMeta.(type) {
		case bson.M:
			for k, v := range m {
				metadata[k] = v
			}
		case map[string]interface{}:
			for k, v := range m {
				metadata[k] = v
			}
		default:
			if b, err := bson.Marshal(m); err == nil {
				_ = bson.Unmarshal(b, &metadata)
			}
		}
	}

	// copia flat senza metadata
	docNoMeta := cloneBsonM(rawResult)
	delete(docNoMeta, "metadata")

	// This
	var baseThis T
	{
		b, err := bson.Marshal(docNoMeta)
		if err != nil {
			return nil, fmt.Errorf("marshal docNoMeta: %w", err)
		}
		if err := bson.Unmarshal(b, &baseThis); err != nil {
			return nil, fmt.Errorf("unmarshal into This: %w", err)
		}

		// _id -> This.ID se vuoto
		if rawId, ok := rawResult["_id"]; ok {
			if ri, ok := any(&baseThis).(interface {
				GetID() *string
				SetID(string)
			}); ok {
				if idPtr := ri.GetID(); idPtr == nil || *idPtr == "" {
					switch v := rawId.(type) {
					case primitive.ObjectID:
						ri.SetID(v.Hex())
					case string:
						ri.SetID(v)
					}
				}
			}
		}
	}

	// CategoryThis
	var categoryThis C
	{
		b, err := bson.Marshal(docNoMeta)
		if err != nil {
			return nil, fmt.Errorf("marshal docNoMeta for CategoryThis: %w", err)
		}
		if err := bson.Unmarshal(b, &categoryThis); err != nil {
			return nil, fmt.Errorf("unmarshal into CategoryThis: %w", err)
		}
	}

	res := &ResourceInstanceSpecialized[T, C]{
		This:         baseThis,
		CategoryThis: categoryThis,
		Metadata:     metadata,
	}
	return res, nil
}

func (r *MongoResourceInstanceSpecializedRepository[T, C]) convertFromSpecialized(data ResourceInstanceSpecialized[T, C]) (bson.M, error) {
	doc := bson.M{}

	// merge This
	{
		bt, err := bson.Marshal(data.This)
		if err != nil {
			return nil, fmt.Errorf("marshal This: %w", err)
		}
		var tMap bson.M
		if err := bson.Unmarshal(bt, &tMap); err != nil {
			return nil, fmt.Errorf("unmarshal This->map: %w", err)
		}
		for k, v := range tMap {
			doc[k] = v
		}
	}

	// merge CategoryThis
	{
		bc, err := bson.Marshal(data.CategoryThis)
		if err != nil {
			return nil, fmt.Errorf("marshal CategoryThis: %w", err)
		}
		var cMap bson.M
		if err := bson.Unmarshal(bc, &cMap); err != nil {
			return nil, fmt.Errorf("unmarshal CategoryThis->map: %w", err)
		}
		for k, v := range cMap {
			doc[k] = v
		}
	}

	// metadata come subdocumento
	if data.Metadata != nil {
		doc["metadata"] = data.Metadata
	}

	return doc, nil
}

// helper shallow copy
func cloneBsonM(src bson.M) bson.M {
	dst := make(bson.M, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
