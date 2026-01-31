package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStaticEntityInstanceRepository provides MongoDB implementation for StaticEntityInstanceRepositoryInterface.
// This implementation works directly with the model type T without the EntityInstance[T] wrapper,
// offering a simpler interface for cases where metadata functionality is not required.
type MongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	collection *mongo.Collection
	options    sdk.StaticEntityInstanceRepositoryOptions
}

// NewMongoStaticEntityInstanceRepository creates a new MongoDB-based static repository
func NewMongoStaticEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, options sdk.StaticEntityInstanceRepositoryOptions) *MongoStaticEntityInstanceRepository[T] {
	client, _ := sdk.GetMongoClient()

	collection := client.Database(sdk_configuration.GetConfig().DynamicEntityDocumentDBName).Collection(entityId)
	return &MongoStaticEntityInstanceRepository[T]{
		collection: collection,
		options:    options,
	}
}

func (r *MongoStaticEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (T, error) {
	var zero T
	var result T

	// Preparazione del filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return zero, sdk.NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Esegui la query e fai decode diretto nella struct T
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return zero, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
		}
		return zero, sdk.NewInternalServerError(fmt.Errorf("failed to find entity instance: %w", err))
	}

	// Imposta l'ID string nel modello (già convertito da BSON)
	result.SetID(dto.Id)

	return result, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]T, error) {
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
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to list entities: %w", err))
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, sdk.NewInternalServerError(fmt.Errorf("failed to decode entities: %w", err))
	}

	// Imposta gli ID per ogni risultato
	for i := range results {
		if *r.options.AutoGenerateID {
			// L'ID sarà già presente come ObjectID, ma dobbiamo convertirlo in stringa per SetID
			if idPtr := results[i].GetID(); idPtr != "" {
				results[i].SetID(idPtr)
			}
		}
	}

	return results, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[T]) (T, error) {
	var zero T
	idPtr := dto.Data.GetID()

	if *r.options.AutoGenerateID {
		oid := primitive.NewObjectID()
		idStr := oid.Hex()
		dto.Data.SetID(idStr)

		// Serializza la struct e imposta l'_id come ObjectID
		docBytes, err := bson.Marshal(dto.Data)
		if err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
		}
		var doc bson.M
		if err := bson.Unmarshal(docBytes, &doc); err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
		}
		doc["_id"] = oid // Sostituisci l'ID stringa con ObjectID

		_, err = r.collection.InsertOne(ctx, doc)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return zero, sdk.NewConflictError(fmt.Errorf("entity instance already exists: %w", err))
			}
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to create entity instance: %w", err))
		}
	} else {
		if idPtr == "" {
			return zero, sdk.NewBadRequestError(fmt.Errorf("ID is required when auto-generation is disabled"))
		}

		// Verifica che l'ID non esista già
		_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: idPtr})
		if err == nil {
			return zero, sdk.NewConflictError(fmt.Errorf("entity instance with id %v already exists", idPtr))
		}
		var endorErr *sdk.EndorError
		if errors.As(err, &endorErr) && endorErr.StatusCode != 404 {
			return zero, err
		}

		// Per ID manuali, inserisci direttamente la struct (l'ID rimane stringa)
		_, err = r.collection.InsertOne(ctx, dto.Data)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return zero, sdk.NewConflictError(fmt.Errorf("entity instance already exists: %w", err))
			}
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to create entity instance: %w", err))
		}
	}

	return dto.Data, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) Replace(ctx context.Context, dto sdk.ReplaceByIdDTO[T]) (T, error) {
	var zero T

	// Verifica che l'istanza esista
	_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return zero, err
	}

	// Prepara il filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return zero, sdk.NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}

		// Assicurati che l'ID sia impostato nella struct
		dto.Data.SetID(dto.Id)

		// Serializza la struct e imposta l'_id come ObjectID
		docBytes, err := bson.Marshal(dto.Data)
		if err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to marshal entity: %w", err))
		}
		var doc bson.M
		if err := bson.Unmarshal(docBytes, &doc); err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to unmarshal entity: %w", err))
		}
		doc["_id"] = objectID // Sostituisci l'ID stringa con ObjectID

		result, err := r.collection.ReplaceOne(ctx, filter, doc)
		if err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to replace entity instance: %w", err))
		}
		if result.MatchedCount == 0 {
			return zero, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
		}
	} else {
		filter = bson.M{"_id": dto.Id}

		// Assicurati che l'ID sia impostato nella struct
		dto.Data.SetID(dto.Id)

		// Per ID manuali, aggiorna direttamente con la struct
		result, err := r.collection.ReplaceOne(ctx, filter, dto.Data)
		if err != nil {
			return zero, sdk.NewInternalServerError(fmt.Errorf("failed to replace entity instance: %w", err))
		}
		if result.MatchedCount == 0 {
			return zero, sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
		}
	}

	return dto.Data, nil
}

func (r *MongoStaticEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	// Verifica che l'istanza esista
	_, err := r.Instance(ctx, sdk.ReadInstanceDTO{Id: dto.Id})
	if err != nil {
		return err
	}

	// Prepara il filtro
	var filter bson.M
	if *r.options.AutoGenerateID {
		objectID, err := primitive.ObjectIDFromHex(dto.Id)
		if err != nil {
			return sdk.NewBadRequestError(fmt.Errorf("invalid ObjectID format: %w", err))
		}
		filter = bson.M{"_id": objectID}
	} else {
		filter = bson.M{"_id": dto.Id}
	}

	// Elimina il documento
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return sdk.NewInternalServerError(fmt.Errorf("failed to delete entity instance: %w", err))
	}

	if result.DeletedCount == 0 {
		return sdk.NewNotFoundError(fmt.Errorf("entity instance with id %v not found", dto.Id))
	}

	return nil
}
