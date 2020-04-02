package thriftbp

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"
)

// BaseplateProcessor is a TProcessor that can be thriftbp.Wrap-ed and
// thriftbp.Merge-d.
//
// The TProcessors generated by the Apache Thrift compiler fufill this
// interface, but not all of them are a part of any interface within Apache
// Thrift.
type BaseplateProcessor interface {
	thrift.TProcessor

	// ProcessorMap returns a map of thrift method names to TProcessorFunctions.
	ProcessorMap() map[string]thrift.TProcessorFunction

	// AddToProcessorMap adds the given TProcessorFunction to the internal
	// processor map at the given key.
	//
	// If one is already set at the given key, it will be replaced with the new
	// TProcessorFunction.
	AddToProcessorMap(string, thrift.TProcessorFunction)
}

// Middleware is a function that can be passed to Wrap to wrap
// the TProcessorFunctions for that TProcessor.
//
// Middlewares are passed in the name of the function as set in the processor
// map of the TProcessor and a logger that can be used by the Middleware.
type Middleware func(name string, next thrift.TProcessorFunction) thrift.TProcessorFunction

// WrappedTProcessorFunc is a conveinence struct that implements the
// TProcessorFunction interface that you can pass in a wrapped function that
// will be called by Process.
type WrappedTProcessorFunc struct {
	// Wrapped is called by WrappedTProcessorFunc and should be a "wrapped"
	// call to a base TProcessorFunc.Process call.
	Wrapped func(ctx context.Context, seqId int32, in, out thrift.TProtocol) (bool, thrift.TException)
}

// Process implements the TProcessorFunction interface by calling and returning
// p.Wrapped.
func (p WrappedTProcessorFunc) Process(ctx context.Context, seqID int32, in, out thrift.TProtocol) (bool, thrift.TException) {
	return p.Wrapped(ctx, seqID, in, out)
}

var (
	_ thrift.TProcessorFunction = WrappedTProcessorFunc{}
	_ thrift.TProcessorFunction = (*WrappedTProcessorFunc)(nil)
)

// Wrap takes an existing BaseplateProcessor and wraps each of its inner
// TProcessorFunctions with the middlewares passed in and returns it.
//
// Middlewares will be called in the order that they are defined:
//
//		1. Middlewares[0]
//		2. Middlewares[1]
//		...
//		N. Middlewares[n]
//
// It is reccomended that you pass in tracing.InjectThriftServerSpan and the
// Middleware returned by edgecontext.InjectThriftEdgeContext as the first two
// middlewares.
func Wrap(processor BaseplateProcessor, middlewares ...Middleware) thrift.TProcessor {
	for name, processorFunc := range processor.ProcessorMap() {
		wrapped := processorFunc
		for i := len(middlewares) - 1; i >= 0; i-- {
			wrapped = middlewares[i](name, wrapped)
		}
		processor.AddToProcessorMap(name, wrapped)
	}
	return processor
}
