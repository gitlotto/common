package logging

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const AmznTraceIdFieldName = "X-Amzn-Trace-Id"

func MustCreateZuluTimeLogger() (logger *zap.Logger) {

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z"))
	}

	config.EncoderConfig.EncodeTime = timeEncoder

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger
}


func WithTraceID(ctx context.Context, logger *zap.Logger) (loggerOut *zap.Logger) {
	loggerOut = logger
	
	baggage := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(baggage))
	
	if amznTraceId, ok := baggage[AmznTraceIdFieldName]; ok {
		loggerOut = loggerOut.With(zap.String(AmznTraceIdFieldName, amznTraceId))
	}

	spanContext := trace.SpanContextFromContext(ctx)
	
	loggerOut = loggerOut.With(zap.String("trace_id", spanContext.TraceID().String()))
	return loggerOut

}