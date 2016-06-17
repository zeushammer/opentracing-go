package opentracing

import "time"

// Tracer is a simple, thin interface for Span creation and SpanMetadata
// propagation.
type Tracer interface {

	// Create, start, and return a new Span with the given `operationName` and
	// incorporate the given StartSpanOptions. (See StartSpanOption for all of
	// the options on that front)
	//
	// Examples:
	//
	//     var tracer opentracing.Tracer = ...
	//
	//     // The root-span case:
	//     sp := tracer.StartSpan("GetFeed")
	//
	//     // The vanilla child span case:
	//     sp := tracer.StartSpan(
	//         "GetFeed",
	//         opentracing.RefBlockedParent.Point(parentSpan.Metadata()))
	//
	//     // All the bells and whistles:
	//     sp := tracer.StartSpan(
	//         "GetFeed",
	//         opentracing.RefBlockedParent.Point(parentSpan.Metadata()),
	//         opentracing.Tags{
	//             "user_agent": loggedReq.UserAgent,
	//         },
	//         opentracing.StartTime(loggedReq.Timestamp),
	//     )
	//
	StartSpan(operationName string, opts ...StartSpanOption) Span

	// Inject() takes the `sm` SpanMetadata instance and injects it for
	// propagation within `carrier`. The actual type of `carrier` depends on
	// the value of `format`.
	//
	// OpenTracing defines a common set of `format` values (see BuiltinFormat),
	// and each has an expected carrier type.
	//
	// Other packages may declare their own `format` values, much like the keys
	// used by `context.Context` (see
	// https://godoc.org/golang.org/x/net/context#WithValue).
	//
	// Example usage (sans error handling):
	//
	//     carrier := opentracing.HTTPHeaderTextMapCarrier(httpReq.Header)
	//     err := tracer.Inject(
	//         span.Metadata(),
	//         opentracing.TextMap,
	//         carrier)
	//
	// NOTE: All opentracing.Tracer implementations MUST support all
	// BuiltinFormats.
	//
	// Implementations may return opentracing.ErrUnsupportedFormat if `format`
	// is not supported by (or not known by) the implementation.
	//
	// Implementations may return opentracing.ErrInvalidCarrier or any other
	// implementation-specific error if the format is supported but injection
	// fails anyway.
	//
	// See Tracer.Extract().
	Inject(sm SpanMetadata, format interface{}, carrier interface{}) error

	// Extract() returns a SpanMetadata instance given `format` and `carrier`.
	//
	// OpenTracing defines a common set of `format` values (see BuiltinFormat),
	// and each has an expected carrier type.
	//
	// Other packages may declare their own `format` values, much like the keys
	// used by `context.Context` (see
	// https://godoc.org/golang.org/x/net/context#WithValue).
	//
	// Example usage:
	//
	//
	//     carrier := opentracing.HTTPHeaderTextMapCarrier(httpReq.Header)
	//     spanMetadata, err := tracer.Extract(opentracing.TextMap, carrier)
	//     startSpanOptions := make([]opentracing.StartSpanOption, 0, 1)
	//
	//     // ... assuming the ultimate goal here is to resume the trace with a
	//     // server-side Span:
	//     if err == nil {
	//         startSpanOptions = append(
	//             startSpanOptions,
	//             opentracing.Reference(opentracing.RefRPCClient, spanMetadata))
	//     }
	//     span := tracer.StartSpan(
	//         rpcMethodName, opentracing.Reference(opentracing.RefRPCClient, spanMetadata))
	//
	//
	// NOTE: All opentracing.Tracer implementations MUST support all
	// BuiltinFormats.
	//
	// Return values:
	//  - A successful Extract returns a SpanMetadata instance and a nil error
	//  - If there was simply no SpanMetadata to extract in `carrier`, Extract()
	//    returns (nil, opentracing.ErrTraceNotFound)
	//  - If `format` is unsupported or unrecognized, Extract() returns (nil,
	//    opentracing.ErrUnsupportedFormat)
	//  - If there are more fundamental problems with the `carrier` object,
	//    Extract() may return opentracing.ErrInvalidCarrier,
	//    opentracing.ErrTraceCorrupted, or implementation-specific errors.
	//
	// See Tracer.Inject().
	Extract(format interface{}, carrier interface{}) (SpanMetadata, error)
}

// StartSpanOptions allows Tracer.StartSpanWithOptions callers to override the
// start timestamp, specify a parent Span, and make sure that Tags are
// available at Span initialization time.
type StartSpanOptions struct {
	// Zero or more causal references to other Spans (via their SpanMetadata).
	// If empty, start a "root" Span (i.e., start a new trace).
	References []SpanReference

	// StartTime overrides the Span's start time, or implicitly becomes
	// time.Now() if StartTime.IsZero().
	StartTime time.Time

	// Tags may have zero or more entries; the restrictions on map values are
	// identical to those for Span.SetTag(). May be nil.
	//
	// If specified, the caller hands off ownership of Tags at
	// StartSpan() invocation time.
	Tags map[string]interface{}
}

// StartSpanOption instances (zero or more) may be passed to Tracer.StartSpan.
type StartSpanOption interface {
	Apply(*StartSpanOptions)
}

// ReferenceType is an enum type describing different categories of
// relationships between two Spans. If Span-2 refers to Span-1, the
// ReferenceType describes Span-1 from Span-2's perspective. For example,
// RefBlockedParent means that Span-1 caused Span-2, and that it's blocked on
// Span-2's finish.
type ReferenceType int

const (
	// RefBlockedParent refers to a "parent" Span that (a) caused the creation
	// of the newly-started span, and (b) is blocked on the newly-started Span.
	// See RefStartedBefore for the non-blocking version. Timing diagram:
	//
	//     [-Referent------------]
	//               [-New Span-]
	//
	RefBlockedParent ReferenceType = iota

	// RefRPCClient is a special case of RefBlockedParent for the server side
	// of RPCs to refer to the client side of RPCs. Timing diagram:
	//
	//     [-Remote Client Referent------]
	//                [-New Server Span-]
	//
	RefRPCClient

	// RefStartedBefore refers to a Span that merely started before the
	// newly-started Span. This is the weakest of the ReferenceType causality
	// assertions. Timing diagram:
	//
	//     [-Referent-]
	//          [-New Span-]
	//
	RefStartedBefore

	// RefFinishedBefore refers to a Span that finished (and started) before
	// the newly-created Span. Timing diagram:
	//
	//     [-Referent-]
	//                    [-New Span-]
	//
	RefFinishedBefore
)

// SpanReference pairs a ReferenceType and a referent SpanMetadata. See the
// ReferenceType documentation.
type SpanReference struct {
	Type     ReferenceType
	Metadata SpanMetadata
}

// Apply satisfies the StartSpanOption interface.
func (r SpanReference) Apply(o *StartSpanOptions) {
	o.References = append(o.References, r)
}

// IntraTrace returns true if the referent is in the same trace as the referee
// (i.e., if the referee should adopt the referent's trace-scoped metadata,
// like a `trace_id`).
func (r ReferenceType) IntraTrace() bool {
	return r <= RefFinishedBefore
}

// Point returns a StartSpanOption that describes a SpanReference (from the
// Span that's about to be started to the `referent`).
func (r ReferenceType) Point(referent SpanMetadata) SpanReference {
	return SpanReference{
		Type:     r,
		Metadata: referent,
	}
}

// StartTime is a StartSpanOption that sets an explicit start timestamp for the
// new Span.
type StartTime time.Time

// Apply satisfies the StartSpanOption interface.
func (t StartTime) Apply(o *StartSpanOptions) {
	o.StartTime = time.Time(t)
}

// Tags are a generic map from an arbitrary string key to an opaque value type.
// The underlying tracing system is responsible for interpreting and
// serializing the values.
type Tags map[string]interface{}

// Merge incorporates the keys and values from `other` into this `Tags`
// instance, then returns same.
func (t Tags) Merge(other Tags) Tags {
	for k, v := range other {
		t[k] = v
	}
	return t
}

// Apply satisfies the StartSpanOption interface.
func (t Tags) Apply(o *StartSpanOptions) {
	o.Tags = t
}
