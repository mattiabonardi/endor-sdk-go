package internal

import "go.mongodb.org/mongo-driver/mongo"

type Repository[T any] interface {
	Instance(id string, options IntanceOptions) (T, error)
	List(options ListOptions) ([]T, error)
	Create(resource T) (T, error)
	Update(id string, resource T) (T, error)
	Delete(id string) error
}

type IntanceOptions struct {
	Projection Projection
}

type ListOptions struct {
	Filter     Filter
	Projection Projection
	Pagination Pagination
}

type Filter struct {
}

type Projection struct {
}

type Pagination struct {
}

type RepositoryAdapters struct {
	MongoClient *mongo.Client
}
