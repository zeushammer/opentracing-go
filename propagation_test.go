package opentracing

import (
	"net/http"
	"strconv"
	"testing"
)

const testHeaderPrefix = "testprefix-"

func TestHTTPHeaderInject(t *testing.T) {
	h := http.Header{}
	h.Add("NotOT", "blah")
	h.Add("opname", "AlsoNotOT")
	tracer := testTracer{}
	span := tracer.StartSpan("someSpan")
	fakeID := span.Metadata().(testSpanMetadata).FakeID

	// Use HTTPHeaderTextMapCarrier to wrap around `h`.
	carrier := HTTPHeaderTextMapCarrier(h)
	if err := span.Tracer().Inject(span.Metadata(), TextMap, carrier); err != nil {
		t.Fatal(err)
	}

	if len(h) != 3 {
		t.Errorf("Unexpected header length: %v", len(h))
	}
	// The prefix comes from just above; the suffix comes from
	// testTracer.Inject().
	if h.Get("testprefix-fakeid") != strconv.Itoa(fakeID) {
		t.Errorf("Could not find fakeid at expected key")
	}
}

func TestHTTPHeaderJoin(t *testing.T) {
	h := http.Header{}
	h.Add("NotOT", "blah")
	h.Add("opname", "AlsoNotOT")
	h.Add("testprefix-fakeid", "42")
	tracer := testTracer{}

	// Use HTTPHeaderTextMapCarrier to wrap around `h`.
	carrier := HTTPHeaderTextMapCarrier(h)
	spanMetadata, err := tracer.Extract(TextMap, carrier)
	if err != nil {
		t.Fatal(err)
	}

	if spanMetadata.(testSpanMetadata).FakeID != 42 {
		t.Errorf("Failed to read testprefix-fakeid correctly")
	}
}
