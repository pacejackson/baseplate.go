package integrations_test

import (
	"context"

	"github.com/go-redis/redis/v7"

	"github.com/reddit/baseplate.go/integrations"
	"github.com/reddit/baseplate.go/tracing"
)

// This example demonstrates how to use RedisSpanHook to automatically add Spans
// around Redis commands using go-redis
//
// baseplate.go also provides a set of MonitoredRedisFactory objects that you can use
// to create Redis clients with a SpanHook already attached.
func ExampleRedisSpanHook() {
	// variables should be properly initialized in production code
	var (
		// baseClient is not actually used to run commands, we register the Hook
		// to it and use it to create clients for each Server Span.
		baseClient redis.Client
		tracer     *tracing.Tracer
	)
	// Add the Hook onto baseClient
	baseClient.AddHook(integrations.RedisSpanHook{ClientName: "redis"})
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := baseClient.WithContext(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}

// This example demonstrates how to use a MonitoredRedisFactory to create
// monitored redis.Client objects.
func ExampleMonitoredRedisFactory_client() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewMonitoredRedisClient(
		"redis",
		redis.NewClient(&redis.Options{Addr: ":6379"}),
	)
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}

// This example demonstrates how to use a MonitoredRedisFactory to create
// monitored redis.ClusterClient objects.
func ExampleMonitoredRedisFactory_cluster() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewMonitoredRedisClusterClient(
		"redis",
		redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{":7000", ":7001", ":7002"},
		}),
	)
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}

// This example demonstrates how to use a MonitoredRedisFactory to create
// monitored redis.Client objects that implement failover using Redis Sentinel.
func ExampleMonitoredRedisFactory_sentinel() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewMonitoredRedisClient(
		"redis",
		redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    "master",
			SentinelAddrs: []string{":6379"},
		}),
	)
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}

// This example demonstrates how to use a MonitoredRedisFactory to create
// monitored redis.Ring objects.
func ExampleMonitoredRedisFactory_ring() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewMonitoredRedisRing(
		"redis",
		redis.NewRing(&redis.RingOptions{
			Addrs: map[string]string{
				"shard0": ":7000",
				"shard1": ":7001",
				"shard2": ":7002",
			},
		}),
	)
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}
