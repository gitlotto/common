package queue

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
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

	logger.Info("Processing events in total", zap.Int("events", len(sqsEvents.Records)))

	failures := []events.SQSBatchItemFailure{}

	for _, event := range sqsEvents.Records {
		errOfTheMessage := eventProcessor.ProcessSingle(ctx, &event, logger)
		if errOfTheMessage != nil {
			eventFailure := &events.SQSBatchItemFailure{
				ItemIdentifier: event.MessageId,
			}
			failures = append(failures, *eventFailure)
		}
	}

	if len(failures) > 0 {
		logger.Error("Some events failed", zap.Int("eventFailures", len(failures)))
		commandsProcessed.BatchItemFailures = failures
	}
	return
}
