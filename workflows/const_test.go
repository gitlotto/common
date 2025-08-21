package workflows

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gitlotto/common/database"
)

const workflowsTableName = "workflows-workflows"
const openWorkflowsIndexName = "workflows-openWorkflows"

var dynamodbClient *dynamodb.Client

func setup() bool {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		panic(err)
	}
	dynamodbClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
	return true
}

var _ = setup()

type Event struct {
	PartitionKey string  `json:"partitionKey"`
	SortKey      *string `json:"sortKey,omitempty"`
}

var workflowRecordTable = WorkflowRecordTable{
	Table: database.Table[WorkflowRecord]{
		Name:         workflowsTableName,
		PartitionKey: "event_id",
		SortKey:      aws.String("target_queue_url"),
	},
	DynamodbClient: dynamodbClient,
}

var openWorkflowsIndex = OpenWorkflowsIndex{
	TableName:      workflowsTableName,
	IndexName:      openWorkflowsIndexName,
	DynamodbClient: dynamodbClient,
}

func deleteAllWorkflows() (err error) {
	ctx := context.TODO()
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(workflowsTableName),
	}
	workflows, err := dynamodbClient.Scan(ctx, scanInput)
	if err != nil {
		return
	}

	for _, item := range workflows.Items {
		deleteInput := &dynamodb.DeleteItemInput{
			TableName: aws.String(workflowsTableName),
			Key: map[string]types.AttributeValue{
				"event_id":         item["event_id"],
				"target_queue_url": item["target_queue_url"],
			},
		}
		_, err = dynamodbClient.DeleteItem(ctx, deleteInput)
		if err != nil {
			return
		}
	}

	return
}
