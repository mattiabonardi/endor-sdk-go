package services_test

import (
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

// Modello base per il service specializzato
type Service2SpecializedBaseModel struct {
	ID        string `json:"id" bson:"_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

func (h Service2SpecializedBaseModel) GetID() *string {
	return &h.ID
}

func (h *Service2SpecializedBaseModel) SetID(id string) {
	h.ID = id
}

// Modello statico per Category 1
type Category1StaticModel struct {
	VATNumber    string `json:"vatNumber"`
	CompanySize  string `json:"companySize"`
	BusinessType string `json:"businessType"`
}

// Modello statico per Category 2
type Category2StaticModel struct {
	TaxID       string `json:"taxId"`
	FiscalYear  string `json:"fiscalYear"`
	AccountType string `json:"accountType"`
}

type Service2SpecializedAction1Payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Service2Specialized struct{}

func (h *Service2Specialized) action1(c *sdk.EndorContext[Service2SpecializedAction1Payload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().
		AddMessage(sdk.NewMessage(sdk.Info, "Hello from Specialized Hybrid Service")).
		Build(), nil
}

// NewService2Specialized crea un servizio con categorie specializzate
func NewService2Specialized() sdk.EndorHybridService {
	service := Service2Specialized{}

	// Schema aggiuntivo per Category 1 (YAML)
	category1AdditionalSchema := `
schema:
  type: object
  properties:
    additionalNote:
      type: string
      title: Additional Note
    priority:
      type: integer
      title: Priority Level
`

	// Schema aggiuntivo per Category 2 (YAML)
	category2AdditionalSchema := `
schema:
  type: object
  properties:
    region:
      type: string
      title: Geographic Region
    currency:
      type: string
      title: Default Currency
`

	// Crea le categorie specializzate info
	cat1Info := sdk.SpecializedCategoryInfo{
		ID:                   "cat-1",
		Description:          "Company Category (Specialized)",
		StaticModelSchema:    sdk.NewSchema(Category1StaticModel{}),
		AdditionalAttributes: category1AdditionalSchema,
	}

	cat2Info := sdk.SpecializedCategoryInfo{
		ID:                   "cat-2",
		Description:          "Tax Entity Category (Specialized)",
		StaticModelSchema:    sdk.NewSchema(Category2StaticModel{}),
		AdditionalAttributes: category2AdditionalSchema,
	}

	return sdk.NewHybridService("resource-2-specialized", "Resource 2 Specialized (EndorHybridService with static and dynamic category models)").
		WithBaseModel(&Service2SpecializedBaseModel{}).
		WithSpecializedCategoryInfo(cat1Info).
		WithSpecializedCategoryInfo(cat2Info).
		WithActions(func() map[string]sdk.EndorServiceAction {
			return map[string]sdk.EndorServiceAction{
				"specialized-action": sdk.NewAction(
					service.action1,
					"Test specialized hybrid action",
				),
			}
		})
}
