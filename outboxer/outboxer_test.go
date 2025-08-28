package outboxer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/gitlotto/common/database"
	"github.com/gitlotto/common/logging"
	"github.com/gitlotto/common/notification"
	"github.com/gitlotto/common/queue"
	"github.com/gitlotto/common/workflows"
	"github.com/gitlotto/common/zulu"
)

var dynamodbClient *dynamodb.Client
var sqsClient *sqs.Client
var snsClient *sns.Client

func setup() bool {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		panic(err)
	}
	dynamodbClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
	sqsClient = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
	snsClient = sns.NewFromConfig(cfg, func(o *sns.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
	return true
}

var _ = setup()

var logger = logging.MustCreateZuluTimeLogger()

const sevenHours time.Duration = time.Hour * 7

const workflowsTableName = "outboxer_dynamodb-workflows"
const notificationTopicArn = "arn:aws:sns:us-east-1:000000000000:outboxer_notification-Notifications.fifo"

var outboxer = Outboxer{
	workflowsTableName:        workflowsTableName,
	openWorkflowsIndexName:    "outboxer_dynamodb-openWorkflows",
	notificationTopicArn:      notificationTopicArn,
	amountOfWorkflowsToOutbox: 3,
	nextStartIn:               sevenHours,
	dynamodbClient:            dynamodbClient,
	sqsClient:                 sqsClient,
	postman: notification.Postman{
		SnsClient: snsClient,
		TopicArn:  notificationTopicArn,
	},
	logger: logger,
}

var notificationQueueUrl = "http://localhost:4566/000000000000/outboxer_notification-Notifications.fifo"

var queueOne = "http://localhost:4566/000000000000/outboxer_random_queues-one.fifo"
var queueTwo = "http://localhost:4566/000000000000/outboxer_random_queues-two.fifo"

var workflowsDynamodbTable = database.Table[workflows.WorkflowRecord]{
	Name:         workflowsTableName,
	PartitionKey: "event_id",
	SortKey:      aws.String("target_queue_url"),
}

func Test_Workflow_Outboxer_should_pick_the_oldest_open_workflows_and_issue_events(t *testing.T) {
	var err error

	ctx := context.Background()

	startOfTesting := time.Now()

	err = deleteAllWorkflows(ctx)
	assert.NoError(t, err)

	oldClosedWorkflowStartAt := startOfTesting.Add(-time.Hour * 5)
	oldClosedWorkflow := makeFifoWorkflowRecord(queueOne, oldClosedWorkflowStartAt)
	oldClosedWorkflow.IsOpen = nil
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, oldClosedWorkflow)
	assert.NoError(t, err)

	firstOpenWorkflowStartAt := startOfTesting.Add(-time.Hour * 4)
	firstOpenWorkflow := makeFifoWorkflowRecord(queueOne, firstOpenWorkflowStartAt)
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, firstOpenWorkflow)
	assert.NoError(t, err)

	secondOpenWorkflowStartAt := startOfTesting.Add(-time.Hour * 3)
	secondOpenWorkflow := makeFifoWorkflowRecord(queueTwo, secondOpenWorkflowStartAt)
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, secondOpenWorkflow)
	assert.NoError(t, err)

	thirdOpenWorkflowStartAt := startOfTesting.Add(-time.Hour * 2)
	thirdOpenWorkflow := makeFifoWorkflowRecord(queueOne, thirdOpenWorkflowStartAt)
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, thirdOpenWorkflow)
	assert.NoError(t, err)

	fourthOpenWorkflowStartAt := startOfTesting.Add(-time.Hour * 1)
	fourthOpenWorkflow := makeFifoWorkflowRecord(queueTwo, fourthOpenWorkflowStartAt)
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, fourthOpenWorkflow)
	assert.NoError(t, err)

	requestId := uuid.New().String()
	err = outboxer.Outbox(ctx, requestId)
	assert.NoError(t, err)

	stratOfChecking := time.Now()

	lastNCommandsFromQueueOne, err := queue.GetLastNCommands(ctx, sqsClient, queueOne, 2)
	assert.NoError(t, err)

	actualEventsFromQueueOne := make([]string, 2)
	for i, command := range lastNCommandsFromQueueOne {
		actualEventsFromQueueOne[i] = *command.Body
	}

	expectedEventsFromQueueOne := []string{
		firstOpenWorkflow.Event, thirdOpenWorkflow.Event,
	}

	assert.ElementsMatch(t, expectedEventsFromQueueOne, actualEventsFromQueueOne)

	lastNCommandsFromQueueTwo, err := queue.GetLastNCommands(ctx, sqsClient, queueTwo, 1)
	assert.NoError(t, err)

	actualEventsFromQueueTwo := make([]string, 1)
	for i, command := range lastNCommandsFromQueueTwo {
		actualEventsFromQueueTwo[i] = *command.Body
	}

	expectedEventsFromQueueTwo := []string{
		secondOpenWorkflow.Event,
	}

	assert.ElementsMatch(t, expectedEventsFromQueueTwo, actualEventsFromQueueTwo)
	actualFirstWorkflow := workflows.WorkflowRecord{
		EventId:        firstOpenWorkflow.EventId,
		TargetQueueUrl: queueOne,
	}
	err = workflowsDynamodbTable.Action(dynamodbClient).Reconstitute(ctx, &actualFirstWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualFirstWorkflow)
	assert.WithinRange(t, actualFirstWorkflow.StartAt.ToTime(), startOfTesting.Add(sevenHours-time.Second), stratOfChecking.Add(sevenHours+time.Second))
	expectedAmountOfStartOfFirstWorkflow := firstOpenWorkflow.AmountOfStarts + 1
	assert.Equal(t, expectedAmountOfStartOfFirstWorkflow, actualFirstWorkflow.AmountOfStarts)

	actualSecondWorkflow := workflows.WorkflowRecord{
		EventId:        secondOpenWorkflow.EventId,
		TargetQueueUrl: queueTwo,
	}
	err = workflowsDynamodbTable.Action(dynamodbClient).Reconstitute(ctx, &actualSecondWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualSecondWorkflow)
	assert.WithinRange(t, actualSecondWorkflow.StartAt.ToTime(), startOfTesting.Add(sevenHours-time.Second), stratOfChecking.Add(sevenHours+time.Second))
	expectedAmountOfStartOfSecondWorkflow := secondOpenWorkflow.AmountOfStarts + 1
	assert.Equal(t, expectedAmountOfStartOfSecondWorkflow, actualSecondWorkflow.AmountOfStarts)

	actualThirdWorkflow := workflows.WorkflowRecord{
		EventId:        thirdOpenWorkflow.EventId,
		TargetQueueUrl: queueOne,
	}
	err = workflowsDynamodbTable.Action(dynamodbClient).Reconstitute(ctx, &actualThirdWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualThirdWorkflow)
	assert.WithinRange(t, actualThirdWorkflow.StartAt.ToTime(), startOfTesting.Add(sevenHours-time.Second), stratOfChecking.Add(sevenHours+time.Second))
	expectedAmountOfStartOfThirdWorkflow := thirdOpenWorkflow.AmountOfStarts + 1
	assert.Equal(t, expectedAmountOfStartOfThirdWorkflow, actualThirdWorkflow.AmountOfStarts)

	actualFourthWorkflow := workflows.WorkflowRecord{
		EventId:        fourthOpenWorkflow.EventId,
		TargetQueueUrl: queueTwo,
	}
	err = workflowsDynamodbTable.Action(dynamodbClient).Reconstitute(ctx, &actualFourthWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualFourthWorkflow)
	assert.Equal(t, fourthOpenWorkflow, actualFourthWorkflow)

	actualOldClosedWorkflow := workflows.WorkflowRecord{
		EventId:        oldClosedWorkflow.EventId,
		TargetQueueUrl: queueOne,
	}
	err = workflowsDynamodbTable.Action(dynamodbClient).Reconstitute(ctx, &actualOldClosedWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualOldClosedWorkflow)
	assert.Equal(t, oldClosedWorkflow, actualOldClosedWorkflow)

}

func Test_Workflow_Outboxer_should_pick_the_oldest_open_workflows_and_issue_events_with_baggage(t *testing.T) {
	var err error

	ctx := context.Background()

	startOfTesting := time.Now()

	err = deleteAllWorkflows(ctx)
	assert.NoError(t, err)

	openWorkflowStartAt := startOfTesting.Add(-time.Hour * 4)
	openWorkflow := makeFifoWorkflowRecord(queueOne, openWorkflowStartAt)
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, openWorkflow)
	assert.NoError(t, err)

	requestId := uuid.New().String()
	err = outboxer.Outbox(ctx, requestId)
	assert.NoError(t, err)

	lastNCommandsFromQueueOne, err := queue.GetLastNCommands(ctx, sqsClient, queueOne, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(lastNCommandsFromQueueOne))

	actualMessageFromQueueOne := lastNCommandsFromQueueOne[0]

	// localstack does not return message attributes
	fmt.Println(actualMessageFromQueueOne.MessageAttributes)
	// assert.Equal(t, "value1", *actualMessageFromQueueOne.MessageAttributes["key1"].StringValue)
	// assert.Equal(t, "value2", *actualMessageFromQueueOne.MessageAttributes["key2"].StringValue)

}

func Test_Workflow_Outboxer_should_notify_if_it_fails_to_publish_an_event(t *testing.T) {
	var err error

	ctx := context.Background()

	startOfTesting := time.Now()

	err = deleteAllWorkflows(ctx)
	assert.NoError(t, err)

	firstOpenWorkflowStartAt := startOfTesting.Add(-time.Hour * 4)
	firstOpenWorkflow := makeFifoWorkflowRecord(queueOne, firstOpenWorkflowStartAt)
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, firstOpenWorkflow)
	assert.NoError(t, err)

	secondOpenWorkflowStartAt := startOfTesting.Add(-time.Hour * 3)
	unknownQueue := "http://localhost:4566/000000000000/unknown_queue.fifo"
	secondOpenWorkflow := makeFifoWorkflowRecord(unknownQueue, secondOpenWorkflowStartAt)
	err = workflowsDynamodbTable.Action(dynamodbClient).Persist(ctx, secondOpenWorkflow)
	assert.NoError(t, err)

	requestId := uuid.New().String()
	err = outboxer.Outbox(ctx, requestId)
	assert.NoError(t, err)

	stratOfChecking := time.Now()

	lastNCommandsFromQueueOne, err := queue.GetLastNCommands(ctx, sqsClient, queueOne, 1)
	assert.NoError(t, err)

	actualEventsFromQueueOne := make([]string, 1)
	for i, command := range lastNCommandsFromQueueOne {
		actualEventsFromQueueOne[i] = *command.Body
	}

	expectedEventsFromQueueOne := []string{firstOpenWorkflow.Event}

	assert.ElementsMatch(t, expectedEventsFromQueueOne, actualEventsFromQueueOne)
	actualFirstWorkflow := workflows.WorkflowRecord{
		EventId:        firstOpenWorkflow.EventId,
		TargetQueueUrl: queueOne,
	}
	err = workflowsDynamodbTable.Action(dynamodbClient).Reconstitute(ctx, &actualFirstWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualFirstWorkflow)
	assert.WithinRange(t, actualFirstWorkflow.StartAt.ToTime(), startOfTesting.Add(sevenHours-time.Second), stratOfChecking.Add(sevenHours+time.Second))
	expectedAmountOfStartOfFirstWorkflow := firstOpenWorkflow.AmountOfStarts + 1
	assert.Equal(t, expectedAmountOfStartOfFirstWorkflow, actualFirstWorkflow.AmountOfStarts)

	actualSecondWorkflow := workflows.WorkflowRecord{
		EventId:        secondOpenWorkflow.EventId,
		TargetQueueUrl: unknownQueue,
	}
	err = workflowsDynamodbTable.Action(dynamodbClient).Reconstitute(ctx, &actualSecondWorkflow)
	assert.NoError(t, err)
	assert.NotNil(t, actualSecondWorkflow)
	assert.WithinRange(t, actualSecondWorkflow.StartAt.ToTime(), startOfTesting.Add(sevenHours-time.Second), stratOfChecking.Add(sevenHours+time.Second))
	expectedAmountOfStartOfSecondWorkflow := secondOpenWorkflow.AmountOfStarts + 1
	assert.Equal(t, expectedAmountOfStartOfSecondWorkflow, actualSecondWorkflow.AmountOfStarts)

	lastNCommandsFromNotificationQueue, err := queue.GetLastNCommands(ctx, sqsClient, notificationQueueUrl, 1)
	assert.NoError(t, err)

	expectedNotification := fmt.Sprintf(
		`{"requestId":"%s","message":"impossible to send 1 events"}`,
		requestId,
	)
	actualNotification := lastNCommandsFromNotificationQueue[0].Body

	assert.Equal(t, expectedNotification, *actualNotification)
}

func makeFifoWorkflowRecord(targetQueueUrl string, startAt time.Time) workflows.WorkflowRecord {
	tableName := uuid.New().String()
	partitionKey := uuid.New().String()
	sortKey := uuid.New().String()
	createdAt := zulu.DateTimeFromTime(time.Date(2023, time.September, 16, 12, 45, 14, 0, time.UTC))
	event := uuid.New().String()
	eventGroupId := uuid.New().String()
	baggage := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	spanContextJson := `{"traceId":"01234567890123456789012345678901","spanId":"01234567890123456789012345678901","traceFlags":0,"traceState":{}}`
	workflow, err := workflows.NewFifoWorkflowRecord(tableName, partitionKey, &sortKey, createdAt, zulu.DateTimeFromTime(startAt), targetQueueUrl, event, eventGroupId, baggage, spanContextJson)
	if err != nil {
		panic(err)
	}
	return *workflow
}

func deleteAllWorkflows(ctx context.Context) (err error) {
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(outboxer.workflowsTableName),
	}
	workflows, err := dynamodbClient.Scan(ctx, scanInput)
	if err != nil {
		return
	}

	for _, item := range workflows.Items {
		deleteItemInput := &dynamodb.DeleteItemInput{
			TableName: aws.String(outboxer.workflowsTableName),
			Key: map[string]types.AttributeValue{
				"event_id":         item["event_id"],
				"target_queue_url": item["target_queue_url"],
			},
		}
		_, err = dynamodbClient.DeleteItem(ctx, deleteItemInput)
		if err != nil {
			return
		}
	}

	return
}
