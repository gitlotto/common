package workflows

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/gitlotto/common/zulu"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_new_fifo_workflowRecord_should_not_be_created_if_queue_is_simple(t *testing.T) {

	tableName := uuid.New().String()
	partitionKey := uuid.New().String()
	createdAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 15, 12, 45, 14, 0, time.UTC))
	startAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 16, 12, 45, 14, 0, time.UTC))
	targetQueueUrl := uuid.New().String()

	event := fmt.Sprintf(`{"partitionKey":"%s","sortKey":null}`, partitionKey)
	eventGroupId := uuid.New().String()

	baggage := map[string]string{
		"key": "value",
	}

	workflow, err := NewFifoWorkflowRecord(tableName, partitionKey, nil, createdAt, startAt, targetQueueUrl, event, eventGroupId, baggage)
	assert.Error(t, err)
	assert.Nil(t, workflow)
	assert.Equal(t, ErrFifoWorkflowQueueMismatch(targetQueueUrl), err)
}

func Test_new_fifo_workflowRecord_should_be_stored_in_correct_form(t *testing.T) {
	var err error
	ctx := context.TODO()

	tableName := uuid.New().String()
	partitionKey := uuid.New().String()
	createdAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 15, 12, 45, 14, 0, time.UTC))
	startAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 16, 12, 45, 14, 0, time.UTC))
	targetQueueUrl := uuid.New().String() + ".fifo"

	event := fmt.Sprintf(`{"partitionKey":"%s","sortKey":null}`, partitionKey)
	eventGroupId := uuid.New().String()

	eventId := fmt.Sprintf("%s#%s", tableName, partitionKey)

	baggage := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	workflow, err := NewFifoWorkflowRecord(tableName, partitionKey, nil, createdAt, startAt, targetQueueUrl, event, eventGroupId, baggage)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)

	actualItems, err := attributevalue.MarshalMap(*workflow)
	assert.NoError(t, err)
	expectedItems := map[string]types.AttributeValue{
		"event_id": &types.AttributeValueMemberS{
			Value: eventId,
		},
		"created_at": &types.AttributeValueMemberS{
			Value: "2023-10-15T12:45:14Z",
		},
		"start_at": &types.AttributeValueMemberS{
			Value: "2023-10-16T12:45:14Z",
		},
		"amount_of_starts": &types.AttributeValueMemberN{
			Value: "0",
		},
		"target_queue_url": &types.AttributeValueMemberS{
			Value: targetQueueUrl,
		},
		"is_open": &types.AttributeValueMemberS{
			Value: string(Open),
		},
		"event": &types.AttributeValueMemberS{
			Value: event,
		},
		"event_message_group_id": &types.AttributeValueMemberS{
			Value: eventGroupId,
		},
		"baggage": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"key1": &types.AttributeValueMemberS{
					Value: "value1",
				},
				"key2": &types.AttributeValueMemberS{
					Value: "value2",
				},
			},
		},
	}

	assert.Equal(t, expectedItems, actualItems)

	err = workflowRecordTable.Action(dynamodbClient).Persist(ctx, *workflow)
	assert.NoError(t, err)

	actualWorkflow := WorkflowRecord{
		EventId:        eventId,
		TargetQueueUrl: targetQueueUrl,
	}
	err = workflowRecordTable.Action(dynamodbClient).Reconstitute(ctx, &actualWorkflow)

	assert.NoError(t, err)
	assert.NotNil(t, actualWorkflow)

	expectedWorkflow := *workflow
	assert.Equal(t, expectedWorkflow, actualWorkflow)
}

func Test_new_fifo_workflowRecord_with_empty_baggage_should_be_stored_in_correct_form(t *testing.T) {
	var err error
	ctx := context.TODO()

	tableName := uuid.New().String()
	partitionKey := uuid.New().String()
	createdAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 15, 12, 45, 14, 0, time.UTC))
	startAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 16, 12, 45, 14, 0, time.UTC))
	targetQueueUrl := uuid.New().String() + ".fifo"

	event := fmt.Sprintf(`{"partitionKey":"%s","sortKey":null}`, partitionKey)
	eventGroupId := uuid.New().String()

	eventId := fmt.Sprintf("%s#%s", tableName, partitionKey)

	baggage := map[string]string{}

	workflow, err := NewFifoWorkflowRecord(tableName, partitionKey, nil, createdAt, startAt, targetQueueUrl, event, eventGroupId, baggage)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)

	actualItems, err := attributevalue.MarshalMap(*workflow)
	assert.NoError(t, err)
	expectedItems := map[string]types.AttributeValue{
		"event_id": &types.AttributeValueMemberS{
			Value: eventId,
		},
		"created_at": &types.AttributeValueMemberS{
			Value: "2023-10-15T12:45:14Z",
		},
		"start_at": &types.AttributeValueMemberS{
			Value: "2023-10-16T12:45:14Z",
		},
		"amount_of_starts": &types.AttributeValueMemberN{
			Value: "0",
		},
		"target_queue_url": &types.AttributeValueMemberS{
			Value: targetQueueUrl,
		},
		"is_open": &types.AttributeValueMemberS{
			Value: string(Open),
		},
		"event": &types.AttributeValueMemberS{
			Value: event,
		},
		"event_message_group_id": &types.AttributeValueMemberS{
			Value: eventGroupId,
		},
		"baggage": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{},
		},
	}

	assert.Equal(t, expectedItems, actualItems)

	err = workflowRecordTable.Action(dynamodbClient).Persist(ctx, *workflow)
	assert.NoError(t, err)

	actualWorkflow := WorkflowRecord{
		EventId:        eventId,
		TargetQueueUrl: targetQueueUrl,
	}
	err = workflowRecordTable.Action(dynamodbClient).Reconstitute(ctx, &actualWorkflow)

	assert.NoError(t, err)
	assert.NotNil(t, actualWorkflow)

	expectedWorkflow := *workflow
	assert.Equal(t, expectedWorkflow, actualWorkflow)
}

func Test_Closed_WorkflowRecord_should_be_stored_in_correct(t *testing.T) {

	ctx := context.TODO()

	tableName := uuid.New().String()
	partitionKey := uuid.New().String()
	sortKey := uuid.New().String()
	createdAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 15, 12, 45, 14, 0, time.UTC))
	startAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 16, 12, 45, 14, 0, time.UTC))
	targetQueueUrl := uuid.New().String() + ".fifo"
	eventGroupId := uuid.New().String()

	event := fmt.Sprintf(`{"partitionKey":"%s","sortKey":"%s"}`, partitionKey, sortKey)

	baggage := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	workflow, err := NewFifoWorkflowRecord(tableName, partitionKey, &sortKey, createdAt, startAt, targetQueueUrl, event, eventGroupId, baggage)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)

	workflow.IsOpen = nil
	amountOfStarts := 1
	workflow.AmountOfStarts = amountOfStarts
	finishedAt := zulu.DateTimeFromTime(time.Date(2023, time.October, 17, 12, 45, 14, 0, time.UTC))
	workflow.FinishedAt = &finishedAt

	eventId := fmt.Sprintf("%s#%s#%s", tableName, partitionKey, sortKey)

	actualItems, err := attributevalue.MarshalMap(*workflow)
	assert.NoError(t, err)
	expectedItems := map[string]types.AttributeValue{
		"event_id": &types.AttributeValueMemberS{
			Value: eventId,
		},
		"created_at": &types.AttributeValueMemberS{
			Value: "2023-10-15T12:45:14Z",
		},
		"start_at": &types.AttributeValueMemberS{
			Value: "2023-10-16T12:45:14Z",
		},
		"amount_of_starts": &types.AttributeValueMemberN{
			Value: "1",
		},
		"target_queue_url": &types.AttributeValueMemberS{
			Value: targetQueueUrl,
		},
		"finished_at": &types.AttributeValueMemberS{
			Value: "2023-10-17T12:45:14Z",
		},
		"event": &types.AttributeValueMemberS{
			Value: event,
		},
		"event_message_group_id": &types.AttributeValueMemberS{
			Value: eventGroupId,
		},
		"baggage": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"key1": &types.AttributeValueMemberS{
					Value: "value1",
				},
				"key2": &types.AttributeValueMemberS{
					Value: "value2",
				},
			},
		},
	}

	assert.Equal(t, expectedItems, actualItems)

	err = workflowRecordTable.Action(dynamodbClient).Persist(ctx, *workflow)
	assert.NoError(t, err)

	actualWorkflow := WorkflowRecord{
		EventId:        eventId,
		TargetQueueUrl: targetQueueUrl,
	}
	err = workflowRecordTable.Action(dynamodbClient).Reconstitute(ctx, &actualWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualWorkflow)

	expectedWorkflow := *workflow
	assert.Equal(t, expectedWorkflow, actualWorkflow)
}
