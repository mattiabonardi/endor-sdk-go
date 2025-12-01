//go:build unit

package composition

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// MockService is a test implementation of EndorServiceInterface
type MockService struct {
	mock.Mock
}

func (m *MockService) GetResource() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockService) GetDescription() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockService) GetMethods() map[string]interfaces.EndorServiceAction {
	args := m.Called()
	return args.Get(0).(map[string]interfaces.EndorServiceAction)
}

func (m *MockService) GetPriority() *int {
	args := m.Called()
	return args.Get(0).(*int)
}

func (m *MockService) GetVersion() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockService) Validate() error {
	args := m.Called()
	return args.Error(0)
}

func TestServiceChain(t *testing.T) {
	t.Run("creates chain with multiple services", func(t *testing.T) {
		service1 := &MockService{}
		service2 := &MockService{}

		chain := ServiceChain(service1, service2)

		assert.NotNil(t, chain)
		assert.Equal(t, 2, chain.Length())

		services := chain.GetServices()
		assert.Len(t, services, 2)
		assert.Equal(t, service1, services[0])
		assert.Equal(t, service2, services[1])
	})

	t.Run("executes services sequentially", func(t *testing.T) {
		service1 := &MockService{}
		service2 := &MockService{}

		service1.On("Validate").Return(nil)
		service2.On("Validate").Return(nil)

		chain := ServiceChain(service1, service2)

		ctx := context.Background()
		request := "test-request"

		result, err := chain.Execute(ctx, request)

		assert.NoError(t, err)
		assert.Equal(t, request, result)

		service1.AssertCalled(t, "Validate")
		service2.AssertCalled(t, "Validate")
	})
}

func TestServiceProxy(t *testing.T) {
	t.Run("creates proxy with target service", func(t *testing.T) {
		targetService := &MockService{}
		interceptor := &NoOpInterceptor{}

		proxy := ServiceProxy(targetService, interceptor)

		assert.NotNil(t, proxy)
		assert.Equal(t, targetService, proxy.GetTarget())
		assert.Equal(t, interceptor, proxy.GetInterceptor())
	})

	t.Run("executes target service through interceptor", func(t *testing.T) {
		targetService := &MockService{}
		interceptor := &NoOpInterceptor{}

		targetService.On("Validate").Return(nil)

		proxy := ServiceProxy(targetService, interceptor)

		ctx := context.Background()
		request := "test-request"

		result, err := proxy.Execute(ctx, request)

		assert.NoError(t, err)
		assert.Equal(t, request, result)

		targetService.AssertCalled(t, "Validate")
	})
}

func TestServiceBranch(t *testing.T) {
	t.Run("creates branch with router and services", func(t *testing.T) {
		router := &SimpleRouter{DefaultService: "service1"}
		services := map[string]interfaces.EndorServiceInterface{
			"service1": &MockService{},
			"service2": &MockService{},
		}

		branch := ServiceBranch(router, services)

		assert.NotNil(t, branch)
		assert.Equal(t, router, branch.GetRouter())
		assert.Len(t, branch.GetServices(), 2)
	})
}

func TestServiceMerger(t *testing.T) {
	t.Run("creates merger with services and merger", func(t *testing.T) {
		service1 := &MockService{}
		service2 := &MockService{}
		services := []interfaces.EndorServiceInterface{service1, service2}
		merger := &AllResultsMerger{}

		mergedService := ServiceMerger(services, merger)

		assert.NotNil(t, mergedService)
		assert.Equal(t, merger, mergedService.GetMerger())
		assert.Len(t, mergedService.GetServices(), 2)
	})
}

func TestCompositionError(t *testing.T) {
	t.Run("creates composition error with context", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewCompositionError("test-op", "test-service", 1, cause)

		assert.Equal(t, "test-op", err.Operation)
		assert.Equal(t, "test-service", err.ServiceName)
		assert.Equal(t, 1, err.ServiceIndex)
		assert.Equal(t, cause, err.Cause)

		errWithContext := err.WithContext("requestId", "12345")
		assert.Equal(t, "12345", errWithContext.Context["requestId"])
	})
}
