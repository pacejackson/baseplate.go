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

type MonitoredRedis interface {
	redis.Cmdable

	AddHook(hook redis.Hook)
	WithMonitoredContext(ctx context.Context) MonitoredRedis
}

type monitoredClient struct {
	*redis.Client
}

func (c *monitoredClient) WithMonitoredContext(ctx context.Context) MonitoredRedis {
	return &monitoredClient{Client: c.Client.WithContext(ctx)}
}

type monitoredClusterClient struct {
	*redis.ClusterClient
}

func (c *monitoredClusterClient) WithMonitoredContext(ctx context.Context) MonitoredRedis {
	return &monitoredClusterClient{ClusterClient: c.ClusterClient.WithContext(ctx)}
}

type monitoredRing struct {
	*redis.Ring
}

func (c *monitoredRing) WithMonitoredContext(ctx context.Context) MonitoredRedis {
	return &monitoredRing{Ring: c.Ring.WithContext(ctx)}
}

// MonitoredRedisFactory is used to create Redis clients that are monitored by
// a RedisSpanHook.
type MonitoredRedisFactory struct {
	client MonitoredRedis
}

func newMonitoredRedisFactory(name string, client MonitoredRedis) MonitoredRedisFactory {
	client.AddHook(RedisSpanHook{ClientName: name})
	return MonitoredRedisFactory{client: client}
}

func NewMonitoredRedisClient(name string, client *redis.Client) MonitoredRedisFactory {
	return newMonitoredRedisFactory(name, &monitoredClient{Client: client})
}

func NewMonitoredRedisClusterClient(name string, client *redis.ClusterClient) MonitoredRedisFactory {
	return newMonitoredRedisFactory(name, &monitoredClusterClient{ClusterClient: client})
}

func NewMonitoredRedisRing(name string, client *redis.Ring) MonitoredRedisFactory {
	return newMonitoredRedisFactory(name, &monitoredRing{Ring: client})
}

// BuildClient returns a new, monitored Redis client with the given context.
func (f MonitoredRedisFactory) BuildClient(ctx context.Context) MonitoredRedis {
	return f.client.WithMonitoredContext(ctx)
}

var (
	_ MonitoredRedis = (*monitoredClient)(nil)
	_ MonitoredRedis = (*monitoredClusterClient)(nil)
	_ MonitoredRedis = (*monitoredRing)(nil)
)
