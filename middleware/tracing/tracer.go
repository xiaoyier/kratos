package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is otel span tracer
type Tracer struct {
	tracer trace.Tracer
	kind   trace.SpanKind
}

// NewTracer create tracer instance
func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	options := options{}
	for _, o := range opts {
		o(&options)
	}
	if options.TracerProvider != nil {
		otel.SetTracerProvider(options.TracerProvider)
	}
	if options.Propagators != nil {
		otel.SetTextMapPropagator(options.Propagators)
	} else {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}))
	}
	var name string
	if kind == trace.SpanKindServer {
		name = "server"
	} else if kind == trace.SpanKindClient {
		name = "client"
	} else {
		panic(fmt.Sprintf("unsupported span kind: %v", kind))
	}
	tracer := otel.Tracer(name)
	return &Tracer{tracer: tracer, kind: kind}
}

// Start start tracing span
func (t *Tracer) Start(ctx context.Context, component string, operation string, carrier propagation.TextMapCarrier) (context.Context, trace.Span) {
	if t.kind == trace.SpanKindServer {
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}
	ctx, span := t.tracer.Start(ctx,
		operation,
		trace.WithAttributes(attribute.String("component", component)),
		trace.WithSpanKind(t.kind),
	)
	if t.kind == trace.SpanKindClient {
		otel.GetTextMapPropagator().Inject(ctx, carrier)
	}
	return ctx, span
}

// End finish tracing span
func (t *Tracer) End(ctx context.Context, span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("event", "error"),
			attribute.String("message", err.Error()),
		)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}
	span.End()
}
