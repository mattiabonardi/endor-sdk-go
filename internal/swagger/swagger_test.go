package swagger_test

import (
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	test_utils_handlers "github.com/mattiabonardi/endor-sdk-go/internal/test_utils/handlers"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSwaggerDefinition(t *testing.T) {
	microServiceId := "endor-sdk-service"
	module := "sdk"
	microServiceAddress := "http://localhost:8080"
	def, err := swagger.CreateSwaggerDefinition(microServiceId, module, microServiceAddress, []sdk.EndorHandler{test_utils_handlers.NewBaseHandlerHandler().ToEndorHandler()}, "/api")
	require.NoError(t, err, "Failed to create swagger definition")
	assert.Equal(t, "3.1.0", def.OpenAPI, "Expected OpenAPI version '3.1.0'")
	assert.Equal(t, microServiceId, def.Info.Title, "Expected correct title")
	assert.Equal(t, microServiceId+" docs", def.Info.Description, "Expected correct description")
	assert.Equal(t, "/", def.Servers[0].URL, "Expected correct server URL")
	// endor entities
	assert.Equal(t, "Base Handler (EndorBaseHandler)", def.Tags[0].Description, "Expected correct tag description")
	// check paths
	assert.Len(t, def.Paths, 2, "Expected 2 paths")
	assert.Contains(t, def.Paths, "/api/v1/sdk/base-handler/action1", "Expected '/api/v1/sdk/base-handler/action1' path to exist")
	assert.Contains(t, def.Paths, "/api/v1/sdk/base-handler/public-action", "Expected '/api/v1/sdk/base-handler/public-action' path to exist")
}
