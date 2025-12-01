package composition

import (
	"context"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// BenchmarkService provides a minimal service implementation for benchmarks
type BenchmarkService struct {
	resource string
}

func (s *BenchmarkService) GetResource() string                                  { return s.resource }
func (s *BenchmarkService) GetDescription() string                               { return "Benchmark service" }
func (s *BenchmarkService) GetMethods() map[string]interfaces.EndorServiceAction { return nil }
func (s *BenchmarkService) GetPriority() *int                                    { return nil }
func (s *BenchmarkService) GetVersion() string                                   { return "1.0.0" }
func (s *BenchmarkService) Validate() error                                      { return nil }

func BenchmarkServiceChain(b *testing.B) {
	services := make([]interfaces.EndorServiceInterface, 5)
	for i := 0; i < 5; i++ {
		services[i] = &BenchmarkService{resource: "test"}
	}

	chain := ServiceChain(services...)
	ctx := context.Background()
	request := "benchmark-request"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := chain.Execute(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceProxy(b *testing.B) {
	target := &BenchmarkService{resource: "test"}
	interceptor := &NoOpInterceptor{}
	proxy := ServiceProxy(target, interceptor)

	ctx := context.Background()
	request := "benchmark-request"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := proxy.Execute(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceBranch(b *testing.B) {
	router := &SimpleRouter{DefaultService: "service1"}
	services := map[string]interfaces.EndorServiceInterface{
		"service1": &BenchmarkService{resource: "service1"},
		"service2": &BenchmarkService{resource: "service2"},
	}

	branch := ServiceBranch(router, services)
	ctx := context.Background()
	request := "benchmark-request"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := branch.Execute(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceMerger(b *testing.B) {
	services := []interfaces.EndorServiceInterface{
		&BenchmarkService{resource: "service1"},
		&BenchmarkService{resource: "service2"},
	}
	merger := &FirstWinsMerger{}
	mergedService := ServiceMerger(services, merger)

	ctx := context.Background()
	request := "benchmark-request"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := mergedService.Execute(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}
