package sdk

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type EndorHybridService struct {
	Resource       string
	Description    string
	Priority       *int
	metadataSchema Schema
	methodsFn      func() map[string]EndorServiceAction // Metodi wrapper che gestiscono tutte le categorie
	baseModel      ResourceInstanceInterface            // Modello base per ResourceInstanceRepository
	categories     map[string]Category                  // Cache delle categorie disponibili
	// Nuovi campi per supportare categorie specializzate
	specializedCategories map[string]interface{} // Cache delle categorie specializzate (tipo-indipendente)
}

func NewHybridService(resource, description string) EndorHybridService {
	return EndorHybridService{
		Resource:              resource,
		Description:           description,
		baseModel:             &DynamicResource{}, // Default model
		specializedCategories: make(map[string]interface{}),
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

// SpecializedCategoryInfo contiene informazioni su una categoria specializzata
type SpecializedCategoryInfo struct {
	ID                   string
	Description          string
	StaticModelSchema    *RootSchema // Schema del modello statico
	AdditionalAttributes string      // YAML degli attributi aggiuntivi
}

// WithSpecializedCategoryInfo aggiunge una categoria specializzata usando type-erased info
func (h EndorHybridService) WithSpecializedCategoryInfo(info SpecializedCategoryInfo) EndorHybridService {
	if h.specializedCategories == nil {
		h.specializedCategories = make(map[string]interface{})
	}

	// Aggiungi anche alla cache normale per backward compatibility
	if h.categories == nil {
		h.categories = make(map[string]Category)
	}
	h.categories[info.ID] = Category{
		ID:                   info.ID,
		Description:          info.Description,
		AdditionalAttributes: info.AdditionalAttributes,
	}

	// Aggiungi alla cache specializzata
	h.specializedCategories[info.ID] = info

	return h
}

// GetSpecializedCategoryInfo recupera informazioni su una categoria specializzata
func (h EndorHybridService) GetSpecializedCategoryInfo(categoryID string) (*SpecializedCategoryInfo, bool) {
	if category, exists := h.specializedCategories[categoryID]; exists {
		if typedCategory, ok := category.(SpecializedCategoryInfo); ok {
			return &typedCategory, true
		}
	}
	return nil, false
}

// Conversione in EndorService, schema iniettato dal framework
func (h EndorHybridService) ToEndorService(attrs Schema) EndorService {
	h.metadataSchema = attrs
	getSchema := func() Schema { return h.metadataSchema }
	var methods = make(map[string]EndorServiceAction)

	// Se non ci sono categorie, usa i metodi di default normali
	if len(h.categories) == 0 {
		rootSchema := h.getRootSchema(getSchema)
		methods = getDefaultActions(h.Resource, *rootSchema, h.Description)

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
		rootSchema := h.getRootSchema(getSchema)
		defaultMethods := getDefaultActions(h.Resource, *rootSchema, h.Description)
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
			rootSchema := h.getRootSchema(combinedSchemaFunc)
			categoryDefaultMethods := getDefaultActionsForCategory(h.Resource, *rootSchema, h.Description, currentCategoryID)

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

	// Prima prova a vedere se è una categoria specializzata
	if specializedInfo, exists := h.GetSpecializedCategoryInfo(category.ID); exists {
		return h.getCombinedSchemaForSpecializedCategory(*specializedInfo, getBaseSchema)
	}

	// Altrimenti usa il metodo tradizionale
	categorySchema, err := category.UnmarshalAdditionalAttributes()
	if err != nil {
		return baseSchema // Ritorna solo lo schema base se c'è un errore
	}

	return h.combineSchemas(baseSchema, categorySchema.Schema)
}

// getCombinedSchemaForSpecializedCategory combina schema per categoria specializzata
func (h EndorHybridService) getCombinedSchemaForSpecializedCategory(info SpecializedCategoryInfo, getBaseSchema func() Schema) Schema {
	baseSchema := getBaseSchema()

	// Inizia con lo schema base
	combined := baseSchema

	// Aggiungi le proprietà del modello statico se presente
	if info.StaticModelSchema != nil && info.StaticModelSchema.Schema.Properties != nil {
		combined = h.combineSchemas(combined, info.StaticModelSchema.Schema)
	}

	// Aggiungi gli attributi aggiuntivi se presenti
	if info.AdditionalAttributes != "" {
		var additionalSchema RootSchema
		err := yaml.Unmarshal([]byte(info.AdditionalAttributes), &additionalSchema)
		if err == nil && additionalSchema.Schema.Properties != nil {
			combined = h.combineSchemas(combined, additionalSchema.Schema)
		}
	}

	return combined
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
