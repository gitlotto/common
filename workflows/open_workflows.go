package workflows

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gitlotto/common/zulu"
)

type OpenWorkflowsIndex struct {
	TableName      string
	IndexName      string
	DynamodbClient *dynamodb.Client
}

func (index OpenWorkflowsIndex) OpenWorkflows(ctx context.Context, limit int, until zulu.DateTime) (workflowRecords []WorkflowRecord, err error) {

	queryInput := &dynamodb.QueryInput{
		TableName:              &index.TableName,
		IndexName:              &index.IndexName,
		KeyConditionExpression: aws.String("is_open = :is_open AND start_at <= :start_at"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":is_open": &types.AttributeValueMemberS{
				Value: string(Open),
			},
			":start_at": &types.AttributeValueMemberS{
				Value: until.String(),
			},
		},
		ScanIndexForward: aws.Bool(true),
		Limit:            aws.Int32(int32(limit)),
	}

	items, err := index.DynamodbClient.Query(ctx, queryInput)
	if err != nil {
		return
	}

	err = attributevalue.UnmarshalListOfMaps(items.Items, &workflowRecords)

	return
}
