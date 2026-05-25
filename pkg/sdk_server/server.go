package sdk_server

import (
	"fmt"
	"io/fs"
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
	endorHandlers *[]sdk.EndorHandlerInterface
	postInitFunc  func()
	version       string
	localesFS     fs.FS
}

type EndorInitializer struct {
	endor *Endor
}

func NewEndorInitializer() *EndorInitializer {
	return &EndorInitializer{
		endor: &Endor{
			version: "v1",
		},
	}
}

func (b *EndorInitializer) WithEndorHandlers(handlers *[]sdk.EndorHandlerInterface) *EndorInitializer {
	b.endor.endorHandlers = handlers
	return b
}

func (b *EndorInitializer) WithPostInitFunc(f func()) *EndorInitializer {
	b.endor.postInitFunc = f
	return b
}

func (b *EndorInitializer) WithLocalesFS(localesFS fs.FS) *EndorInitializer {
	if sub, err := fs.Sub(localesFS, "locales"); err == nil {
		b.endor.localesFS = sub
	} else {
		b.endor.localesFS = localesFS
	}
	return b
}

func (b *EndorInitializer) Build() *Endor {
	return b.endor
}

func (h *Endor) Init(module string) {
	microServiceId := fmt.Sprintf("endor-%s-service", module)
	// load configuration
	config := sdk_configuration.GetConfig()
	config.ModuleDBName = microServiceId

	// create initialization logger
	logger := sdk.NewLogger(sdk.LogConfig{
		LogType: sdk.LogType(config.LogType),
	}, sdk.LogContext{})

	// load i18n translations
	translator := sdk_i18n.NewTranslator(h.localesFS)

	// create router
	router := gin.New()

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
		*h.endorHandlers = append(*h.endorHandlers, sdk_entity_aggregation.NewAggregationHandler(0, nil))
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
		*h.endorHandlers = append(*h.endorHandlers, sdk_entity.NewEntityHandler(microServiceId, module, h.endorHandlers, logger, 0, nil))
		*h.endorHandlers = append(*h.endorHandlers, sdk_entity.NewEntityActionHandler(microServiceId, module, h.endorHandlers, logger))
	}

	// Initialize the singleton repository after all handlers are registered.
	// Must be called after all handler appends so the registry is complete.
	EndorHandlerRepository := sdk_entity.InitEndorEntityRepository(microServiceId, module, h.endorHandlers, logger, h.localesFS)
	entities, err := EndorHandlerRepository.EndorHandlerList()
	if err != nil {
		log.Fatal(err)
	}

	actionRepo := EndorHandlerRepository.ActionRepository()

	router.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		if strings.HasPrefix(urlPath, "/api/v1/") {
			// actionId format: module/entity/[category/]action
			actionId := strings.TrimPrefix(urlPath, "/api/v1/")
			_, entity, category, action, err := sdk.ParseEntityActionID(actionId)
			if err == nil {
				// Build the session here: single point of header parsing.
				// Development=true activates the per-user ephemeral registry overlay.
				session := sdk.Session{
					Id:          c.GetHeader("x-user-session"),
					Username:    c.GetHeader("x-user-id"),
					Development: c.GetHeader("x-development") == "true",
					Locale:      sdk_i18n.NormalizeLocale(c.GetHeader("Accept-Language")),
				}
				dict, err := actionRepo.DictionaryActionInstance(session, sdk.ReadInstanceDTO{Id: actionId})
				if err == nil {
					dict.EndorHandlerAction.CreateHTTPCallback(module, entity, action, category, session, dict.Container)(c)
					return
				}
			}
		}
		response := sdk.NewDefaultResponseBuilder()
		locale := sdk_i18n.NormalizeLocale(c.GetHeader("Accept-Language"))
		response.AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityFatal, translator.T(locale, "sdk.commons.not_found", map[string]any{"uri": c.Request.RequestURI, "method": c.Request.Method})))
		c.JSON(http.StatusNotFound, response.Build())
	})

	err = api_gateway.InitializeApiGatewayConfiguration(microServiceId, module, fmt.Sprintf("http://%s:%s", microServiceId, config.ServerPort), entities)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := swagger.CreateSwaggerConfiguration(microServiceId, module, fmt.Sprintf("http://localhost:%s", config.ServerPort), entities, "/api", h.localesFS)
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
