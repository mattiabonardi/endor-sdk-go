package sdk_entity_test

// core_test.go — unit tests for RegistryCore.Dictionary (prod and dev runtimes).
//
// Test fixtures (internal/examples/dsl/) are committed to the repo so any clone can run the suite.
// Layout:
//   internal/examples/dsl/prod/entities/sdk/
//     dynamic-entity.yaml              — pure DSL entity (no static handler)
//     dynamic-specialized-entity.yaml  — pure DSL entity with categories
//     hybrid-entity.yaml               — extends the "hybrid-entity" static handler
//     hybrid-specialized-entity.yaml  — extends the "hybrid-specialized-entity" static handler
//   internal/examples/dsl/dev/user1/entities/sdk/
//     dev-entity.yaml       — user1-only entity
//     hybrid-entity.yaml    — user1 extension of "hybrid-entity" (different field than prod)
//   internal/examples/dsl/dev/user2/entities/sdk/
//     user2-entity.yaml     — user2-only entity
//
// NOTE: Tests that involve DSL dynamic entities or hybrid DSL extensions will call
// collectAllRepositories, which lazily attempts a MongoDB connection on first invocation.
// If MongoDB is not running, the connection attempt times out (default 10 s, once per
// test binary). All assertions remain valid regardless — repos are created with a nil
// base but the dictionary structure is unaffected.

import (
	"path/filepath"
	"sync"
	"testing"

	examples_handlers "github.com/mattiabonardi/endor-sdk-go/internal/examples/handlers"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

const coreTestModule = "sdk"

// Paths to committed DSL fixtures (relative to the package directory).
const (
	testdataProdPath = "../../internal/examples/dsl/prod"
	testdataDevPath  = "../../internal/examples/dsl/dev"
)

// newTestRegistryCore builds a RegistryCore ready for unit tests.
//   - prodDSLPath: base path for the prod DSLDAO; uses an isolated temp dir when empty.
//   - devDSLBase:  base path for per-user dev DSLDAOs; uses an isolated temp dir when empty.
func newTestRegistryCore(t *testing.T, handlers []sdk.EndorHandlerInterface, prodDSLPath, devDSLBase string) *sdk_entity.RegistryCore {
	t.Helper()
	if prodDSLPath == "" {
		prodDSLPath = t.TempDir()
	}
	if devDSLBase == "" {
		devDSLBase = t.TempDir()
	}
	logger := sdk.NewLogger(sdk.LogConfig{LogType: sdk.StringLog}, sdk.LogContext{})
	return &sdk_entity.RegistryCore{
		Module:                coreTestModule,
		MicroServiceId:        "test-service",
		InternalEndorHandlers: &handlers,
		Logger:                logger,
		Mu:                    &sync.RWMutex{},
		ProdDAO:               &sdk.DSLDAO{BasePath: prodDSLPath},
		EphemeralCache:        sdk_entity.NewEphemeralCacheManager(),
		DevDAOFactory: func(username string) *sdk.DSLDAO {
			return &sdk.DSLDAO{BasePath: filepath.Join(devDSLBase, username)}
		},
	}
}

// ---------------------------------------------------------------------------
// Prod runtime — static handlers
// ---------------------------------------------------------------------------

// TestDictionary_Prod_StaticHandlersPresent verifies that all registered static
// handlers appear in the production dictionary with the correct entity type and key.
func TestDictionary_Prod_StaticHandlersPresent(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
		examples_handlers.NewBaseSpecializedEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", "")

	dict, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	assert.Len(t, dict, 2)
	assert.Contains(t, dict, "sdk/base-entity")
	assert.Contains(t, dict, "sdk/base-specialized-entity")
	assert.Equal(t, string(sdk.EntityTypeBase), dict["sdk/base-entity"].Entity.Type)
	assert.Equal(t, string(sdk.EntityTypeBaseSpecialized), dict["sdk/base-specialized-entity"].Entity.Type)
}

// TestDictionary_Prod_DSL_AddsNewDynamicEntity verifies that a pure-DSL dynamic entity
// defined in testdata/prod is added to the production dictionary.
func TestDictionary_Prod_DSL_AddsNewDynamicEntity(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, testdataProdPath, "")

	dict, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	// Static handler present
	assert.Contains(t, dict, "sdk/base-entity")
	// DSL-only dynamic entity added
	require.Contains(t, dict, "sdk/dynamic-entity")
	assert.Equal(t, string(sdk.EntityTypeDynamic), dict["sdk/dynamic-entity"].Entity.Type)
	assert.Equal(t, "Prod Dynamic Entity", dict["sdk/dynamic-entity"].Entity.Title)
}

// TestDictionary_Prod_DSL_ExtendsHybridHandlerSchema verifies that the prod DSL
// correctly merges additional properties into an existing static hybrid handler's schema.
func TestDictionary_Prod_DSL_ExtendsHybridHandlerSchema(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewHybridEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, testdataProdPath, "")

	dict, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	require.Contains(t, dict, "sdk/hybrid-entity")
	entry := dict["sdk/hybrid-entity"]

	require.NotNil(t, entry.EndorHandler.EntitySchema.Properties)
	props := *entry.EndorHandler.EntitySchema.Properties

	// Base model property
	assert.Contains(t, props, "id", "base model 'id' property should be present")
	// DSL-added property
	assert.Contains(t, props, "additionalProdField", "prod DSL property should be merged in")
}

// TestDictionary_Prod_DSL_AddsNewDynamicSpecializedEntity verifies that a pure-DSL entity
// with categories becomes a dynamic-specialized entry in the production dictionary.
func TestDictionary_Prod_DSL_AddsNewDynamicSpecializedEntity(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{}
	core := newTestRegistryCore(t, handlers, testdataProdPath, "")

	dict, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	require.Contains(t, dict, "sdk/dynamic-specialized-entity")
	entry := dict["sdk/dynamic-specialized-entity"]
	assert.Equal(t, string(sdk.EntityTypeDynamicSpecialized), entry.Entity.Type)
	assert.Equal(t, "Prod Dynamic Specialized Entity", entry.Entity.Title)

	// Both DSL-declared categories must be present.
	catIDs := make([]string, 0, len(entry.Entity.Categories))
	for _, cat := range entry.Entity.Categories {
		catIDs = append(catIDs, cat.ID)
	}
	assert.Contains(t, catIDs, "cat-a")
	assert.Contains(t, catIDs, "cat-b")
}

// TestDictionary_Prod_DSL_ExtendsHybridSpecializedHandlerWithCategory verifies that a
// prod DSL file can add extra categories to an existing hybrid-specialized static handler.
func TestDictionary_Prod_DSL_ExtendsHybridSpecializedHandlerWithCategory(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewHybridSpecializedEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, testdataProdPath, "")

	dict, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	require.Contains(t, dict, "sdk/hybrid-specialized-entity")
	entry := dict["sdk/hybrid-specialized-entity"]
	assert.Equal(t, string(sdk.EntityTypeHybridSpecialized), entry.Entity.Type)

	// Static handler already has cat-1 and cat-2; DSL adds dsl-cat.
	catIDs := make([]string, 0, len(entry.Entity.Categories))
	for _, cat := range entry.Entity.Categories {
		catIDs = append(catIDs, cat.ID)
	}
	assert.Contains(t, catIDs, "cat-1", "static category cat-1 should be present")
	assert.Contains(t, catIDs, "cat-2", "static category cat-2 should be present")
	assert.Contains(t, catIDs, "dsl-cat", "DSL-added category dsl-cat should be present")
}

// TestDictionary_Prod_IsCached verifies that successive calls to Dictionary return
// consistent results without rebuilding the dictionary each time.
func TestDictionary_Prod_IsCached(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", "")

	dict1, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	// Cache should be populated after the first call.
	core.Mu.RLock()
	initialized := core.CacheInitialized
	core.Mu.RUnlock()
	assert.True(t, initialized)

	dict2, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	// Both calls should return the same set of entity keys.
	assert.Equal(t, len(dict1), len(dict2))
	for k := range dict1 {
		assert.Contains(t, dict2, k)
	}
}

// TestDictionary_Prod_SyncResetsCache verifies that Sync() invalidates the production
// cache so the next Dictionary call triggers a full rebuild.
func TestDictionary_Prod_SyncResetsCache(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", "")

	_, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	core.Mu.RLock()
	assert.True(t, core.CacheInitialized, "cache should be populated before Sync")
	core.Mu.RUnlock()

	core.Sync()

	core.Mu.RLock()
	assert.False(t, core.CacheInitialized, "cache should be invalidated after Sync")
	assert.Nil(t, core.CachedDictionary)
	assert.Nil(t, core.CachedDIContainer)
	core.Mu.RUnlock()
}

// TestDictionary_Prod_DictionaryInstance_Found checks that DictionaryInstance resolves
// a valid entity ID to the correct dictionary entry.
func TestDictionary_Prod_DictionaryInstance_Found(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", "")

	entry, err := core.DictionaryInstance(sdk.Session{}, sdk.ReadInstanceDTO{Id: "sdk/base-entity"})
	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "sdk/base-entity", entry.Entity.ID)
}

// TestDictionary_Prod_DictionaryInstance_NotFound checks that DictionaryInstance returns
// a not-found error for an unknown entity ID.
func TestDictionary_Prod_DictionaryInstance_NotFound(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", "")

	_, err := core.DictionaryInstance(sdk.Session{}, sdk.ReadInstanceDTO{Id: "sdk/does-not-exist"})
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Dev runtime — isolation and overlay
// ---------------------------------------------------------------------------

// TestDictionary_Dev_StartsFromStaticNotProdDSL is the key regression test for the
// "dev starts from static entries only" fix.
// It verifies that a dynamic entity defined exclusively in the prod DSL is NOT visible
// in a dev session, proving that dev builds start from the static dictionary and do not
// inherit the prod DSL overlay.
func TestDictionary_Dev_StartsFromStaticNotProdDSL(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	// Prod DSL adds "dynamic-entity"; dev DSL (user1) adds "dev-entity".
	core := newTestRegistryCore(t, handlers, testdataProdPath, testdataDevPath)

	devSession := sdk.Session{Development: true, Username: "user1"}
	devDict, err := core.Dictionary(devSession)
	require.NoError(t, err)

	// Prod DSL entity must NOT be visible in dev.
	assert.NotContains(t, devDict, "sdk/dynamic-entity",
		"dev dictionary must not inherit prod DSL dynamic entities")
}

// TestDictionary_Dev_IncludesDevDSLEntities verifies that entities defined in the
// user's own dev DSL are present in their dev dictionary.
func TestDictionary_Dev_IncludesDevDSLEntities(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, testdataProdPath, testdataDevPath)

	devSession := sdk.Session{Development: true, Username: "user1"}
	devDict, err := core.Dictionary(devSession)
	require.NoError(t, err)

	// Static entity still present.
	assert.Contains(t, devDict, "sdk/base-entity")
	// User1 dev-only entity present.
	require.Contains(t, devDict, "sdk/dev-entity")
	assert.Equal(t, string(sdk.EntityTypeDynamic), devDict["sdk/dev-entity"].Entity.Type)
}

// TestDictionary_Dev_HybridExtension_IsolatedFromProdDSL verifies that the dev DSL
// extension of a hybrid handler uses ONLY the dev-defined properties and not any
// properties from the prod DSL extension of the same handler.
func TestDictionary_Dev_HybridExtension_IsolatedFromProdDSL(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewHybridEntityHandler(),
	}
	// Both prod and dev DSL extend "hybrid-entity" with different fields.
	core := newTestRegistryCore(t, handlers, testdataProdPath, testdataDevPath)

	// Prod sees the prod field.
	prodDict, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)
	require.Contains(t, prodDict, "sdk/hybrid-entity")
	prodProps := *prodDict["sdk/hybrid-entity"].EndorHandler.EntitySchema.Properties
	assert.Contains(t, prodProps, "additionalProdField")
	assert.NotContains(t, prodProps, "additionalDevField")

	// Dev sees ONLY the dev field — prod field must NOT bleed through.
	devSession := sdk.Session{Development: true, Username: "user1"}
	devDict, err := core.Dictionary(devSession)
	require.NoError(t, err)
	require.Contains(t, devDict, "sdk/hybrid-entity")
	devProps := *devDict["sdk/hybrid-entity"].EndorHandler.EntitySchema.Properties
	assert.Contains(t, devProps, "additionalDevField",
		"dev DSL property should be present in dev session")
	assert.NotContains(t, devProps, "additionalProdField",
		"prod DSL property must not be visible in dev session")
}

// TestDictionary_Dev_UserIsolation verifies that two different users each see only
// their own dev DSL entities and not each other's.
func TestDictionary_Dev_UserIsolation(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", testdataDevPath)

	user1Dict, err := core.Dictionary(sdk.Session{Development: true, Username: "user1"})
	require.NoError(t, err)

	user2Dict, err := core.Dictionary(sdk.Session{Development: true, Username: "user2"})
	require.NoError(t, err)

	// user1 sees their own entity, not user2's.
	assert.Contains(t, user1Dict, "sdk/dev-entity")
	assert.NotContains(t, user1Dict, "sdk/user2-entity")

	// user2 sees their own entity, not user1's.
	assert.Contains(t, user2Dict, "sdk/user2-entity")
	assert.NotContains(t, user2Dict, "sdk/dev-entity")
}

// TestDictionary_Dev_EphemeralCacheIsUsed verifies that successive Dictionary calls
// for the same dev user return the cached entry without rebuilding.
func TestDictionary_Dev_EphemeralCacheIsUsed(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", testdataDevPath)

	session := sdk.Session{Development: true, Username: "user1"}

	// First call — builds and caches.
	_, err := core.Dictionary(session)
	require.NoError(t, err)

	cached, _ := core.EphemeralCache.Get("user1")
	require.NotNil(t, cached, "ephemeral cache should be populated after first call")

	// Second call — must hit the cache.
	dict2, err := core.Dictionary(session)
	require.NoError(t, err)
	assert.NotNil(t, dict2)
	assert.Equal(t, len(cached), len(dict2))
}

// TestDictionary_Dev_FallsBackToStaticWhenNoDSL verifies that when building the dev
// dictionary fails (unknown user, missing DSL path), Dictionary falls back to the
// static dictionary rather than returning an error.
//
// Note: because buildDevDictionary in this implementation never hard-errors
// (DSL overlay failures are logged as warnings), this test exercises the case
// where a user with no dev DSL files at all sees only the static dictionary.
func TestDictionary_Dev_FallsBackToStaticWhenNoDSL(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewBaseEntityHandler(),
	}
	// Empty devDSLBase: no DSL files exist for "unknown-user".
	core := newTestRegistryCore(t, handlers, "", "")

	devSession := sdk.Session{Development: true, Username: "unknown-user"}
	dict, err := core.Dictionary(devSession)
	require.NoError(t, err)

	// Only static handlers should be present.
	assert.Contains(t, dict, "sdk/base-entity")
	assert.Len(t, dict, 1)
}

// ---------------------------------------------------------------------------
// DI container wiring
// ---------------------------------------------------------------------------

// TestContainer_Prod_RepositoriesArePopulated verifies that after the production
// dictionary is built the DI container holds all expected repository entries —
// one per registered hybrid handler (and their categories).
func TestContainer_Prod_RepositoriesArePopulated(t *testing.T) {
	handlers := []sdk.EndorHandlerInterface{
		examples_handlers.NewHybridEntityHandler(),
		examples_handlers.NewHybridSpecializedEntityHandler(),
	}
	core := newTestRegistryCore(t, handlers, "", "")

	_, err := core.Dictionary(sdk.Session{})
	require.NoError(t, err)

	container, err := core.Container(sdk.Session{})
	require.NoError(t, err)
	require.NotNil(t, container)

	repos := container.GetRepositories()
	require.NotEmpty(t, repos, "container should hold at least one repository")

	// hybrid-entity contributes one repo; hybrid-specialized-entity contributes
	// one root repo plus one per category (keyed as "<entity>/<category>").
	assert.Contains(t, repos, "hybrid-entity", "hybrid-entity repository should be registered")
	assert.Contains(t, repos, "hybrid-specialized-entity", "hybrid-specialized-entity repository should be registered")
	assert.Contains(t, repos, "hybrid-specialized-entity/cat-1", "cat-1 repository should be registered")
	assert.Contains(t, repos, "hybrid-specialized-entity/cat-2", "cat-2 repository should be registered")
}
