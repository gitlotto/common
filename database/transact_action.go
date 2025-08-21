package database

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (table Table[R]) TransactInsert(
	record R,
) (item types.TransactWriteItem, err error) {
	items, err := attributevalue.MarshalMap(record)
	if err != nil {
		return
	}

	condition := fmt.Sprintf("attribute_not_exists(%s)", table.PartitionKey)
	if table.SortKey != nil {
		condition = fmt.Sprintf("%s AND attribute_not_exists(%s)", condition, *table.SortKey)
	}

	item = types.TransactWriteItem{
		Put: &types.Put{
			TableName:           aws.String(table.Name),
			Item:                items,
			ConditionExpression: aws.String(condition),
		},
	}

	return
}

func (table Table[R]) TransactUpsert(
	record R,
) (item types.TransactWriteItem, err error) {
	items, err := attributevalue.MarshalMap(record)
	if err != nil {
		return
	}

	item = types.TransactWriteItem{
		Put: &types.Put{
			TableName: aws.String(table.Name),
			Item:      items,
		},
	}

	return
}

func (table Table[R]) TransactDelete(
	record R,
) (item types.TransactWriteItem, err error) {
	keys, err := table.PrimaryKey(record)
	if err != nil {
		return
	}

	item = types.TransactWriteItem{
		Delete: &types.Delete{
			TableName: aws.String(table.Name),
			Key:       keys,
		},
	}

	return
}
