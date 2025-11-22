package documentation

import (
	"fmt"
)

func ShowCategoriesDocumentation() {
	fmt.Println("🎉 Sistema di Categorizzazione EndorHybridService implementato con successo!")
	fmt.Println()

	fmt.Println("📋 Funzionalità implementate:")
	fmt.Println("✅ 1. Esteso il modello Resource con campo Categories")
	fmt.Println("✅ 2. Creata struct Category con attributi dinamici")
	fmt.Println("✅ 3. Aggiunto CategoryID al EndorContext")
	fmt.Println("✅ 4. Implementato getCategorySchema in EndorHybridService")
	fmt.Println("✅ 5. Routing con pattern resource__categoryID")
	fmt.Println("✅ 6. Aggiornato EndorServiceRepository per categorie")
	fmt.Println("✅ 7. Migliorato metodo ToEndorService")
	fmt.Println("✅ 8. Aggiornato service2.go con test di categorizzazione")
	fmt.Println()

	fmt.Println("🔧 Esempi di utilizzo:")
	fmt.Println()

	fmt.Println("📝 1. Definizione di una risorsa con categorie (MongoDB):")
	fmt.Println(`{
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
}`)
	fmt.Println()

	fmt.Println("🌐 2. API disponibili:")
	fmt.Println("   • /api/v1/product/schema          - Schema base del prodotto")
	fmt.Println("   • /api/v1/product__SUR/schema     - Schema per prodotti surgelati")
	fmt.Println("   • /api/v1/product__FRU/schema     - Schema per prodotti frutta")
	fmt.Println("   • /api/v1/product/list            - Lista generale prodotti")
	fmt.Println("   • /api/v1/product__SUR/list       - Lista prodotti surgelati")
	fmt.Println("   • /api/v1/product__FRU/list       - Lista prodotti frutta")
	fmt.Println()

	fmt.Println("💻 3. Esempio di EndorHybridService con categorie:")
	fmt.Println(`return sdk.NewHybridService("product", "Gestione Prodotti").
	WithActions(func(getSchema func() sdk.Schema, getCategorySchema func(categoryID string) sdk.Schema) map[string]sdk.EndorServiceAction {
		return map[string]sdk.EndorServiceAction{
			"validate": sdk.NewAction(
				func(c *sdk.EndorContext[ValidateDTO]) (*sdk.Response[any], error) {
					message := "Validazione prodotto"
					if c.CategoryID != nil {
						// Accesso allo schema della categoria specifica
						categorySchema := getCategorySchema(*c.CategoryID)
						message += " per categoria: " + *c.CategoryID
						// Logica di validazione specifica per categoria...
					}
					return sdk.NewResponseBuilder[any]().
						AddMessage(sdk.NewMessage(sdk.Info, message)).
						Build(), nil
				},
			),
		}
	})`)
	fmt.Println()

	fmt.Println("🚀 Sistema pronto per l'uso!")
	fmt.Println("   Il framework ora supporta pienamente la specializzazione")
	fmt.Println("   tramite categorizzazione con attributi dinamici e API dedicate.")
}
