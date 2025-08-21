package database

import (
	"context"
	"testing"

	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_Transaction_insert_should_put_a_simple_record_into_the_database(t *testing.T) {
	var err error
	ctx := context.TODO()

	simpleRecord1 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 1",
	}

	err = NewTransaction().
		Include(simpleRecordsTable.TransactInsert(simpleRecord1)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord1 := simpleRecord{ThePartitionKey: simpleRecord1.ThePartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualSimpleRecord1)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord1, actualSimpleRecord1)

}

func Test_Transaction_insert_should_put_a_composite_record_into_the_database(t *testing.T) {
	var err error
	ctx := context.TODO()
	compositeRecord1 := compositeRecord{
		ThePartitionKey: uuid.NewString(),
		TheSortKey:      rand.Int(),
		SomeValue:       "some value",
	}

	err = NewTransaction().
		Include(compositeRecordsTable.TransactInsert(compositeRecord1)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord1 := compositeRecord{
		ThePartitionKey: compositeRecord1.ThePartitionKey,
		TheSortKey:      compositeRecord1.TheSortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualCompositeRecord1)
	assert.NoError(t, err)
	assert.Equal(t, compositeRecord1, actualCompositeRecord1)
}

func Test_Transaction_insert_should_not_put_a_simple_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error
	ctx := context.TODO()
	simpleRecord1 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 1",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(ctx, simpleRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(simpleRecordsTable.TransactInsert(simpleRecord1)).
		Execute(ctx, dynamodbClient)
	assert.ErrorIs(t, err, ErrConditionalCheckFailed)
}

func Test_Transaction_insert_should_not_put_a_composite_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error
	ctx := context.TODO()
	compositeRecord1 := compositeRecord{
		ThePartitionKey: uuid.NewString(),
		TheSortKey:      rand.Int(),
		SomeValue:       "some value",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, compositeRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(compositeRecordsTable.TransactInsert(compositeRecord1)).
		Execute(ctx, dynamodbClient)
	assert.ErrorIs(t, err, ErrConditionalCheckFailed)

}

func Test_Transaction_upsert_should_put_a_simple_record_into_the_database(t *testing.T) {
	var err error
	ctx := context.TODO()
	simpleRecord1 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 1",
	}

	err = NewTransaction().
		Include(simpleRecordsTable.TransactUpsert(simpleRecord1)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord1 := simpleRecord{ThePartitionKey: simpleRecord1.ThePartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualSimpleRecord1)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord1, actualSimpleRecord1)

}

func Test_Transaction_upsert_should_put_a_composite_record_into_the_database(t *testing.T) {
	var err error
	ctx := context.TODO()
	compositeRecord1 := compositeRecord{
		ThePartitionKey: uuid.NewString(),
		TheSortKey:      rand.Int(),
		SomeValue:       "some value",
	}

	err = NewTransaction().
		Include(compositeRecordsTable.TransactUpsert(compositeRecord1)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord1 := compositeRecord{
		ThePartitionKey: compositeRecord1.ThePartitionKey,
		TheSortKey:      compositeRecord1.TheSortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualCompositeRecord1)
	assert.NoError(t, err)
	assert.Equal(t, compositeRecord1, actualCompositeRecord1)
}

func Test_Transaction_upsert_should_upsert_a_simple_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error
	ctx := context.TODO()
	simpleRecord1 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 1",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(ctx, simpleRecord1)
	assert.NoError(t, err)

	simpleRecord2 := simpleRecord{
		ThePartitionKey: simpleRecord1.ThePartitionKey,
		SomeValue:       "some value 2",
	}

	err = NewTransaction().
		Include(simpleRecordsTable.TransactUpsert(simpleRecord2)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord2 := simpleRecord{ThePartitionKey: simpleRecord2.ThePartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualSimpleRecord2)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord2, actualSimpleRecord2)
}

func Test_Transaction_upsert_should_upsert_a_composite_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error
	ctx := context.TODO()
	compositeRecord1 := compositeRecord{
		ThePartitionKey: uuid.NewString(),
		TheSortKey:      rand.Int(),
		SomeValue:       "some value",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, compositeRecord1)
	assert.NoError(t, err)

	compositeRecord2 := compositeRecord{
		ThePartitionKey: compositeRecord1.ThePartitionKey,
		TheSortKey:      compositeRecord1.TheSortKey,
		SomeValue:       "some value 2",
	}

	err = NewTransaction().
		Include(compositeRecordsTable.TransactUpsert(compositeRecord2)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord2 := compositeRecord{
		ThePartitionKey: compositeRecord2.ThePartitionKey,
		TheSortKey:      compositeRecord2.TheSortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualCompositeRecord2)
	assert.NoError(t, err)
	assert.Equal(t, compositeRecord2, actualCompositeRecord2)
}

func Test_Transaction_delete_should_delete_a_simple_record_from_the_database(t *testing.T) {
	var err error
	ctx := context.TODO()
	simpleRecord1 := simpleRecord{
		ThePartitionKey: uuid.NewString(),
		SomeValue:       "some value 1",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(ctx, simpleRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(simpleRecordsTable.TransactDelete(simpleRecord1)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord1 := simpleRecord{ThePartitionKey: simpleRecord1.ThePartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualSimpleRecord1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func Test_Transaction_delete_should_delete_a_composite_record_from_the_database(t *testing.T) {
	var err error
	ctx := context.TODO()
	compositeRecord1 := compositeRecord{
		ThePartitionKey: uuid.NewString(),
		TheSortKey:      rand.Int(),
		SomeValue:       "some value",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, compositeRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(compositeRecordsTable.TransactDelete(compositeRecord1)).
		Execute(ctx, dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord1 := compositeRecord{
		ThePartitionKey: compositeRecord1.ThePartitionKey,
		TheSortKey:      compositeRecord1.TheSortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &actualCompositeRecord1)
	assert.ErrorIs(t, err, ErrNotFound)
}
