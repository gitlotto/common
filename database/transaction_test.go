package database

import (
	"context"
	"testing"

	"math/rand/v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type simpleRecordTransactions struct {
	table Table[simpleRecord]
}

func (table simpleRecordTransactions) write(r simpleRecord) (transaction types.TransactWriteItem, err error) {
	item, err := attributevalue.MarshalMap(r)
	candidate := types.TransactWriteItem{
		Put: &types.Put{
			Item:      item,
			TableName: aws.String(table.table.Name),
		},
	}
	transaction = candidate
	return
}

func (table simpleRecordTransactions) writeIfDoesNotExist(r simpleRecord) (transaction types.TransactWriteItem, err error) {
	item, err := attributevalue.MarshalMap(r)
	candidate := types.TransactWriteItem{
		Put: &types.Put{
			Item:                item,
			TableName:           aws.String(table.table.Name),
			ConditionExpression: aws.String("attribute_not_exists(partition_key)"),
		},
	}
	transaction = candidate
	return
}

type compositeReordsTransactions struct {
	table Table[compositeRecord]
}

func (table compositeReordsTransactions) write(record compositeRecord) (transaction types.TransactWriteItem, err error) {
	item, err := attributevalue.MarshalMap(record)
	candidate := types.TransactWriteItem{
		Put: &types.Put{
			Item:      item,
			TableName: aws.String(table.table.Name),
		},
	}
	transaction = candidate
	return
}

var simpleRecordTransactionsImpl = simpleRecordTransactions{
	table: simpleRecordsTable,
}

var compositeReordTransactionsImpl = compositeReordsTransactions{
	table: compositeRecordsTable,
}

func Test_Transaction_should_execute(t *testing.T) {
	var err error
	ctx := context.TODO()

	simpleRecord1 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 1",
	}

	simpleRecord2 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 2",
	}

	compositeRecord1 := compositeRecord{
		ThePartitionKey: uuid.NewString(),
		TheSortKey:      rand.Int(),
		SomeValue:       "some value",
	}

	err = NewTransaction().
		Include(simpleRecordTransactionsImpl.write(simpleRecord1)).
		Include(simpleRecordTransactionsImpl.write(simpleRecord2)).
		Include(compositeReordTransactionsImpl.write(compositeRecord1)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord1 := simpleRecord{ThePartitionKey: simpleRecord1.ThePartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualSimpleRecord1)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord1, actualSimpleRecord1)

	actualSimpleRecord2 := simpleRecord{ThePartitionKey: simpleRecord2.ThePartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualSimpleRecord2)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord2, actualSimpleRecord2)

	actualCompositeRecord1 := compositeRecord{
		ThePartitionKey: compositeRecord1.ThePartitionKey,
		TheSortKey:      compositeRecord1.TheSortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualCompositeRecord1)
	assert.NoError(t, err)
	assert.Equal(t, compositeRecord1, actualCompositeRecord1)
}

func Test_Transaction_should_handle_the_error_upon_the_failure_of_the_condition(t *testing.T) {
	var err error
	ctx := context.TODO()

	simpleRecord1 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 1",
	}

	simpleRecord2 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 2",
	}

	compositeRecord1 := compositeRecord{
		ThePartitionKey: uuid.NewString(),
		TheSortKey:      rand.Int(),
		SomeValue:       "some value",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(ctx, simpleRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(simpleRecordTransactionsImpl.writeIfDoesNotExist(simpleRecord1)).
		Include(simpleRecordTransactionsImpl.write(simpleRecord2)).
		Include(compositeReordTransactionsImpl.write(compositeRecord1)).
		Execute(ctx, dynamodbClient)
	assert.ErrorIs(t, err, ErrConditionalCheckFailed)

	actualSimpleRecord2 := simpleRecord{ThePartitionKey: simpleRecord2.ThePartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualSimpleRecord2)
	assert.ErrorIs(t, err, ErrNotFound)

	actualCompositeRecord1 := compositeRecord{
		ThePartitionKey: compositeRecord1.ThePartitionKey,
		TheSortKey:      compositeRecord1.TheSortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualCompositeRecord1)
	assert.ErrorIs(t, err, ErrNotFound)
}
