package sdk_server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/internal/api_gateway"
	"github.com/mattiabonardi/endor-sdk-go/internal/configuration"
	"github.com/mattiabonardi/endor-sdk-go/internal/swagger"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_resource"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Endor struct {
	endorServices *[]sdk.EndorServiceInterface
	postInitFunc  func()
}

type EndorInitializer struct {
	endor *Endor
}

func NewEndorInitializer() *EndorInitializer {
	return &EndorInitializer{
		endor: &Endor{},
	}
}

func (b *EndorInitializer) WithEndorServices(services *[]sdk.EndorServiceInterface) *EndorInitializer {
	b.endor.endorServices = services
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
	config := configuration.GetConfig()

	// define runtime configuration
	config.DynamicResourceDocumentDBName = microserviceId

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

	// Check if an EndorService with resource == "resource" is already defined
	resourceServiceExists := false
	if h.endorServices != nil {
		for _, svc := range *h.endorServices {
			if svc.GetResource() == "resource" {
				resourceServiceExists = true
				break
			}
		}
	}
	if !resourceServiceExists {
		*h.endorServices = append(*h.endorServices, sdk_resource.NewResourceService(microserviceId, h.endorServices))
		*h.endorServices = append(*h.endorServices, sdk_resource.NewResourceActionService(microserviceId, h.endorServices))
	}

	// get all resources
	EndorServiceRepository := sdk_resource.NewEndorServiceRepository(microserviceId, h.endorServices)
	resources, err := EndorServiceRepository.EndorServiceList()
	if err != nil {
		log.Fatal(err)
	}

	router.NoRoute(func(c *gin.Context) {
		// find the resource in path /api/{version}/{resource}/{method}
		pathSegments := strings.Split(c.Request.URL.Path, "/")
		if len(pathSegments) > 4 {
			resource := pathSegments[3]
			action := pathSegments[4]
			if len(pathSegments) == 6 {
				action = pathSegments[4] + "/" + pathSegments[5]
			}
			endorRepositoryDictionary, err := EndorServiceRepository.Instance(sdk.ReadInstanceDTO{
				Id: resource,
			})
			if err == nil {
				if method, ok := endorRepositoryDictionary.EndorService.Actions[action]; ok {
					category := ""
					if strings.Contains(action, "/") {
						parts := strings.SplitN(action, "/", 2)
						category = parts[0]
					}
					method.CreateHTTPCallback(microserviceId, resource, action, category)(c)
					return
				}
			}
		}
		response := sdk.NewDefaultResponseBuilder()
		response.AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityFatal, "404 page not found (uri: "+c.Request.RequestURI+", method: "+c.Request.Method+")"))
		c.JSON(http.StatusNotFound, response.Build())
	})

	err = api_gateway.InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), resources)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := swagger.CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), resources, "/api")
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
