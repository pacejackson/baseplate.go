package redisbp

import (
	"context"

	"github.com/go-redis/redis/v7"
)

type MonitoredCmdable interface {
	redis.Cmdable

	AddHook(hook redis.Hook)
	WithMonitoredContext(ctx context.Context) MonitoredCmdable
}

type monitoredClient struct {
	*redis.Client
}

func (c *monitoredClient) WithMonitoredContext(ctx context.Context) MonitoredCmdable {
	return &monitoredClient{Client: c.Client.WithContext(ctx)}
}

type monitoredCluster struct {
	*redis.ClusterClient
}

func (c *monitoredCluster) WithMonitoredContext(ctx context.Context) MonitoredCmdable {
	return &monitoredCluster{ClusterClient: c.ClusterClient.WithContext(ctx)}
}

type monitoredRing struct {
	*redis.Ring
}

func (c *monitoredRing) WithMonitoredContext(ctx context.Context) MonitoredCmdable {
	return &monitoredRing{Ring: c.Ring.WithContext(ctx)}
}

// MonitoredCmdableFactory is used to create Redis clients that are monitored by
// a SpanHook.
type MonitoredCmdableFactory struct {
	client MonitoredCmdable
}

func newMonitoredCmdableFactory(name string, client MonitoredCmdable) MonitoredCmdableFactory {
	client.AddHook(SpanHook{ClientName: name})
	return MonitoredCmdableFactory{client: client}
}

func NewMonitoredClientFactory(name string, client *redis.Client) MonitoredCmdableFactory {
	return newMonitoredCmdableFactory(name, &monitoredClient{Client: client})
}

func NewMonitoredClusterFactory(name string, client *redis.ClusterClient) MonitoredCmdableFactory {
	return newMonitoredCmdableFactory(name, &monitoredCluster{ClusterClient: client})
}

func NewMonitoredRingFactory(name string, client *redis.Ring) MonitoredCmdableFactory {
	return newMonitoredCmdableFactory(name, &monitoredRing{Ring: client})
}

// BuildClient returns a new, monitored Redis client with the given context.
func (f MonitoredCmdableFactory) BuildClient(ctx context.Context) MonitoredCmdable {
	return f.client.WithMonitoredContext(ctx)
}

var (
	_ MonitoredCmdable = (*monitoredClient)(nil)
	_ MonitoredCmdable = (*monitoredCluster)(nil)
	_ MonitoredCmdable = (*monitoredRing)(nil)
)
