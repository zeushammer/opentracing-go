package opentracing

import (
	"strconv"
	"strings"
)

const testHTTPHeaderPrefix = "testprefix-"

// testTracer is a most-noop Tracer implementation that makes it possible for
// unittests to verify whether certain methods were / were not called.
type testTracer struct{}

var fakeIDSource = 1

func nextFakeID() int {
	fakeIDSource++
	return fakeIDSource
}

type testSpanMetadata struct {
	HasParent bool
	FakeID    int
}

func (n testSpanMetadata) SetBaggageItem(key, val string) SpanMetadata { return n }
func (n testSpanMetadata) BaggageItem(key string) string               { return "" }

type testSpan struct {
	spanMetadata  testSpanMetadata
	OperationName string
}

// testSpan:
func (n testSpan) Metadata() SpanMetadata                                { return n.spanMetadata }
func (n testSpan) SetTag(key string, value interface{}) Span             { return n }
func (n testSpan) Finish()                                               {}
func (n testSpan) FinishWithOptions(opts FinishOptions)                  {}
func (n testSpan) LogEvent(event string)                                 {}
func (n testSpan) LogEventWithPayload(event string, payload interface{}) {}
func (n testSpan) Log(data LogData)                                      {}
func (n testSpan) SetOperationName(operationName string) Span            { return n }
func (n testSpan) Tracer() Tracer                                        { return testTracer{} }

// StartSpan belongs to the Tracer interface.
func (n testTracer) StartSpan(operationName string, opts ...StartSpanOption) Span {
	sso := StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return n.startSpanWithOptions(operationName, sso)
}

func (n testTracer) startSpanWithOptions(name string, opts StartSpanOptions) Span {
	fakeID := nextFakeID()
	if len(opts.References) > 0 {
		fakeID = opts.References[0].Metadata.(testSpanMetadata).FakeID
	}
	return testSpan{
		OperationName: name,
		spanMetadata: testSpanMetadata{
			HasParent: len(opts.References) > 0,
			FakeID:    fakeID,
		},
	}
}

// Inject belongs to the Tracer interface.
func (n testTracer) Inject(sp SpanMetadata, format interface{}, carrier interface{}) error {
	spanMetadata := sp.(testSpanMetadata)
	switch format {
	case TextMap:
		carrier.(TextMapWriter).Set(testHTTPHeaderPrefix+"fakeid", strconv.Itoa(spanMetadata.FakeID))
		return nil
	}
	return ErrUnsupportedFormat
}

// Extract belongs to the Tracer interface.
func (n testTracer) Extract(format interface{}, carrier interface{}) (SpanMetadata, error) {
	switch format {
	case TextMap:
		// Just for testing purposes... generally not a worthwhile thing to
		// propagate.
		sm := testSpanMetadata{}
		err := carrier.(TextMapReader).ForeachKey(func(key, val string) error {
			switch strings.ToLower(key) {
			case testHTTPHeaderPrefix + "fakeid":
				i, err := strconv.Atoi(val)
				if err != nil {
					return err
				}
				sm.FakeID = i
			}
			return nil
		})
		return sm, err
	}
	return nil, ErrSpanMetadataNotFound
}
