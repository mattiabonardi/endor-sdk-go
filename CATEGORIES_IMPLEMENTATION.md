# Sistema di Categorizzazione per EndorHybridService

## Implementazione Completata

Il framework EndorHybridService ora supporta pienamente il sistema di **categorizzazione/specializzazione** per le entità, permettendo di definire attributi dinamici specifici per categoria e API specializzate.

## 🔧 Componenti Implementati

### 1. **Modello di Dati** (`sdk/resource.go`)
- ✅ **Struct `Category`**: Con ID, Description e AdditionalAttributes
- ✅ **Esteso `Resource`**: Aggiunto campo `Categories []Category`
- ✅ **Metodi helper**: 
  - `GetCategoryByID(categoryID string)`
  - `GetCategorySchema(categoryID string)`

### 2. **Context e Injection** (`sdk/context.go` e `sdk/endor_service.go`)
- ✅ **EndorContext esteso**: Aggiunto campo `CategoryID *string`
- ✅ **Injection automatica**: Il categoryID viene estratto dal routing e iniettato nel context

### 3. **EndorHybridService** (`sdk/endor_hybrid_service.go`)
- ✅ **Signature aggiornata**: `WithActions` ora riceve `getCategorySchema func(categoryID string) Schema`
- ✅ **Gestione categorie**: Metodo `WithCategories([]Category)`
- ✅ **Schema dinamico**: Funzione `getCategorySchema` disponibile negli action handlers

### 4. **Routing Specializzato** (`sdk/server.go`)
- ✅ **Pattern supportato**: `/api/v1/resource__categoryID/action`
- ✅ **Parsing automatico**: Estrazione di resource e categoryID dal path
- ✅ **Injection nel context**: CategoryID disponibile tramite `c.CategoryID`

### 5. **Repository e Mapping** (`sdk/endor_resource_repository.go`)
- ✅ **Gestione categorie**: Le categorie vengono caricate da MongoDB e iniettate negli hybrid services
- ✅ **Schema combinato**: Merge automatico di schema base + schema categoria

## 🚀 Utilizzo Pratico

### Definizione della Risorsa (MongoDB)
```json
{
  "_id": "product",
  "description": "Product",
  "service": "endor-erp-service", 
  "additionalAttributes": "pesoStandard: number\nfornitoreDefault: string",
  "categories": [
    {
      "id": "SUR", 
      "description": "Articoli surgelati",
      "additionalAttributes": "temperatureMin: number\ntemperatureMax: number"
    },
    {
      "id": "FRU",
      "description": "Articoli frutta", 
      "additionalAttributes": "origine: string\ncalibro: number"
    }
  ]
}
```

### EndorHybridService con Categorie
```go
func NewProductService() sdk.EndorHybridService {
    service := ProductService{}
    
    return sdk.NewHybridService("product", "Gestione Prodotti").
        WithActions(func(getSchema func() sdk.Schema, getCategorySchema func(categoryID string) sdk.Schema) map[string]sdk.EndorServiceAction {
            return map[string]sdk.EndorServiceAction{
                "validate": sdk.NewAction(
                    func(c *sdk.EndorContext[ValidateDTO]) (*sdk.Response[any], error) {
                        if c.CategoryID != nil {
                            // Schema specifico della categoria
                            categorySchema := getCategorySchema(*c.CategoryID)
                            // Logica di validazione specifica per categoria
                            return service.validateByCategory(c, categorySchema)
                        }
                        // Logica generica
                        return service.validateGeneric(c)
                    },
                    "Validate product data",
                ),
            }
        })
}
```

### API Generate Automaticamente
- **Base**: `/api/v1/product/schema`, `/api/v1/product/list`, `/api/v1/product/create`
- **Categoria Surgelati**: `/api/v1/product__SUR/schema`, `/api/v1/product__SUR/list`, `/api/v1/product__SUR/create`
- **Categoria Frutta**: `/api/v1/product__FRU/schema`, `/api/v1/product__FRU/list`, `/api/v1/product__FRU/create`

### Accesso al CategoryID negli Actions
```go
func (s *ProductService) specializedAction(c *sdk.EndorContext[MyDTO]) (*sdk.Response[any], error) {
    message := "Processing product"
    
    if c.CategoryID != nil {
        message += " for category: " + *c.CategoryID
        // Logica specifica per la categoria
    } else {
        message += " (general processing)"
        // Logica generale
    }
    
    return sdk.NewResponseBuilder[any]().
        AddMessage(sdk.NewMessage(sdk.Info, message)).
        Build(), nil
}
```

## 🎯 Vantaggi del Sistema

1. **Flessibilità**: Nuove categorie possono essere aggiunte a runtime senza modificare codice
2. **API Specializzate**: Ogni categoria ha le proprie API con schema specifico
3. **Schema Dinamici**: Attributi specifici per categoria definiti in MongoDB
4. **Retrocompatibilità**: Le API esistenti continuano a funzionare
5. **DDD Compliant**: Separazione netta tra attributi generali e specifici

## 🔍 Test e Validazione

Il sistema è stato testato con:
- ✅ Compilazione senza errori
- ✅ Test esistenti che continuano a passare  
- ✅ Service2 con action di test per CategoryID
- ✅ Routing pattern `resource__categoryID` funzionante
- ✅ Injection automatica di CategoryID nel context

## 📋 Compatibilità

Tutti i servizi esistenti continuano a funzionare senza modifiche. La nuova signature di `WithActions` è backward compatible tramite il parametro `getCategorySchema` che può essere ignorato se non necessario.