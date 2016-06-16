package mocktracer

import (
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
)

// New returns a MockTracer opentracing.Tracer implementation that's intended
// to facilitate tests of OpenTracing instrumentation.
func New() *MockTracer {
	return &MockTracer{
		FinishedSpans: []*MockSpan{},
	}
}

// MockTracer is a for-testing-only opentracing.Tracer implementation. It is
// entirely unsuitable for production use but appropriate for tests that want
// to verify tracing behavior.
type MockTracer struct {
	FinishedSpans []*MockSpan
}

// MockSpanMetadata is an opentracing.SpanMetadata implementation.
type MockSpanMetadata struct {
	SpanID  int
	Baggage map[string]string
}

// MockSpan is an opentracing.Span implementation that exports its internal
// state for testing purposes.
type MockSpan struct {
	ParentID      int
	OperationName string
	StartTime     time.Time
	FinishTime    time.Time
	Tags          map[string]interface{}
	Logs          []opentracing.LogData

	tracer       *MockTracer
	spanMetadata *MockSpanMetadata
}

// Reset clears the exported MockTracer.FinishedSpans field. Note that any
// extant MockSpans will still append to FinishedSpans when they Finish(), even
// after a call to Reset().
func (t *MockTracer) Reset() {
	t.FinishedSpans = []*MockSpan{}
}

// StartSpan belongs to the Tracer interface.
func (t *MockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o(&sso)
	}
	return newMockSpan(t, operationName, sso)
}

const mockTextMapIdsPrefix = "mockpfx-ids-"
const mockTextMapBaggagePrefix = "mockpfx-baggage-"

// Inject belongs to the Tracer interface.
func (t *MockTracer) Inject(sm opentracing.SpanMetadata, format interface{}, carrier interface{}) error {
	spanMetadata := sm.(*MockSpanMetadata)
	switch format {
	case opentracing.TextMap:
		writer := carrier.(opentracing.TextMapWriter)
		// Ids:
		writer.Set(mockTextMapIdsPrefix+"spanid", strconv.Itoa(spanMetadata.SpanID))
		// Baggage:
		for baggageKey, baggageVal := range spanMetadata.Baggage {
			writer.Set(mockTextMapBaggagePrefix+baggageKey, baggageVal)
		}
		return nil
	}
	return opentracing.ErrUnsupportedFormat
}

// Extract belongs to the Tracer interface.
func (t *MockTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanMetadata, error) {
	switch format {
	case opentracing.TextMap:
		rval := newMockSpanMetadata(0)
		err := carrier.(opentracing.TextMapReader).ForeachKey(func(key, val string) error {
			lowerKey := strings.ToLower(key)
			switch {
			case lowerKey == mockTextMapIdsPrefix+"spanid":
				// Ids:
				i, err := strconv.Atoi(val)
				if err != nil {
					return err
				}
				rval.SpanID = i
			case strings.HasPrefix(lowerKey, mockTextMapBaggagePrefix):
				// Baggage:
				rval.Baggage[lowerKey[len(mockTextMapBaggagePrefix):]] = val
			}
			return nil
		})
		return rval, err
	}
	return nil, opentracing.ErrSpanMetadataNotFound
}

var mockIDSource = 1

func nextMockID() int {
	mockIDSource++
	return mockIDSource
}

func newMockSpanMetadata(spanID int) *MockSpanMetadata {
	return &MockSpanMetadata{
		SpanID:  spanID,
		Baggage: make(map[string]string),
	}
}

// SetBaggageItem belongs to the SpanMetadata interface
func (s *MockSpanMetadata) SetBaggageItem(key, val string) opentracing.SpanMetadata {
	s.Baggage[key] = val
	return s
}

// BaggageItem belongs to the SpanMetadata interface
func (s *MockSpanMetadata) BaggageItem(key string) string {
	return s.Baggage[key]
}

func newMockSpan(t *MockTracer, name string, opts opentracing.StartSpanOptions) *MockSpan {
	tags := opts.Tags
	if tags == nil {
		tags = map[string]interface{}{}
	}
	parentID := int(0)
	if len(opts.CausalReferences) > 0 {
		parentID = opts.CausalReferences[0].SpanMetadata.(*MockSpanMetadata).SpanID
	}
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}
	return &MockSpan{
		ParentID:      parentID,
		OperationName: name,
		StartTime:     startTime,
		Tags:          tags,
		Logs:          []opentracing.LogData{},

		tracer:       t,
		spanMetadata: newMockSpanMetadata(nextMockID()),
	}
}

// Metadata belongs to the Span interface
func (s *MockSpan) Metadata() opentracing.SpanMetadata {
	return s.spanMetadata
}

// SetTag belongs to the Span interface
func (s *MockSpan) SetTag(key string, value interface{}) opentracing.Span {
	s.Tags[key] = value
	return s
}

// Finish belongs to the Span interface
func (s *MockSpan) Finish() {
	s.FinishTime = time.Now()
	s.tracer.FinishedSpans = append(s.tracer.FinishedSpans, s)
}

// FinishWithOptions belongs to the Span interface
func (s *MockSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	s.FinishTime = opts.FinishTime
	s.Logs = append(s.Logs, opts.BulkLogData...)
	s.tracer.FinishedSpans = append(s.tracer.FinishedSpans, s)
}

// LogEvent belongs to the Span interface
func (s *MockSpan) LogEvent(event string) {
	s.Log(opentracing.LogData{
		Event: event,
	})
}

// LogEventWithPayload belongs to the Span interface
func (s *MockSpan) LogEventWithPayload(event string, payload interface{}) {
	s.Log(opentracing.LogData{
		Event:   event,
		Payload: payload,
	})
}

// Log belongs to the Span interface
func (s *MockSpan) Log(data opentracing.LogData) {
	s.Logs = append(s.Logs, data)
}

// SetOperationName belongs to the Span interface
func (s *MockSpan) SetOperationName(operationName string) opentracing.Span {
	s.OperationName = operationName
	return s
}

// Tracer belongs to the Span interface
func (s *MockSpan) Tracer() opentracing.Tracer {
	return s.tracer
}
