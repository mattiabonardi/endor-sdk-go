package sdk

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Endor struct {
	internalEndorServices  *[]EndorService
	internalHybridServices *[]EndorHybridService
	postInitFunc           func()
}

type EndorInitializer struct {
	endor *Endor
}

func NewEndorInitializer() *EndorInitializer {
	return &EndorInitializer{
		endor: &Endor{},
	}
}

func (b *EndorInitializer) WithEndorServices(services *[]EndorService) *EndorInitializer {
	b.endor.internalEndorServices = services
	return b
}

func (b *EndorInitializer) WithHybridServices(services *[]EndorHybridService) *EndorInitializer {
	b.endor.internalHybridServices = services
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
	config := GetConfig()

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
	if h.internalEndorServices != nil {
		for _, svc := range *h.internalEndorServices {
			if svc.Resource == "resource" {
				resourceServiceExists = true
				break
			}
		}
	}
	if !resourceServiceExists {
		*h.internalEndorServices = append(*h.internalEndorServices, *NewResourceService(microserviceId, h.internalEndorServices, h.internalHybridServices))
		*h.internalEndorServices = append(*h.internalEndorServices, *NewResourceActionService(microserviceId, h.internalEndorServices, h.internalHybridServices))
	}

	// get all resources
	EndorServiceRepository := NewEndorServiceRepository(microserviceId, h.internalEndorServices, h.internalHybridServices)
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
			endorRepositoryDictionary, err := EndorServiceRepository.Instance(ReadInstanceDTO{
				Id: resource,
			})
			if err == nil {
				if method, ok := endorRepositoryDictionary.EndorService.Methods[action]; ok {
					method.CreateHTTPCallback(microserviceId)(c)
					return
				}
			}
		}
		response := NewDefaultResponseBuilder()
		response.AddMessage(NewMessage(Fatal, "404 page not found (uri: "+c.Request.RequestURI+", method: "+c.Request.Method+")"))
		c.JSON(http.StatusNotFound, response.Build())
	})

	err = InitializeApiGatewayConfiguration(microserviceId, fmt.Sprintf("http://%s:%s", microserviceId, config.ServerPort), resources)
	if err != nil {
		log.Fatal(err)
	}
	swaggerPath, err := CreateSwaggerConfiguration(microserviceId, fmt.Sprintf("http://localhost:%s", config.ServerPort), resources, "/api")
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
