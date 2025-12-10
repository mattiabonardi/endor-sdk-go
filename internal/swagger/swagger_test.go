package swagger_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	test_utils_services "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/services"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSwaggerDefinition(t *testing.T) {
	def, err := swagger.CreateSwaggerDefinition("endor-sdk-service", "endorsdkservice.com", []sdk.EndorService{test_utils_services.NewBaseServiceService().ToEndorService()}, "/api")
	require.NoError(t, err, "Failed to create swagger definition")
	assert.Equal(t, "3.1.0", def.OpenAPI, "Expected OpenAPI version '3.1.0'")
	assert.Equal(t, "endor-sdk-service", def.Info.Title, "Expected correct title")
	assert.Equal(t, "endor-sdk-service docs", def.Info.Description, "Expected correct description")
	assert.Equal(t, "/", def.Servers[0].URL, "Expected correct server URL")
	// endor resources
	assert.Equal(t, "Base Service (EndorBaseService)", def.Tags[0].Description, "Expected correct tag description")
	// check paths
	assert.Len(t, def.Paths, 1, "Expected 1 paths")
	assert.Contains(t, def.Paths, "/api/v1/base-service/action-1", "Expected '/api/v1/base-service/action-1' path to exist")
}
