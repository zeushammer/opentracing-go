package ext_test

import (
	"reflect"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Not equal: %#v (expected)\n"+
			"        != %#v (actual)", expected, actual)
	}
}

func TestPeerTags(t *testing.T) {
	if ext.PeerService != "peer.service" {
		t.Fatalf("Invalid PeerService %v", ext.PeerService)
	}
	var span opentracing.Span = newNoopTagSpan()
	ext.PeerService.Set(span, "my-service")
	ext.PeerHostname.Set(span, "my-hostname")
	ext.PeerHostIPv4.Set(span, uint32(127<<24|1))
	ext.PeerHostIPv6.Set(span, "::")
	ext.PeerPort.Set(span, uint16(8080))
	ext.SamplingPriority.Set(span, uint16(1))
	ext.SpanKind.Set(span, ext.SpanKindRPCServer)
	ext.SpanKind.Set(span, ext.SpanKindRPCClient)
	span.Finish()

	rawSpan := span.(*noopSpan)
	assertEqual(t, "my-service", rawSpan.Tags["peer.service"])
	assertEqual(t, "my-hostname", rawSpan.Tags["peer.hostname"])
	assertEqual(t, uint32(127<<24|1), rawSpan.Tags["peer.ipv4"])
	assertEqual(t, "::", rawSpan.Tags["peer.ipv6"])
	assertEqual(t, uint16(8080), rawSpan.Tags["peer.port"])
}

func TestHTTPTags(t *testing.T) {
	var span opentracing.Span = newNoopTagSpan()
	ext.HTTPUrl.Set(span, "test.biz/uri?protocol=false")
	ext.HTTPMethod.Set(span, "GET")
	ext.HTTPStatusCode.Set(span, 301)
	span.Finish()

	rawSpan := span.(*noopSpan)
	assertEqual(t, "test.biz/uri?protocol=false", rawSpan.Tags["http.url"])
	assertEqual(t, "GET", rawSpan.Tags["http.method"])
	assertEqual(t, uint16(301), rawSpan.Tags["http.status_code"])
}

func TestMiscTags(t *testing.T) {
	var span opentracing.Span = newNoopTagSpan()
	ext.Component.Set(span, "my-awesome-library")
	ext.SamplingPriority.Set(span, 1)
	span.Finish()

	rawSpan := span.(*noopSpan)
	assertEqual(t, "my-awesome-library", rawSpan.Tags["component"])
	assertEqual(t, uint16(1), rawSpan.Tags["sampling.priority"])
}

type noopSpanMetadata struct{}

type noopSpan struct {
	Tags opentracing.Tags
}

func (n noopSpanMetadata) SetBaggageItem(key, val string) opentracing.SpanMetadata { return n }
func (n noopSpanMetadata) BaggageItem(key string) string                           { return "" }

func (n noopSpan) Metadata() opentracing.SpanMetadata { return noopSpanMetadata{} }

func (n noopSpan) SetTag(key string, value interface{}) opentracing.Span {
	n.Tags[key] = value
	return n
}

func newNoopTagSpan() *noopSpan { return &noopSpan{make(opentracing.Tags)} }

func (n noopSpan) Finish()                                                {}
func (n noopSpan) FinishWithOptions(opts opentracing.FinishOptions)       {}
func (n noopSpan) LogEvent(event string)                                  {}
func (n noopSpan) LogEventWithPayload(event string, payload interface{})  {}
func (n noopSpan) Log(data opentracing.LogData)                           {}
func (n noopSpan) SetOperationName(operationName string) opentracing.Span { return n }
func (n noopSpan) Tracer() opentracing.Tracer                             { return nil }
