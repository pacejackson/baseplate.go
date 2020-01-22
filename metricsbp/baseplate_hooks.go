package metricsbp

import (
	"fmt"

	"github.com/reddit/baseplate.go/tracing"
)

const (
	success = "success"
	fail    = "fail"
)

// BaseplateHook registers each Server Span with a MetricsSpanHook.
type BaseplateHook struct {
	Prefix  string
	Metrics Statsd
}

// OnServerSpanCreate registers MetricSpanHooks on a Server Span.
func (h BaseplateHook) OnServerSpanCreate(span *tracing.Span) error {
	serverSpanHook := newSpanHook(h.Prefix, h.Metrics, span)
	span.RegisterHook(serverSpanHook)
	return nil
}

// SpanHook wraps a Span in a Timer and records a "success" or "fail"
// metric when the Span ends based on whether an error was passed to `span.End`
// or not.
type SpanHook struct {
	Name    string
	Metrics Statsd

	timer *Timer
}

func newSpanHook(prefix string, metrics Statsd, span *tracing.Span) SpanHook {
	name := fmt.Sprintf("%s.%s.%s", prefix, span.Type().String(), span.Name)
	return SpanHook{
		Name:    name,
		Metrics: metrics,
		timer:   NewTimer(metrics.Histogram(name)),
	}
}

// OnCreateChild registers a child MetricsSpanHook on the child Span and starts
// a new Timer around the Span.
func (h SpanHook) OnCreateChild(child *tracing.Span) error {
	childHook := newSpanHook(h.Name, h.Metrics, child)
	child.RegisterHook(childHook)
	return nil
}

// OnStart is a nop
func (h SpanHook) OnStart(span *tracing.Span) error {
	return nil
}

// OnEnd stops the Timer started in OnStart and records a metric indicating if
// the span was a "success" or "fail".
//
// A span is marked as "fail" if `err != nil` otherwise it is marked as
// "success".
func (h SpanHook) OnEnd(span *tracing.Span, err error) error {
	h.timer.ObserveDuration()
	var statusMetricPath string
	if err != nil {
		statusMetricPath = fmt.Sprintf("%s.%s", h.Name, fail)
	} else {
		statusMetricPath = fmt.Sprintf("%s.%s", h.Name, success)
	}
	h.Metrics.Counter(statusMetricPath).Add(1)
	return nil
}

var (
	_ tracing.BaseplateHook = BaseplateHook{}
	_ tracing.SpanHook      = SpanHook{}
)
