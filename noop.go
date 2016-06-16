package opentracing

// A NoopTracer is a trivial implementation of Tracer for which all operations
// are no-ops.
type NoopTracer struct{}

type noopSpan struct{}
type noopSpanMetadata struct{}

var (
	defaultNoopSpanMetadata = noopSpanMetadata{}
	defaultNoopSpan         = noopSpan{}
	defaultNoopTracer       = NoopTracer{}
)

const (
	emptyString = ""
)

// noopSpanMetadata:
func (n noopSpanMetadata) SetBaggageItem(key, val string) SpanMetadata { return n }
func (n noopSpanMetadata) BaggageItem(key string) string               { return emptyString }

// noopSpan:
func (n noopSpan) Metadata() SpanMetadata                                { return defaultNoopSpanMetadata }
func (n noopSpan) SetTag(key string, value interface{}) Span             { return n }
func (n noopSpan) Finish()                                               {}
func (n noopSpan) FinishWithOptions(opts FinishOptions)                  {}
func (n noopSpan) LogEvent(event string)                                 {}
func (n noopSpan) LogEventWithPayload(event string, payload interface{}) {}
func (n noopSpan) Log(data LogData)                                      {}
func (n noopSpan) SetOperationName(operationName string) Span            { return n }
func (n noopSpan) Tracer() Tracer                                        { return defaultNoopTracer }

// StartSpan belongs to the Tracer interface.
func (n NoopTracer) StartSpan(operationName string, opts ...StartSpanOption) Span {
	return defaultNoopSpan
}

// Inject belongs to the Tracer interface.
func (n NoopTracer) Inject(sp SpanMetadata, format interface{}, carrier interface{}) error {
	return nil
}

// Extract belongs to the Tracer interface.
func (n NoopTracer) Extract(format interface{}, carrier interface{}) (SpanMetadata, error) {
	return nil, ErrSpanMetadataNotFound
}
