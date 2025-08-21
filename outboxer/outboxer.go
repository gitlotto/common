package outboxer

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gitlotto/common/database"
	"github.com/gitlotto/common/notification"
	"github.com/gitlotto/common/workflows"
	"github.com/gitlotto/common/zulu"
	"go.uber.org/zap"
)

type Outboxer struct {
	workflowsTableName        string
	openWorkflowsIndexName    string
	notificationTopicArn      string
	amountOfWorkflowsToOutbox int
	nextStartIn               time.Duration
	dynamodbClient            *dynamodb.Client
	sqsClient                 *sqs.Client
	postman                   notification.Postman
	logger                    *zap.Logger
}

func (outboxer *Outboxer) Outbox(ctx context.Context, requestId string) (err error) {

	logger := outboxer.logger
	defer logger.Sync()

	defer func() {
		if err != nil {
			outboxer.postman.SendNotification(ctx, requestId, err.Error())
		}
	}()

	logger = logger.With(zap.String("requestId", requestId))
	logger.Info("outboxing workflows ...")

	workflowsTable := workflows.WorkflowRecordTable{
		Table: database.Table[workflows.WorkflowRecord]{
			Name:         outboxer.workflowsTableName,
			PartitionKey: "event_id",
			SortKey:      aws.String("target_queue_url"),
		},
		DynamodbClient: outboxer.dynamodbClient,
	}

	openWorkflowIndex := workflows.OpenWorkflowsIndex{
		TableName:      outboxer.workflowsTableName,
		IndexName:      outboxer.openWorkflowsIndexName,
		DynamodbClient: outboxer.dynamodbClient,
	}

	now := time.Now()

	workflowRecords, err := openWorkflowIndex.OpenWorkflows(ctx, outboxer.amountOfWorkflowsToOutbox, zulu.DateTimeFromTime(now))

	logger = logger.With(zap.Int("amountOfWorkflows", len(workflowRecords)))
	logger.Info("fetched open workflows")

	if err != nil {
		logger.Error("impossible to fetch open workflows", zap.Error(err))
	}

	var errorsFromEventSending []error

	defer func() {
		if len(errorsFromEventSending) > 0 {
			outboxer.postman.SendNotification(ctx, requestId, fmt.Sprintf("impossible to send %d events", len(errorsFromEventSending)))
		}
	}()

	for _, workflowRecord := range workflowRecords {
		logger = logger.With(zap.String("eventId", workflowRecord.EventId))
		logger = logger.With(zap.String("targetQueueUrl", workflowRecord.TargetQueueUrl))
		logger.Info("sending event ...")
		sendMessageInput := &sqs.SendMessageInput{
			MessageBody:            aws.String(workflowRecord.Event),
			QueueUrl:               aws.String(workflowRecord.TargetQueueUrl),
			MessageGroupId:         aws.String(workflowRecord.EventMessageGroupId),
			MessageDeduplicationId: aws.String(workflowRecord.EventMessageDeduplicationId()),
			MessageAttributes: map[string]types.MessageAttributeValue{
				"EventId": {
					DataType:    aws.String("String"),
					StringValue: aws.String(workflowRecord.EventId),
				},
				"TargetQueueUrl": {
					DataType:    aws.String("String"),
					StringValue: aws.String(workflowRecord.TargetQueueUrl),
				},
			},
		}
		_, errFromEventSending := outboxer.sqsClient.SendMessage(ctx, sendMessageInput)

		if errFromEventSending != nil {
			logger.Error("impossible to send event", zap.Error(errFromEventSending))
			errorsFromEventSending = append(errorsFromEventSending, errFromEventSending)
		}

		logger.Info("event sent. Postponing workflow ...")
		now := time.Now()
		nextStartAt := now.Add(outboxer.nextStartIn)
		err = workflowsTable.Postpone(ctx, workflowRecord, zulu.DateTimeFromTime(nextStartAt))

		if err != nil {
			logger.Error("impossible to postpone workflow", zap.Error(err))
			return
		}
		logger.Info("workflow postponed")

	}

	logger.Info("workflows outboxed")
	return
}
