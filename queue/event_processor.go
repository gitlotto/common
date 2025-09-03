package queue

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type EventProcessor interface {
	ProcessSingle(ctx context.Context, event *events.SQSMessage, logger *zap.Logger) (err error)
}

func ProcessMultiple(
	ctx context.Context,
	sqsEvents events.SQSEvent,
	eventProcessor EventProcessor,
	logger *zap.Logger,
) (commandsProcessed events.SQSEventResponse) {

	tracer := otel.Tracer("github.com/gitlotto/common/queue")
	ctx, span := tracer.Start(ctx, "queue.process_multiple", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	logger.Info("Processing events in total", zap.Int("events", len(sqsEvents.Records)))

	failures := []events.SQSBatchItemFailure{}

	for _, event := range sqsEvents.Records {
		var otelAttributes propagation.MapCarrier = make(propagation.MapCarrier)

		for key, value := range event.MessageAttributes {
			if value.DataType != "String" || value.StringValue == nil {
				continue
			}

			stringValue := *value.StringValue

			switch key {
			case "X-Amzn-Trace-Id":
				logger.Info("Found X-Amzn-Trace-Id header", zap.String("traceId", stringValue))
				otelAttributes["X-Amzn-Trace-Id"] = stringValue
			case "SpanContextJson":
				logger.Info("Found SpanContextJson header", zap.String("spanContextJson", stringValue))
			}
		}

		propagator := otel.GetTextMapPropagator()
		extractedCtx := propagator.Extract(ctx, otelAttributes)
		singleEventCtx, span := tracer.Start(extractedCtx, "queue.process_single", trace.WithSpanKind(trace.SpanKindInternal))

		errOfTheMessage := eventProcessor.ProcessSingle(singleEventCtx, &event, logger)
		if errOfTheMessage != nil {
			eventFailure := &events.SQSBatchItemFailure{
				ItemIdentifier: event.MessageId,
			}
			failures = append(failures, *eventFailure)
		}

		span.End()
	}

	if len(failures) > 0 {
		logger.Error("Some events failed", zap.Int("eventFailures", len(failures)))
		commandsProcessed.BatchItemFailures = failures
	}
	return
}
