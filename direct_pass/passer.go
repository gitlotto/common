package direct_pass

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gitlotto/common/database"
	"github.com/gitlotto/common/notification"
	"github.com/gitlotto/common/workflows"
	"github.com/gitlotto/common/zulu"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DirectPasser struct {
	workflowsTableName   string
	notificationTopicArn string
	nextStartIn          time.Duration
	dynamodbClient       *dynamodb.Client
	sqsClient            *sqs.Client
	postman              notification.Postman
	logger               *zap.Logger
}

func (passer *DirectPasser) Pass(ctx context.Context, event events.DynamoDBEvent) (err error) {

	logger := passer.logger
	defer logger.Sync()
	requestId := uuid.New().String()

	defer func() {
		if err != nil {
			passer.postman.SendNotification(ctx, requestId, err.Error())
		}
	}()

	logger = logger.With(zap.String("requestId", requestId))
	logger.Info("outboxing workflows ...")

	workflowsTable := workflows.WorkflowRecordTable{
		Table: database.Table[workflows.WorkflowRecord]{
			Name:         passer.workflowsTableName,
			PartitionKey: "event_id",
			SortKey:      aws.String("target_queue_url"),
		},
		DynamodbClient: passer.dynamodbClient,
	}

	processSingle := func(record events.DynamoDBEventRecord, logger *zap.Logger) {
		var err error
		defer func() {
			if err != nil {
				passer.postman.SendNotification(ctx, record.EventID, fmt.Sprintf("impossible to directly pass the event %s to SQS", record.EventID))
			}
		}()

		logger = logger.With(zap.String("dynamodbEventID", record.EventID))
		logger.Info("processing single record ...")

		if record.EventName != "INSERT" {
			logger.Info("skipping non-insert event")
			return
		}

		workflowRecord, err := unmarshalWorkflow(record.Change.NewImage)
		if err != nil {
			logger.Error("impossible to unmarshal workflow record", zap.Error(err))
			return
		}

		now := time.Now()

		if workflowRecord.StartAt.ToTime().After(now) {
			logger.Info("workflow is not ready to be passed")
			return
		}

		logger = logger.With(zap.String("eventId", workflowRecord.EventId))
		logger = logger.With(zap.String("targetQueueUrl", workflowRecord.TargetQueueUrl))
		logger.Info("sending event ...")
		inputMessage := &sqs.SendMessageInput{
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
		_, err = passer.sqsClient.SendMessage(ctx, inputMessage)

		if err != nil {
			logger.Error("impossible to send event", zap.Error(err))
			return
		}

		logger.Info("event sent. Postponing workflow ...")
		nextStartAt := now.Add(passer.nextStartIn)
		err = workflowsTable.Postpone(ctx, workflowRecord, zulu.DateTimeFromTime(nextStartAt))

		if err != nil {
			logger.Error("impossible to postpone workflow", zap.Error(err))
			return
		}
		logger.Info("workflow postponed")
	}

	for _, record := range event.Records {
		processSingle(record, logger)
	}

	logger.Info("workflows outboxed")
	return
}
