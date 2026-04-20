package sdk_server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity_aggregation"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Endor struct {
	endorHandlers     *[]sdk.EndorHandlerInterface
	endorRepositories []sdk.EndorRepositoryInterface
	postInitFunc      func()
}

type EndorInitializer struct {
	endor *Endor
}

func NewEndorInitializer() *EndorInitializer {
	return &EndorInitializer{
		endor: &Endor{},
	}
}

func (b *EndorInitializer) WithEndorHandlers(handlers *[]sdk.EndorHandlerInterface) *EndorInitializer {
	b.endor.endorHandlers = handlers
	return b
}

func (b *EndorInitializer) WithEndorRepositories(repositories []sdk.EndorRepositoryInterface) *EndorInitializer {
	b.endor.endorRepositories = repositories
	return b
}

func (b *EndorInitializer) WithPostInitFunc(f func()) *EndorInitializer {
	b.endor.postInitFunc = f
	return b
}

func (b *EndorInitializer) Build() *Endor {
	return b.endor
}

func (h *Endor) Init(microserviceId string) {
	// load configuration
	config := sdk_configuration.GetConfig()

	// create initialization logger
	logger := sdk.NewLogger(sdk.LogConfig{
		LogType: sdk.LogType(config.LogType),
	}, sdk.LogContext{})

	// load i18n translations
	if err := sdk_i18n.Init("./locales"); err != nil {
		logger.Info("i18n: failed to initialize translations: " + err.Error())
	}

	// define runtime configuration
	config.DynamicEntityDocumentDBName = microserviceId

	// create router
	router := gin.New()

	// registrer repositories
	for _, r := range h.endorRepositories {
		sdk.GetRepositoryRegistry().Register(r.GetEntity(), r)
	}

	// monitoring
	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Check if an EndorHandler with entity == "aggregation" is already defined
	aggregationServiceExists := false
	if h.endorHandlers != nil {
		for _, svc := range *h.endorHandlers {
			if svc.GetEntity() == "aggregation" {
				aggregationServiceExists = true
				break
			}
		}
	}
	if !aggregationServiceExists {
		*h.endorHandlers = append(*h.endorHandlers, sdk_entity_aggregation.NewAggregationHandler(0))
	}

	// Check if an EndorHandler with entity == "entity" is already defined
	entityServiceExists := false
	if h.endorHandlers != nil {
		for _, svc := range *h.endorHandlers {
			if svc.GetEntity() == "entity" {
				entityServiceExists = true
				break
			}
		}
	}
	if !entityServiceExists {
		*h.endorHandlers = append(*h.endorHandlers, sdk_entity.NewEntityHandler(microserviceId, h.endorHandlers, nil, logger, 0, config.HybridEntitiesEnabled, config.DynamicEntitiesEnabled))
		*h.endorHandlers = append(*h.endorHandlers, sdk_entity.NewEntityActionHandler(microserviceId, h.endorHandlers))
	}

	// get all entities (initialize singleton repository)
	EndorHandlerRepository := sdk_entity.InitEndorHandlerRepository(microserviceId, h.endorHandlers, logger)
	entities, err := EndorHandlerRepository.EndorHandlerList()
	if err != nil {
		log.Fatal(err)
	}

	actionRepo := EndorHandlerRepository.ActionRepository()

	router.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		if strings.HasPrefix(urlPath, "/api/") {
			// actionId format: ms-id/version/entity/[category/]action
			actionId := strings.TrimPrefix(urlPath, "/api/")
			// Build the session here: single point of header parsing.
			// Development=true activates the per-user ephemeral registry overlay.
			session := sdk.Session{
				Id:          c.GetHeader("x-user-session"),
				Username:    c.GetHeader("x-user-id"),
				Development: c.GetHeader("x-development") == "true",
			}
			dict, err := actionRepo.DictionaryActionInstance(session, sdk.ReadInstanceDTO{Id: actionId})
			if err == nil {
				// segments: [0]=ms-id [1]=version [2]=entity [3+]=action parts
				segments := strings.Split(actionId, "/")
				entity, actionKey, category := "", "", ""
				if len(segments) >= 4 {
					entity = segments[2]
					actionKey = strings.Join(segments[3:], "/")
					if len(segments) >= 5 {
						category = segments[3]
					}
				}
				dict.EndorHandlerAction.CreateHTTPCallback(microserviceId, entity, actionKey, category, session, &dict.Container)(c)
				return
			}
		}
		response := sdk.NewDefaultResponseBuilder()
		locale := sdk_i18n.NormalizeLocale(c.GetHeader("Accept-Language"))
		response.AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityFatal, sdk_i18n.T(locale, "commons.not_found", map[string]any{"uri": c.Request.RequestURI, "method": c.Request.Method})))
		c.JSON(http.StatusNotFound, response.Build())
	})

	err = api_gateway.InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), entities)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := swagger.CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), entities, "/api")
	if err != nil {
		log.Fatal(err)
	}

	// swagger
	router.StaticFS("/swagger", http.Dir(swaggerPath))

	// post initialization
	if h.postInitFunc != nil {
		h.postInitFunc()
	}

	// start http server
	router.Run()
}
