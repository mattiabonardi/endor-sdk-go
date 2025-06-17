package sdk

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type AbstractResourceService struct {
	resource    string
	definition  ResourceDefinition
	mongoClient *mongo.Client
	mongoDB     string
}

func NewAbstractResourceService(resource string, description string, definition ResourceDefinition, mongoClient *mongo.Client, mongoDB string) EndorService {
	service := AbstractResourceService{
		resource:    resource,
		definition:  definition,
		mongoClient: mongoClient,
		mongoDB:     mongoDB,
	}
	return EndorService{
		Resource:    resource,
		Description: description,
		Methods: map[string]EndorServiceMethod{
			"list": NewMethod(
				service.list,
			),
			"create": NewMethod(
				service.create,
			),
			/*"instance": NewMethod(
				resourceService.instance,
			),
			"update": NewMethod(
				resourceService.update,
			),
			"delete": NewMethod(
				resourceService.delete,
			),*/
		},
	}
}

func (h *AbstractResourceService) list(c *EndorContext[NoPayload]) {
	c.End(NewResponseBuilder[[]any]().AddSchema(&h.definition.Schema).Build())
}

func (h *AbstractResourceService) create(c *EndorContext[CreateDTO[any]]) {
	repo, err := NewAbstractResourceRepository(h.definition, h.mongoClient, h.mongoDB)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	err = repo.Create(c.Payload)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			c.Conflict(err)
			return
		} else {
			c.InternalServerError(err)
			return
		}
	}
	c.End(NewResponseBuilder[any]().AddData(&c.Payload.Data).AddSchema(&h.definition.Schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.resource))).Build())
}
