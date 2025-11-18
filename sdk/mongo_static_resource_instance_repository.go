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

// MongoStaticResourceInstanceRepository provides MongoDB implementation for StaticResourceInstanceRepositoryInterface.
// This implementation works directly with the model type T without the ResourceInstance[T] wrapper,
// offering a simpler interface for cases where metadata functionality is not required.
type MongoStaticResourceInstanceRepository[T ResourceInstanceInterface] struct {
	collection *mongo.Collection
	options    StaticResourceInstanceRepositoryOptions
}

// NewMongoStaticResourceInstanceRepository creates a new MongoDB-based static repository
func NewMongoStaticResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options StaticResourceInstanceRepositoryOptions) *MongoStaticResourceInstanceRepository[T] {
	client, _ := GetMongoClient()

	collection := client.Database(GetConfig().DynamicResourceDocumentDBName).Collection(resourceId)
	return &MongoStaticResourceInstanceRepository[T]{
		collection: collection,
		options:    options,
	}
}

func (r *MongoStaticResourceInstanceRepository[T]) Instance(ctx context.Context, dto ReadInstanceDTO) (T, error) {
	var zero T
	var result T

	// Preparazione del filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return zero, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Esegui la query e fai decode diretto nella struct T
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return zero, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
		return zero, NewInternalServerError(fmt.Errorf("failed to find resource instance: %w", err))
	}

	// Imposta l'ID string nel modello (già convertito da BSON)
	result.SetID(dto.Id)

	return result, nil
}

func (r *MongoStaticResourceInstanceRepository[T]) List(ctx context.Context, dto ReadDTO) ([]T, error) {
	// Usa filtro e projezione direttamente dai DTO (semplificati)
	mongoFilter := dto.Filter
	if mongoFilter == nil {
		mongoFilter = bson.M{}
	}

	var opts *options.FindOptions
	if dto.Projection != nil {
		opts = options.Find().SetProjection(dto.Projection)
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to list resources: %w", err))
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, NewInternalServerError(fmt.Errorf("failed to decode resources: %w", err))
	}

	// Imposta gli ID per ogni risultato
	for i := range results {
		if *r.options.AutoGenerateID {
			// L'ID sarà già presente come ObjectID, ma dobbiamo convertirlo in stringa per SetID
			if idPtr := results[i].GetID(); idPtr != nil {
				results[i].SetID(*idPtr)
			}
		}
	}

	return results, nil
}

func (r *MongoStaticResourceInstanceRepository[T]) Create(ctx context.Context, dto CreateDTO[T]) (T, error) {
	var zero T
	idPtr := dto.Data.GetID()

	if *r.options.AutoGenerateID {
		oid := primitive.NewObjectID()
		idStr := oid.Hex()
		dto.Data.SetID(idStr)

		// Serializza la struct e imposta l'_id come ObjectID
		docBytes, err := bson.Marshal(dto.Data)
		if err != nil {
			return zero, NewInternalServerError(fmt.Errorf("failed to marshal resource: %w", err))
		}
		var doc bson.M
		if err := bson.Unmarshal(docBytes, &doc); err != nil {
			return zero, NewInternalServerError(fmt.Errorf("failed to unmarshal resource: %w", err))
		}
		doc["_id"] = oid // Sostituisci l'ID stringa con ObjectID

		_, err = r.collection.InsertOne(ctx, doc)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return zero, NewConflictError(fmt.Errorf("resource instance already exists: %w", err))
			}
			return zero, NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
		}
	} else {
		if idPtr == nil || *idPtr == "" {
			return zero, NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}

		// Verifica che l'ID non esista già
		_, err := r.Instance(ctx, ReadInstanceDTO{Id: *idPtr})
		if err == nil {
			return zero, NewConflictError(fmt.Errorf("resource instance with id %v already exists", *idPtr))
		}
		var endorErr *EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return zero, err
		}

		// Per ID manuali, inserisci direttamente la struct (l'ID rimane stringa)
		_, err = r.collection.InsertOne(ctx, dto.Data)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return zero, NewConflictError(fmt.Errorf("resource instance already exists: %w", err))
			}
			return zero, NewInternalServerError(fmt.Errorf("failed to create resource instance: %w", err))
		}
	}

	return dto.Data, nil
}

func (r *MongoStaticResourceInstanceRepository[T]) Update(ctx context.Context, dto UpdateByIdDTO[T]) (T, error) {
	var zero T

	// Verifica che l'istanza esista
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return zero, err
	}

	// Prepara il filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return zero, NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}

		// Assicurati che l'ID sia impostato nella struct
		dto.Data.SetID(dto.Id)

		// Serializza la struct e imposta l'_id come ObjectID
		docBytes, err := bson.Marshal(dto.Data)
		if err != nil {
			return zero, NewInternalServerError(fmt.Errorf("failed to marshal resource: %w", err))
		}
		var doc bson.M
		if err := bson.Unmarshal(docBytes, &doc); err != nil {
			return zero, NewInternalServerError(fmt.Errorf("failed to unmarshal resource: %w", err))
		}
		doc["_id"] = objectID // Sostituisci l'ID stringa con ObjectID

		result, err := r.collection.ReplaceOne(ctx, filter, doc)
		if err != nil {
			return zero, NewInternalServerError(fmt.Errorf("failed to update resource instance: %w", err))
		}
		if result.MatchedCount == 0 {
			return zero, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
	} else {
		filter = bson.M{"_id": dto.Id}

		// Assicurati che l'ID sia impostato nella struct
		dto.Data.SetID(dto.Id)

		// Per ID manuali, aggiorna direttamente con la struct
		result, err := r.collection.ReplaceOne(ctx, filter, dto.Data)
		if err != nil {
			return zero, NewInternalServerError(fmt.Errorf("failed to update resource instance: %w", err))
		}
		if result.MatchedCount == 0 {
			return zero, NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
		}
	}

	return dto.Data, nil
}

func (r *MongoStaticResourceInstanceRepository[T]) Delete(ctx context.Context, dto ReadInstanceDTO) error {
	// Verifica che l'istanza esista
	_, err := r.Instance(ctx, ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return err
	}

	// Prepara il filtro
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

	// Elimina il documento
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return NewInternalServerError(fmt.Errorf("failed to delete resource instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return NewNotFoundError(fmt.Errorf("resource instance with id %v not found", dto.Id))
	}

	return nil
}
