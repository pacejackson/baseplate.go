package integrations

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v7"

	"github.com/reddit/baseplate.go/batcherror"
	"github.com/reddit/baseplate.go/tracing"
)

// RedisSpanHook is a redis.Hook for wrapping Redis commands and pipelines
// in Client Spans and metrics.
type RedisSpanHook struct {
	ClientName string
}

var _ redis.Hook = RedisSpanHook{}

// BeforeProcess starts a client Span before processing a Redis command and
// starts a timer to record how long the command took.
func (h RedisSpanHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return h.startChildSpan(ctx, cmd.Name()), nil
}

// AfterProcess ends the client Span started by BeforeProcess, publishes the
// time the Redis command took to complete, and a metric indicating whether the
// command was a "success" or "fail"
func (h RedisSpanHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	return h.endChildSpan(ctx, cmd.Err())
}

// BeforeProcessPipeline starts a client span before processing a Redis pipeline
// and starts a timer to record how long the pipeline took.
func (h RedisSpanHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return h.startChildSpan(ctx, "pipeline"), nil
}

// AfterProcessPipeline ends the client span started by BeforeProcessPipeline,
// publishes the time the Redis pipeline took to complete, and a metric
// indicating whether the pipeline was a "success" or "fail"
func (h RedisSpanHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	var errs batcherror.BatchError
	for _, cmd := range cmds {
		errs.Add(cmd.Err())
	}
	return h.endChildSpan(ctx, errs.Compile())
}

func (h RedisSpanHook) startChildSpan(ctx context.Context, cmdName string) context.Context {
	// Get the current span tracing the work being done by ctx.  Try to get a
	// sub-span first and fall back to the server span if we are not currently
	// in a sub-span.
	//
	// We are going to use this span to create a child span that is attached to
	// a new context and used by go-redis to trace the command/pipeline.
	span := tracing.GetActiveSpan(ctx)
	if span == nil {
		span = tracing.GetServerSpan(ctx)
	}
	if span == nil {
		return ctx
	}
	name := fmt.Sprintf("%s.%s", h.ClientName, cmdName)
	ctx, _ = span.CreateClientChildForContext(ctx, name)
	return ctx
}

func (h RedisSpanHook) endChildSpan(ctx context.Context, err error) error {
	if span := tracing.GetActiveSpan(ctx); span != nil {
		return span.Stop(ctx, err)
	}
	return nil
}

// MonitoredRedisFactory is used to create Redis clients that are monitored by
// a RedisSpanHook.
type MonitoredRedisFactory interface {
	// BuildClient returns a new, monitored redis.Cmdable with the given context.
	BuildClient(ctx context.Context) redis.Cmdable
}

// RedisClientFactory is used by a service to create a new, non-failover redis.Client
// using the current context and monitored by a baseplate.go RedisSpanHook to
// inject into an endpoint that needs to use Redis.
//
// See https://pkg.go.dev/github.com/go-redis/redis/v7?tab=doc#Client for documentation
// about redis.Client.
type RedisClientFactory struct {
	client *redis.Client
}

// NewRedisClientFactory creates a new RedisClusterFactory with the given name and
// options.
func NewRedisClientFactory(name string, opt *redis.Options) RedisClientFactory {
	client := redis.NewClient(opt)
	client.AddHook(RedisSpanHook{ClientName: name})
	return RedisClientFactory{client: client}
}

// NewRedisSentinelClientFactory creates a new RedisClusterFactory with the
// given name and options.
func NewRedisSentinelClientFactory(name string, opt *redis.FailoverOptions) RedisClientFactory {
	client := redis.NewFailoverClient(opt)
	client.AddHook(RedisSpanHook{ClientName: name})
	return RedisClientFactory{client: client}
}

// BuildClient returns a new, monitored redis.Client with the given context.
func (f RedisClientFactory) BuildClient(ctx context.Context) redis.Cmdable {
	return f.client.WithContext(ctx)
}

// RedisClusterFactory is used by a service to create a new redis.ClusterClient
// using the current context and monitored by a baseplate.go RedisSpanHook to
// inject into an endpoint that needs to use Redis.
//
// See https://pkg.go.dev/github.com/go-redis/redis/v7?tab=doc#ClusterClient for
// documentation about redis.ClusterClient and https://redis.io/topics/cluster-tutorial
// for information about Redis Cluster.
type RedisClusterFactory struct {
	client *redis.ClusterClient
}

// NewRedisClusterFactory creates a new RedisClusterFactory with the given name and
// options.
func NewRedisClusterFactory(name string, opt *redis.ClusterOptions) RedisClusterFactory {
	client := redis.NewClusterClient(opt)
	client.AddHook(RedisSpanHook{ClientName: name})
	return RedisClusterFactory{client: client}
}

// BuildClient returns a new, monitored redis.ClusterClient with the given context.
func (f RedisClusterFactory) BuildClient(ctx context.Context) redis.Cmdable {
	return f.client.WithContext(ctx)
}

// RedisRingFactory is used by a service to create a new redis.Ring
// using the current context and monitored by a baseplate.go RedisSpanHook to
// inject into an endpoint that needs to use Redis.
//
// See https://pkg.go.dev/github.com/go-redis/redis/v7?tab=doc#Ring for documentation
// about redis.Ring
type RedisRingFactory struct {
	client *redis.Ring
}

// NewRedisRingFactory creates a new RedisRingFactory with the given name and
// cluster options.
func NewRedisRingFactory(name string, opt *redis.RingOptions) RedisRingFactory {
	client := redis.NewRing(opt)
	client.AddHook(RedisSpanHook{ClientName: name})
	return RedisRingFactory{client: client}
}

// BuildClient returns a new, monitored redis.RingClient with the given context.
func (f RedisRingFactory) BuildClient(ctx context.Context) redis.Cmdable {
	return f.client.WithContext(ctx)
}

var (
	_ MonitoredRedisFactory = RedisClientFactory{}
	_ MonitoredRedisFactory = RedisClusterFactory{}
	_ MonitoredRedisFactory = RedisRingFactory{}
)
