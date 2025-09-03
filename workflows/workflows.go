package workflows

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gitlotto/common/zulu"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/gitlotto/common/workflows")

func ErrFifoWorkflowQueueMismatch(queueUrl string) error {
	return fmt.Errorf("fifo workflow should not delegate to a simple SQS queue %s", queueUrl)
}

func ErrSpanContextMarshalFailed(err error) error {
	return fmt.Errorf("failed to marshal span context: %w", err)
}

type WorkflowRecord struct {
	EventId             string            `dynamodbav:"event_id"`
	TargetQueueUrl      string            `dynamodbav:"target_queue_url"`
	CreatedAt           zulu.DateTime     `dynamodbav:"created_at"`
	StartAt             zulu.DateTime     `dynamodbav:"start_at"`
	AmountOfStarts      int               `dynamodbav:"amount_of_starts"`
	IsOpen              *IsOpen           `dynamodbav:"is_open,omitempty"`
	FinishedAt          *zulu.DateTime    `dynamodbav:"finished_at,omitempty"`
	Event               string            `dynamodbav:"event"`
	EventMessageGroupId string            `dynamodbav:"event_message_group_id"`
	Baggage             map[string]string `dynamodbav:"baggage"`
	SpanContextJson     string            `dynamodbav:"span_context_json"`
}

func (record WorkflowRecord) PartitionKey() types.AttributeValue {
	return &types.AttributeValueMemberS{
		Value: record.EventId,
	}
}

func (record WorkflowRecord) SortKey() *types.AttributeValue {
	var sk types.AttributeValue = &types.AttributeValueMemberS{
		Value: record.TargetQueueUrl,
	}
	return &sk
}

func (record WorkflowRecord) EventMessageDeduplicationId() string {
	deduplicationIdInBytes := sha256.Sum256([]byte(record.EventId))
	deduplicationIdInString := hex.EncodeToString(deduplicationIdInBytes[:])
	return deduplicationIdInString
}

func NewFifoWorkflowRecord(
	tableName string,
	partitionKey string,
	sortKey *string,
	createdAt zulu.DateTime,
	startAt zulu.DateTime,
	targetQueueUrl string,
	event string,
	eventGroupId string,
	baggage map[string]string,
	spanContextJson string,
) (*WorkflowRecord, error) {
	if targetQueueUrl == "" || !strings.HasSuffix(targetQueueUrl, ".fifo") {
		return nil, ErrFifoWorkflowQueueMismatch(targetQueueUrl)
	}
	isOpen := Open
	eventId := NewEventId(tableName, partitionKey, sortKey)
	amountOfStarts := 0
	workflowRecord := WorkflowRecord{
		EventId:             eventId.String(),
		CreatedAt:           createdAt,
		StartAt:             startAt,
		AmountOfStarts:      amountOfStarts,
		IsOpen:              &isOpen,
		FinishedAt:          nil,
		TargetQueueUrl:      targetQueueUrl,
		Event:               event,
		EventMessageGroupId: eventGroupId,
		Baggage:             baggage,
		SpanContextJson:     spanContextJson,
	}
	return &workflowRecord, nil
}

func NewFifoWorkflowRecordFromContext(
	ctx context.Context,
	tableName string,
	partitionKey string,
	sortKey *string,
	createdAt zulu.DateTime,
	startAt zulu.DateTime,
	targetQueueUrl string,
	event string,
	eventGroupId string,
) (context.Context, *WorkflowRecord, error) {
	ctx, span := tracer.Start(ctx, "workflows.publish_command", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	baggage := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(baggage))

	spanContext := trace.SpanContextFromContext(ctx)
	spanContextJson, err := json.Marshal(spanContext)
	if err != nil {
		return ctx, nil, ErrSpanContextMarshalFailed(err)
	}
	spanContextJsonString := string(spanContextJson)
	if targetQueueUrl == "" || !strings.HasSuffix(targetQueueUrl, ".fifo") {
		err = ErrFifoWorkflowQueueMismatch(targetQueueUrl)
		return ctx, nil, err
	}
	isOpen := Open
	eventId := NewEventId(tableName, partitionKey, sortKey)
	amountOfStarts := 0
	workflowRecord := &WorkflowRecord{
		EventId:             eventId.String(),
		CreatedAt:           createdAt,
		StartAt:             startAt,
		AmountOfStarts:      amountOfStarts,
		IsOpen:              &isOpen,
		FinishedAt:          nil,
		TargetQueueUrl:      targetQueueUrl,
		Event:               event,
		EventMessageGroupId: eventGroupId,
		Baggage:             baggage,
		SpanContextJson:     spanContextJsonString,
	}
	return ctx, workflowRecord, nil
}

type IsOpen string

const (
	Open IsOpen = "OPEN"
)

type EventId struct {
	value string
}

func NewEventId(tableName string, partitionKey string, sortKey *string) EventId {
	idPrefix := tableName + "#" + partitionKey
	if sortKey == nil {
		return EventId{value: idPrefix}
	}
	return EventId{value: idPrefix + "#" + *sortKey}
}

func (id EventId) String() string {
	return id.value
}
