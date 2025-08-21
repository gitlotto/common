package database

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

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

const simpleRecordsTableName = "database.simpleRecords"

type simpleRecord struct {
	ThePartitionKey string `dynamodbav:"partition_key"`
	SomeValue       string `dynamodbav:"some_value"`
}

func (record simpleRecord) PartitionKey() types.AttributeValue {
	return &types.AttributeValueMemberS{Value: record.ThePartitionKey}
}

func (record simpleRecord) SortKey() *types.AttributeValue {
	return nil
}

var simpleRecordsTable = Table[simpleRecord]{
	Name:         simpleRecordsTableName,
	PartitionKey: "partition_key",
	SortKey:      nil,
}

const compositeRecordsTableName = "database.compositeRecords"

type compositeRecord struct {
	ThePartitionKey string `dynamodbav:"partition_key"`
	TheSortKey      int    `dynamodbav:"sort_key"`
	SomeValue       string `dynamodbav:"some_value"`
}

func (record compositeRecord) PartitionKey() types.AttributeValue {
	return &types.AttributeValueMemberS{Value: record.ThePartitionKey}
}

func (record compositeRecord) SortKey() *types.AttributeValue {
	var attValue types.AttributeValue = &types.AttributeValueMemberN{Value: strconv.Itoa(record.TheSortKey)}
	return &attValue
}

var compositeRecordsTable = Table[compositeRecord]{
	Name:         compositeRecordsTableName,
	PartitionKey: "partition_key",
	SortKey:      aws.String("sort_key"),
}
