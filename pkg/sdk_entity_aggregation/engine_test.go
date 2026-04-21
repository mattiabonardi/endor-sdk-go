package sdk_entity_aggregation

import (
	"context"
	"fmt"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// orderDocs is the shared test dataset used by aggregation tests.
var orderDocs = []map[string]interface{}{
	{"id": "1", "customerId": "c1", "status": "completed", "amount": float64(100)},
	{"id": "2", "customerId": "c1", "status": "completed", "amount": float64(200)},
	{"id": "3", "customerId": "c2", "status": "completed", "amount": float64(150)},
	{"id": "4", "customerId": "c2", "status": "pending", "amount": float64(50)},
	{"id": "5", "customerId": "c3", "status": "completed", "amount": float64(300)},
}

// entityResults built from the shared orderDocs dataset plus extra customer docs.
var customerDocs = []map[string]interface{}{
	{"id": "c1", "name": "Alice", "country": "IT"},
	{"id": "c2", "name": "Bob", "country": "US"},
	{"id": "c3", "name": "Carol", "country": "FR"},
}

// #region selection

func TestGroupBy_ByCustomer(t *testing.T) {
	cleanup := registerMock(newMockRepository("order", orderDocs))
	defer cleanup()

	p := AggregationPipeline{
		{
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{"id": "$customerId"}},
			},
		},
	}

	result, _, _, err := NewAggregationEngine(testDI).Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(result))
	}

	byID := indexByID(result)

	if _, ok := byID["c1"]; !ok {
		t.Errorf("expected group c1")
	}
	if _, ok := byID["c2"]; !ok {
		t.Errorf("expected group c2")
	}
	if _, ok := byID["c3"]; !ok {
		t.Errorf("expected group c3")
	}
}

// #endregion

// #region accumulation

func TestGroupBy_ByCustomer_WithSum(t *testing.T) {
	customerEntity := "customer"
	orderSchema := &sdk.RootSchema{
		Schema: sdk.Schema{
			Type: sdk.SchemaTypeObject,
			Properties: &map[string]sdk.Schema{
				"customerId": {
					Type: sdk.SchemaTypeString,
					UISchema: &sdk.UISchema{
						Entity: &customerEntity,
					},
				},
				"amount": {Type: sdk.SchemaTypeNumber},
				"status": {Type: sdk.SchemaTypeString},
			},
		},
	}
	orderRepo := newMockRepository("order", orderDocs)
	orderRepo.schema = orderSchema
	cleanup := registerMock(orderRepo)
	defer cleanup()

	p := AggregationPipeline{
		{
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{
					"id":    "$customerId",
					"total": map[string]interface{}{"$sum": "$amount"},
				}},
			},
		},
	}

	result, schema, _, err := NewAggregationEngine(testDI).Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c2"]["total"].(float64); got != 200 {
		t.Errorf("c2 total: got %v, want 200", got)
	}
	if got := byID["c3"]["total"].(float64); got != 300 {
		t.Errorf("c3 total: got %v, want 300", got)
	}

	// Schema checks: "id" must inherit UISchema.Entity from "customerId";
	// "total" must be a number (produced by $sum).
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if schema.Properties == nil {
		t.Fatal("expected schema.Properties to be non-nil")
	}
	idProp, ok := (*schema.Properties)["id"]
	if !ok {
		t.Fatal("expected schema to contain \"id\"")
	}
	if idProp.UISchema == nil || idProp.UISchema.Entity == nil || *idProp.UISchema.Entity != customerEntity {
		t.Errorf("expected id.UISchema.Entity = %q, got %v", customerEntity, idProp.UISchema)
	}
	totalProp, ok := (*schema.Properties)["total"]
	if !ok {
		t.Fatal("expected schema to contain \"total\"")
	}
	if totalProp.Type != sdk.SchemaTypeNumber {
		t.Errorf("expected total.Type = %q, got %q", sdk.SchemaTypeNumber, totalProp.Type)
	}
}

// #endregion

// #region combination

func TestMergeResults(t *testing.T) {
	customerEntity := "customer"
	orderSchema := &sdk.RootSchema{
		Schema: sdk.Schema{
			Type: sdk.SchemaTypeObject,
			Properties: &map[string]sdk.Schema{
				"customerId": {
					Type:     sdk.SchemaTypeString,
					UISchema: &sdk.UISchema{Entity: &customerEntity},
				},
				"amount": {Type: sdk.SchemaTypeNumber},
				"status": {Type: sdk.SchemaTypeString},
			},
		},
	}
	customerSchema := &sdk.RootSchema{
		Schema: sdk.Schema{
			Type: sdk.SchemaTypeObject,
			Properties: &map[string]sdk.Schema{
				"id":      {Type: sdk.SchemaTypeString},
				"name":    {Type: sdk.SchemaTypeString},
				"country": {Type: sdk.SchemaTypeString},
			},
		},
	}

	orderRepo := newMockRepository("order", orderDocs)
	orderRepo.schema = orderSchema
	customerRepo := newMockRepository("customer", customerDocs)
	customerRepo.schema = customerSchema
	customerRepo.refDescs = sdk.EntityReferenceGroupDescriptions{
		"c1": "Alice",
		"c2": "Bob",
		"c3": "Carol",
	}
	cleanupOrders := registerMock(orderRepo)
	defer cleanupOrders()
	cleanupCustomers := registerMock(customerRepo)
	defer cleanupCustomers()

	// Group orders by customerId → each doc gets "id" = customerId.
	// Then merge with customer docs (which also carry "id") to get a combined view.
	p := AggregationPipeline{
		{
			ID:     "grouped_orders",
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{
					"id":    "$customerId",
					"total": map[string]interface{}{"$sum": "$amount"},
				}},
			},
		},
		{
			ID:       "customers",
			Entity:   "customer",
			Pipeline: []StageSpec{},
		},
		{
			DependsOn: []string{"grouped_orders", "customers"},
			Pipeline: []StageSpec{
				{"$mergeResults": map[string]interface{}{"on": "id"}},
			},
		},
	}

	result, schema, refs, err := NewAggregationEngine(testDI).Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 merged docs, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c1"]["name"].(string); got != "Alice" {
		t.Errorf("c1 name: got %v, want Alice", got)
	}
	if got := byID["c2"]["total"].(float64); got != 200 {
		t.Errorf("c2 total: got %v, want 200", got)
	}
	if got := byID["c2"]["country"].(string); got != "US" {
		t.Errorf("c2 country: got %v, want US", got)
	}
	if got := byID["c3"]["name"].(string); got != "Carol" {
		t.Errorf("c3 name: got %v, want Carol", got)
	}

	// Schema checks: merged schema must contain fields from both stages.
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if schema.Properties == nil {
		t.Fatal("expected schema.Properties to be non-nil")
	}
	// "id" comes from grouped_orders and inherits UISchema.Entity="customer".
	idProp, ok := (*schema.Properties)["id"]
	if !ok {
		t.Fatal("expected merged schema to contain \"id\"")
	}
	if idProp.UISchema == nil || idProp.UISchema.Entity == nil || *idProp.UISchema.Entity != customerEntity {
		t.Errorf("expected id.UISchema.Entity = %q, got %v", customerEntity, idProp.UISchema)
	}
	// "total" comes from grouped_orders ($sum → number).
	totalProp, ok := (*schema.Properties)["total"]
	if !ok {
		t.Fatal("expected merged schema to contain \"total\"")
	}
	if totalProp.Type != sdk.SchemaTypeNumber {
		t.Errorf("expected total.Type = %q, got %q", sdk.SchemaTypeNumber, totalProp.Type)
	}
	// "name" and "country" come from customers schema.
	if _, ok := (*schema.Properties)["name"]; !ok {
		t.Error("expected merged schema to contain \"name\"")
	}
	if _, ok := (*schema.Properties)["country"]; !ok {
		t.Error("expected merged schema to contain \"country\"")
	}

	// References checks: the order stage resolves "id" → "customer" references.
	if refs == nil {
		t.Fatal("expected non-nil references")
	}
	customerRefs, ok := refs["customer"]
	if !ok {
		t.Fatal("expected references[\"customer\"] to be present")
	}
	if got := customerRefs["c1"]; got != "Alice" {
		t.Errorf("references[customer][c1]: got %q, want \"Alice\"", got)
	}
	if got := customerRefs["c2"]; got != "Bob" {
		t.Errorf("references[customer][c2]: got %q, want \"Bob\"", got)
	}
	if got := customerRefs["c3"]; got != "Carol" {
		t.Errorf("references[customer][c3]: got %q, want \"Carol\"", got)
	}
}

// #endregion

// #region entity_stage_handler

// TestEntityStageHandler_ReplacesBuiltinLogic verifies that, when an
// EntityStageHandler is provided via WithEntityStageHandler, the engine calls
// the callback for every entity stage and uses its returned docs instead of
// hitting the repository registry. The callback here computes a result from the
// stage metadata so the test is self-contained and does not need a mock repo.
func TestEntityStageHandler_ReplacesBuiltinLogic(t *testing.T) {
	// computedByEntity simulates what a master query layer would do: it returns
	// a synthetic document set whose content depends on the entity name.
	computedByEntity := map[string][]map[string]interface{}{
		"order": {
			{"id": "c1", "total": float64(300)},
			{"id": "c2", "total": float64(200)},
			{"id": "c3", "total": float64(300)},
		},
		"customer": {
			{"id": "c1", "name": "Alice"},
			{"id": "c2", "name": "Bob"},
			{"id": "c3", "name": "Carol"},
		},
	}

	handler := func(_ context.Context, stage EntityPipelineStage) ([]map[string]interface{}, *sdk.Schema, sdk.EntityRefererenceGroup, error) {
		docs, ok := computedByEntity[stage.Entity]
		if !ok {
			return []map[string]interface{}{}, nil, nil, nil
		}
		return docs, nil, nil, nil
	}

	p := AggregationPipeline{
		{ID: "orders", Entity: "order", Pipeline: []StageSpec{}},
		{ID: "customers", Entity: "customer", Pipeline: []StageSpec{}},
		{
			DependsOn: []string{"orders", "customers"},
			Pipeline:  []StageSpec{{"$mergeResults": map[string]interface{}{"on": "id"}}},
		},
	}

	engine := NewAggregationEngine(testDI, WithEntityStageHandler(handler))
	result, _, _, err := engine.Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 merged docs, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c1"]["name"].(string); got != "Alice" {
		t.Errorf("c1 name: got %v, want Alice", got)
	}
	if got := byID["c2"]["total"].(float64); got != 200 {
		t.Errorf("c2 total: got %v, want 200", got)
	}
	if got := byID["c2"]["name"].(string); got != "Bob" {
		t.Errorf("c2 name: got %v, want Bob", got)
	}
	if got := byID["c3"]["name"].(string); got != "Carol" {
		t.Errorf("c3 name: got %v, want Carol", got)
	}
}

// TestEntityStageHandler_OwnsFullStage verifies that the EntityStageHandler
// receives the complete EntityPipelineStage — including its Pipeline — and that
// its return value is used as-is, without the engine re-applying any in-memory
// operators. This models a master microservice that forwards the entire stage
// (entity + pipeline) to a child microservice via HTTP and returns its result.
func TestEntityStageHandler_OwnsFullStage(t *testing.T) {
	// The handler simulates a child microservice that has already executed the
	// $group+$sum locally and returns the aggregated result directly.
	handler := func(_ context.Context, stage EntityPipelineStage) ([]map[string]interface{}, *sdk.Schema, sdk.EntityRefererenceGroup, error) {
		// Assert that the full pipeline is forwarded to the handler.
		if len(stage.Pipeline) != 1 {
			return nil, nil, nil, fmt.Errorf("expected 1 pipeline stage forwarded, got %d", len(stage.Pipeline))
		}
		// Return the already-aggregated result (as a child microservice would).
		return []map[string]interface{}{
			{"id": "c1", "total": float64(300)},
			{"id": "c2", "total": float64(150)},
		}, nil, nil, nil
	}

	p := AggregationPipeline{
		{
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{
					"id":    "$customerId",
					"total": map[string]interface{}{"$sum": "$amount"},
				}},
			},
		},
	}

	engine := NewAggregationEngine(testDI, WithEntityStageHandler(handler))
	result, _, _, err := engine.Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c2"]["total"].(float64); got != 150 {
		t.Errorf("c2 total: got %v, want 150", got)
	}
}

// #endregion

// #region references

// TestExecute_References verifies that, after a $group stage that maps
// "$productId" → "id", the engine resolves the entity references for the
// output IDs using the "product" repository registered in the global registry.
func TestExecute_References(t *testing.T) {
	productEntity := "product"
	stockSchema := &sdk.RootSchema{
		Schema: sdk.Schema{
			Type: sdk.SchemaTypeObject,
			Properties: &map[string]sdk.Schema{
				"productId": {
					Type: sdk.SchemaTypeString,
					UISchema: &sdk.UISchema{
						Entity: &productEntity,
					},
				},
				"quantity": {Type: sdk.SchemaTypeNumber},
			},
		},
	}

	stockDocs := []map[string]interface{}{
		{"productId": "p1", "quantity": float64(10)},
		{"productId": "p1", "quantity": float64(5)},
		{"productId": "p2", "quantity": float64(3)},
	}

	stockRepo := newMockRepository("stock", stockDocs)
	stockRepo.schema = stockSchema

	productRepo := newMockRepository("product", nil)
	productRepo.refDescs = sdk.EntityReferenceGroupDescriptions{
		"p1": "Widget A",
		"p2": "Widget B",
	}

	cleanupStock := registerMock(stockRepo)
	defer cleanupStock()
	cleanupProduct := registerMock(productRepo)
	defer cleanupProduct()

	p := AggregationPipeline{
		{
			Entity: "stock",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{
					"id":            "$productId",
					"totalQuantity": map[string]interface{}{"$sum": "$quantity"},
				}},
			},
		},
	}

	result, schema, refs, err := NewAggregationEngine(testDI).Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result))
	}

	// Verify schema: id should carry UISchema.Entity="product", totalQuantity should be number.
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if schema.Properties == nil {
		t.Fatal("expected schema.Properties to be non-nil")
	}
	idProp, ok := (*schema.Properties)["id"]
	if !ok {
		t.Fatal("expected schema to have \"id\" property")
	}
	if idProp.UISchema == nil || idProp.UISchema.Entity == nil || *idProp.UISchema.Entity != productEntity {
		t.Errorf("expected id.UISchema.Entity = %q, got %v", productEntity, idProp.UISchema)
	}
	totalProp, ok := (*schema.Properties)["totalQuantity"]
	if !ok {
		t.Fatal("expected schema to have \"totalQuantity\" property")
	}
	if totalProp.Type != sdk.SchemaTypeNumber {
		t.Errorf("expected totalQuantity type = %q, got %q", sdk.SchemaTypeNumber, totalProp.Type)
	}

	if refs == nil {
		t.Fatal("expected non-nil references")
	}
	productRefs, ok := refs["product"]
	if !ok {
		t.Fatal("expected references[\"product\"] to be present")
	}
	if got := productRefs["p1"]; got != "Widget A" {
		t.Errorf("references[product][p1]: got %q, want %q", got, "Widget A")
	}
	if got := productRefs["p2"]; got != "Widget B" {
		t.Errorf("references[product][p2]: got %q, want %q", got, "Widget B")
	}
}

// #endregion

// indexByID indexes grouped results by the "id" field produced by $group.
func indexByID(docs []map[string]interface{}) map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{}, len(docs))
	for _, doc := range docs {
		key := ""
		if id := doc["id"]; id != nil {
			key = id.(string)
		}
		m[key] = doc
	}
	return m
}

// #endregion
