package sdk

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

type EndorHybridService struct {
	Resource       string
	Description    string
	Priority       *int
	metadataSchema Schema
	methodsFn      func() map[string]EndorServiceAction // Metodi wrapper che gestiscono tutte le categorie
	baseModel      ResourceInstanceInterface            // Modello base per ResourceInstanceRepository
	categories     map[string]Category                  // Cache delle categorie disponibili
}

func NewHybridService(resource, description string) EndorHybridService {
	return EndorHybridService{
		Resource:    resource,
		Description: description,
		baseModel:   &DynamicResource{}, // Default model
	}
}

// WithBaseModel permette di specificare un modello base personalizzato
func (h EndorHybridService) WithBaseModel(model ResourceInstanceInterface) EndorHybridService {
	h.baseModel = model
	return h
}

// Definizione dei metodi tramite funzione che restituisce metodi wrapper
// I metodi wrapper gestiranno tutte le categorie e verranno esplosi in ToEndorService
func (h EndorHybridService) WithActions(
	fn func() map[string]EndorServiceAction,
) EndorHybridService {
	h.methodsFn = fn
	return h
}

// WithCategories permette di configurare le categorie disponibili
func (h EndorHybridService) WithCategories(categories []Category) EndorHybridService {
	if h.categories == nil {
		h.categories = make(map[string]Category)
	}
	for _, category := range categories {
		h.categories[category.ID] = category
	}
	return h
}

// getDefaultActions fornisce le 6 azioni CRUD di default per hybrid services
func (h EndorHybridService) getDefaultActions(getSchema func() Schema) map[string]EndorServiceAction {
	// Crea repository usando DynamicResource come default (per ora)
	autogenerateID := true
	repository := NewResourceInstanceRepository[*DynamicResource](h.Resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return h.defaultSchema(c, getSchema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", h.Resource, h.Description),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[*DynamicResource]], error) {
				return h.defaultList(c, getSchema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s)", h.Resource, h.Description),
		),
		"create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", h.Resource, h.Description),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": h.getRootSchema(getSchema).Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
				return h.defaultCreate(c, getSchema, repository)
			},
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[*DynamicResource]], error) {
				return h.defaultInstance(c, getSchema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", h.Resource, h.Description),
		),
		"update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", h.Resource, h.Description),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": h.getRootSchema(getSchema).Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
				return h.defaultUpdate(c, getSchema, repository)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return h.defaultDelete(c, repository)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", h.Resource, h.Description),
		),
	}
}

// Conversione in EndorService, schema iniettato dal framework
func (h EndorHybridService) ToEndorService(attrs Schema) EndorService {
	h.metadataSchema = attrs
	getSchema := func() Schema { return h.metadataSchema }
	var methods = make(map[string]EndorServiceAction)

	// Se non ci sono categorie, usa i metodi di default normali
	if len(h.categories) == 0 {
		methods = h.getDefaultActions(getSchema)

		// Aggiungi metodi custom se esistono
		if h.methodsFn != nil {
			wrapperMethods := h.methodsFn()
			for methodName, wrapperAction := range wrapperMethods {
				methods[methodName] = h.createActionForCategory(wrapperAction, "", getSchema)
			}
		}
	} else {
		// Con categorie: esplodi sia i metodi di default che quelli custom

		// 1. Aggiungi i metodi di default base (senza categoria)
		defaultMethods := h.getDefaultActions(getSchema)
		for methodName, action := range defaultMethods {
			methods[methodName] = action
		}

		// 2. Esplodi i metodi di default per ogni categoria
		for categoryID, category := range h.categories {
			// Copia la categoria per evitare problemi di closure
			currentCategory := category
			currentCategoryID := categoryID

			// Crea lo schema combinato per questa categoria
			combinedSchemaFunc := func() Schema {
				return h.getCombinedSchemaForCategory(currentCategory, getSchema)
			}

			// Ottieni le azioni di default con lo schema della categoria
			categoryDefaultMethods := h.getDefaultActions(combinedSchemaFunc)

			for methodName, action := range categoryDefaultMethods {
				// Crea il nome del metodo esploso per default actions
				explodedMethodName := currentCategoryID + "/" + methodName
				methods[explodedMethodName] = h.createActionForCategory(action, currentCategoryID, combinedSchemaFunc)
			}
		}

		// 3. Esplodi i metodi custom per ogni categoria (se esistono)
		if h.methodsFn != nil {
			wrapperMethods := h.methodsFn()

			for methodName, wrapperAction := range wrapperMethods {
				// Aggiungi il metodo base (senza categoria)
				methods[methodName] = h.createActionForCategory(wrapperAction, "", getSchema)

				// Esplodi per ogni categoria
				for categoryID, category := range h.categories {
					// Copia la categoria per evitare problemi di closure
					currentCategory := category
					currentCategoryID := categoryID

					explodedMethodName := currentCategoryID + "/" + methodName
					combinedSchemaFunc := func() Schema {
						return h.getCombinedSchemaForCategory(currentCategory, getSchema)
					}
					methods[explodedMethodName] = h.createActionForCategory(wrapperAction, currentCategoryID, combinedSchemaFunc)
				}
			}
		}
	}

	return EndorService{
		Resource:    h.Resource,
		Description: h.Description,
		Priority:    h.Priority,
		Methods:     methods,
	}
}

// Helper method per creare rootSchema
func (h EndorHybridService) getRootSchema(getSchema func() Schema) *RootSchema {
	schema := getSchema()
	rootSchema := NewSchema(h.baseModel)
	if schema.Properties != nil {
		for k, v := range *schema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}
	return rootSchema
}

// Implementazioni dei metodi di default
func (h EndorHybridService) defaultSchema(_ *EndorContext[NoPayload], getSchema func() Schema) (*Response[any], error) {
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[any]().AddSchema(rootSchema).Build(), nil
}

func (h EndorHybridService) defaultInstance(c *EndorContext[ReadInstanceDTO], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[*ResourceInstance[*DynamicResource]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[*ResourceInstance[*DynamicResource]]().AddData(&instance).AddSchema(rootSchema).Build(), nil
}

func (h EndorHybridService) defaultList(c *EndorContext[ReadDTO], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[[]ResourceInstance[*DynamicResource]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[[]ResourceInstance[*DynamicResource]]().AddData(&list).AddSchema(rootSchema).Build(), nil
}

func (h EndorHybridService) defaultCreate(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[ResourceInstance[*DynamicResource]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(created).AddSchema(rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.Resource))).Build(), nil
}

func (h EndorHybridService) defaultUpdate(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[ResourceInstance[*DynamicResource]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(updated).AddSchema(rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.Resource))).Build(), nil
}

func (h EndorHybridService) defaultDelete(c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceRepository[*DynamicResource]) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.Resource))).Build(), nil
}

// createActionForCategory crea un'azione specifica per una categoria
func (h EndorHybridService) createActionForCategory(wrapperAction EndorServiceAction, categoryID string, getSchemaForCategory func() Schema) EndorServiceAction {
	// Ottieni le opzioni dell'action wrapper
	options := wrapperAction.GetOptions()

	// Aggiorna l'InputSchema se presente per riflettere lo schema della categoria
	newInputSchema := options.InputSchema
	if newInputSchema != nil && categoryID != "" {
		newInputSchema = h.updateInputSchemaForCategory(options.InputSchema, getSchemaForCategory)
	}

	// Crea nuove opzioni con lo schema specifico della categoria
	newOptions := EndorServiceActionOptions{
		Description:     options.Description,
		Public:          options.Public,
		ValidatePayload: options.ValidatePayload,
		InputSchema:     newInputSchema,
	}

	// Crea un wrapper che inietta il categoryID nel context
	callback := wrapperAction.CreateHTTPCallback
	newCallback := func(microserviceId string) func(c *gin.Context) {
		originalHandler := callback(microserviceId)
		return func(c *gin.Context) {
			// Inietta il categoryID nel context se specificato
			if categoryID != "" {
				c.Set("categoryID", categoryID)
			}
			originalHandler(c)
		}
	}

	// Restituisci una nuova action con le opzioni aggiornate
	return &hybridActionWrapper{
		options:         newOptions,
		callbackCreator: newCallback,
	}
}

// getCombinedSchemaForCategory combina lo schema base con quello della categoria
func (h EndorHybridService) getCombinedSchemaForCategory(category Category, getBaseSchema func() Schema) Schema {
	baseSchema := getBaseSchema()

	// Ottieni lo schema della categoria
	categorySchema, err := category.UnmarshalAdditionalAttributes()
	if err != nil {
		return baseSchema // Ritorna solo lo schema base se c'è un errore
	}

	return h.combineSchemas(baseSchema, categorySchema.Schema)
}

// combineSchemas combina due schemi
func (h EndorHybridService) combineSchemas(baseSchema, categorySchema Schema) Schema {
	combined := Schema{
		Type:       ObjectType,
		Properties: &map[string]Schema{},
	}

	// Aggiungi proprietà dello schema base
	if baseSchema.Properties != nil {
		for k, v := range *baseSchema.Properties {
			(*combined.Properties)[k] = v
		}
	}

	// Aggiungi proprietà dello schema categoria (sovrascrivono se esistenti)
	if categorySchema.Properties != nil {
		for k, v := range *categorySchema.Properties {
			(*combined.Properties)[k] = v
		}
	}

	return combined
}

// updateInputSchemaForCategory aggiorna l'InputSchema di un'azione per riflettere lo schema della categoria
func (h EndorHybridService) updateInputSchemaForCategory(originalSchema *RootSchema, getSchemaForCategory func() Schema) *RootSchema {
	if originalSchema == nil {
		return nil
	}

	// Crea una copia dello schema originale
	newSchema := &RootSchema{
		Schema: Schema{
			Type:       originalSchema.Schema.Type,
			Properties: &map[string]Schema{},
		},
	}

	// Copia tutte le proprietà originali
	if originalSchema.Schema.Properties != nil {
		for k, v := range *originalSchema.Schema.Properties {
			(*newSchema.Schema.Properties)[k] = v
		}
	}

	// Se c'è una proprietà "data", aggiornala con lo schema della categoria
	if _, exists := (*newSchema.Schema.Properties)["data"]; exists {
		// Crea il rootSchema con il modello base + categoria
		updatedRootSchema := h.getRootSchemaWithCategory(getSchemaForCategory)

		// Aggiorna la proprietà "data" con il nuovo schema
		(*newSchema.Schema.Properties)["data"] = updatedRootSchema.Schema
	}

	return newSchema
}

// getRootSchemaWithCategory crea un rootSchema che include sia il base model che lo schema della categoria
func (h EndorHybridService) getRootSchemaWithCategory(getSchemaForCategory func() Schema) *RootSchema {
	// Inizia con il base model
	rootSchema := NewSchema(h.baseModel)

	// Aggiungi le proprietà dello schema della categoria
	categorySchema := getSchemaForCategory()
	if categorySchema.Properties != nil {
		for k, v := range *categorySchema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}

	return rootSchema
}

// hybridActionWrapper implementa EndorServiceAction per le azioni esplose
type hybridActionWrapper struct {
	options         EndorServiceActionOptions
	callbackCreator func(microserviceId string) func(c *gin.Context)
}

func (h *hybridActionWrapper) CreateHTTPCallback(microserviceId string) func(c *gin.Context) {
	return h.callbackCreator(microserviceId)
}

func (h *hybridActionWrapper) GetOptions() EndorServiceActionOptions {
	return h.options
}
