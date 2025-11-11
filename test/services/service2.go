package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

type TestHybridPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Service2 struct {
}

func (h *Service2) hybridTest(c *sdk.EndorContext[TestHybridPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.Info, "Hello from Hybrid Service")).
		Build(), nil
}

func (h *Service2) getSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.Info, "Schema retrieved from hybrid service")).
		Build(), nil
}

func NewService2() sdk.EndorHybridService {
	service2 := Service2{}

	return sdk.NewHybridService("test-hybrid", "Testing hybrid resource").
		WithActions(func(getSchema func() sdk.Schema) map[string]sdk.EndorServiceAction {
			// Qui possiamo accedere allo schema dinamico tramite getSchema() se necessario
			// schema := getSchema()

			return map[string]sdk.EndorServiceAction{
				"hybrid-test": sdk.NewAction(
					service2.hybridTest,
					"Test hybrid action with dynamic schema",
				),
				"get-schema": sdk.NewAction(
					service2.getSchema,
					"Get the current schema of the hybrid resource",
				),
			}
		})
}
