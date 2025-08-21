package workflows

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gitlotto/common/database"
	"github.com/gitlotto/common/zulu"
)

type WorkflowRecordTable struct {
	database.Table[WorkflowRecord]
	DynamodbClient *dynamodb.Client
}

var ErrWorkflowHadBeenFinished = errors.New("workflow had been finished")

func (table WorkflowRecordTable) Postpone(ctx context.Context, workflow WorkflowRecord, nextStartAt zulu.DateTime) (err error) {

	keys, err := table.Table.PrimaryKey(workflow)
	if err != nil {
		return
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName:        aws.String(table.Table.Name),
		Key:              keys,
		UpdateExpression: aws.String("SET start_at = :start_at ADD amount_of_starts :by_one"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":start_at": &types.AttributeValueMemberS{
				Value: nextStartAt.String(),
			},
			":by_one": &types.AttributeValueMemberN{
				Value: "1",
			},
		},
		ConditionExpression: aws.String("attribute_exists(is_open)"),
	}
	_, err = table.DynamodbClient.UpdateItem(ctx, updateInput)

	var conditionalCheckFailedException *types.ConditionalCheckFailedException
	if errors.As(err, &conditionalCheckFailedException) {
		err = ErrWorkflowHadBeenFinished
		return
	}

	var transactionCanceledException *types.TransactionCanceledException
	if errors.As(err, &transactionCanceledException) {
		for _, reason := range transactionCanceledException.CancellationReasons {
			if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
				err = ErrWorkflowHadBeenFinished
				return
			}
		}
	}

	return err
}

func (table WorkflowRecordTable) TransactionalClose(
	worflowEventId string,
	workflowTargetQueueUrl string,
	finishedAt zulu.DateTime,
) (item types.TransactWriteItem, err error) {
	workflow := WorkflowRecord{
		EventId:        worflowEventId,
		TargetQueueUrl: workflowTargetQueueUrl,
	}
	keys, err := table.Table.PrimaryKey(workflow)
	if err != nil {
		return
	}
	update := &types.Update{
		TableName: aws.String(table.Table.Name),
		Key:       keys,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":finished_at": &types.AttributeValueMemberS{
				Value: finishedAt.String(),
			},
		},
		UpdateExpression:    aws.String("SET finished_at = :finished_at REMOVE is_open"),
		ConditionExpression: aws.String("attribute_exists(is_open)"),
	}
	item = types.TransactWriteItem{
		Update: update,
	}
	return
}

func (table WorkflowRecordTable) Close(ctx context.Context, worflowEventId string, workflowTargetQueueUrl string, finishedAt zulu.DateTime) (err error) {
	workflow := WorkflowRecord{
		EventId:        worflowEventId,
		TargetQueueUrl: workflowTargetQueueUrl,
	}
	keys, err := table.Table.PrimaryKey(workflow)
	if err != nil {
		return
	}
	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(table.Table.Name),
		Key:       keys,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":finished_at": &types.AttributeValueMemberS{
				Value: finishedAt.String(),
			},
		},
		UpdateExpression:    aws.String("SET finished_at = :finished_at REMOVE is_open"),
		ConditionExpression: aws.String("attribute_exists(is_open)"),
	}
	_, err = table.DynamodbClient.UpdateItem(ctx, updateInput)

	var conditionalCheckFailedException *types.ConditionalCheckFailedException
	if errors.As(err, &conditionalCheckFailedException) {
		err = ErrWorkflowHadBeenFinished
		return
	}

	var transactionCanceledException *types.TransactionCanceledException
	if errors.As(err, &transactionCanceledException) {
		for _, reason := range transactionCanceledException.CancellationReasons {
			if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
				err = ErrWorkflowHadBeenFinished
				return
			}
		}
	}

	return err
}
