package swagger_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	test_utils_service "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/service"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSwaggerDefinition(t *testing.T) {
	def, err := swagger.CreateSwaggerDefinition("endor-sdk-service", "endorsdkservice.com", []sdk.EndorService{test_utils_service.NewService1()}, "/api")
	require.NoError(t, err, "Failed to create swagger definition")
	assert.Equal(t, "3.1.0", def.OpenAPI, "Expected OpenAPI version '3.1.0'")
	assert.Equal(t, "endor-sdk-service", def.Info.Title, "Expected correct title")
	assert.Equal(t, "endor-sdk-service docs", def.Info.Description, "Expected correct description")
	assert.Equal(t, "/", def.Servers[0].URL, "Expected correct server URL")
	// endor resources
	assert.Equal(t, "Resource 1 (EndorService with categories)", def.Tags[0].Description, "Expected correct tag description")
	// check paths
	assert.Len(t, def.Paths, 3, "Expected 3 paths")
	assert.Contains(t, def.Paths, "/api/v1/resource-1/action1", "Expected '/api/v1/resource-1/action1' path to exist")
	assert.Contains(t, def.Paths, "/api/v1/resource-1/cat_1/action1", "Expected '/api/v1/resource-1/cat_1/action1' path to exist")
	assert.Contains(t, def.Paths, "/api/v1/resource-1/public-action", "Expected '/api/v1/resource-1/public-action' path to exist")
}
