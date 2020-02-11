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

// This example demonstrates how to use a RedisClientFactory to create monitored
// Redis clients.
func ExampleRedisClientFactory() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewRedisClientFactory("redis", &redis.Options{Addr: ":6379"})
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}

// This example demonstrates how to use a RedisClusterFactory to create monitored
// Redis Cluster clients.
func ExampleRedisClusterFactory() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewRedisClusterFactory(
		"redis",
		&redis.ClusterOptions{
			Addrs: []string{":6379", ":6380", ":6381"},
		},
	)
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}

// This example demonstrates how to use a RedisSentinelClientFactory to create monitored
// Redis clients that implement failover using Redis Sentinel.
func ExampleRedisClientFactory_sentinel() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewRedisSentinelClientFactory(
		"redis",
		&redis.FailoverOptions{
			MasterName:    "master",
			SentinelAddrs: []string{":6379"},
		},
	)
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}

// This example demonstrates how to use a RedisRingFactory to create monitored
// Redis go-redis ring clients.
func ExampleRedisRingFactory() {
	// variables should be properly initialized in production code
	var tracer *tracing.Tracer
	// Create a factory
	factory := integrations.NewRedisRingFactory(
		"redis",
		&redis.RingOptions{
			Addrs: map[string]string{
				"shard0": ":6379",
				"shard1": ":6380",
				"shard2": ":6381",
			},
		},
	)
	// Get a context object and a server Span, with the server Span set on the
	// context
	ctx, _ := tracing.CreateServerSpanForContext(context.Background(), tracer, "test")
	// Create a new client using the context for the Server Span
	client := factory.BuildClient(ctx)
	// Commands should now be wrapped using Client Spans
	client.Ping()
}
