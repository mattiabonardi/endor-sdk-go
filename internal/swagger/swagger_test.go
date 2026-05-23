package swagger_test

import (
	"os"
	"testing"

	examples_handlers "github.com/mattiabonardi/endor-sdk-go/internal/examples/handlers"
	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSwaggerDefinition(t *testing.T) {
	microServiceId := "endor-sdk-service"
	module := "sdk"
	microServiceAddress := "http://localhost:8080"
	// Go test runs with cwd = package directory; ../../locales is the project locales dir.
	localesFS := os.DirFS("../../locales")
	def, err := swagger.CreateSwaggerDefinition(microServiceId, module, microServiceAddress, []sdk.EndorHandler{examples_handlers.NewBaseEntityHandler().ToEndorHandler()}, "/api", localesFS)
	require.NoError(t, err, "Failed to create swagger definition")
	assert.Equal(t, "3.1.0", def.OpenAPI, "Expected OpenAPI version '3.1.0'")
	assert.Equal(t, microServiceId, def.Info.Title, "Expected correct title")
	assert.Equal(t, microServiceId+" docs", def.Info.Description, "Expected correct description")
	assert.Equal(t, "/", def.Servers[0].URL, "Expected correct server URL")
	// endor entities
	assert.Equal(t, "Base entity", def.Tags[0].Description, "Expected correct tag description")
	// check paths
	assert.Len(t, def.Paths, 2, "Expected 2 paths")
	assert.Contains(t, def.Paths, "/api/v1/sdk/base-entity/action1", "Expected '/api/v1/sdk/base-entity/action1' path to exist")
	assert.Contains(t, def.Paths, "/api/v1/sdk/base-entity/public-action", "Expected '/api/v1/sdk/base-entity/public-action' path to exist")
}
