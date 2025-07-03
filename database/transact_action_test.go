package database

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func Test_Transaction_insert_should_put_a_simple_record_into_the_database(t *testing.T) {
	var err error

	simpleRecord1 := simpleRecord{
		PartitionKey: uuid.New().String(),
		SomeValue:    "some value 1",
	}

	err = NewTransaction().
		Include(simpleRecordsTable.TransactInsert(simpleRecord1)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord1 := simpleRecord{PartitionKey: simpleRecord1.PartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(&actualSimpleRecord1)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord1, actualSimpleRecord1)

}

func Test_Transaction_insert_should_put_a_composite_record_into_the_database(t *testing.T) {
	var err error

	compositeRecord1 := compositeRecord{
		PartitionKey: uuid.New().String(),
		SortKey:      rand.Int(),
		SomeValue:    "some value",
	}

	err = NewTransaction().
		Include(compositeRecordsTable.TransactInsert(compositeRecord1)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord1 := compositeRecord{
		PartitionKey: compositeRecord1.PartitionKey,
		SortKey:      compositeRecord1.SortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(&actualCompositeRecord1)
	assert.NoError(t, err)
	assert.Equal(t, compositeRecord1, actualCompositeRecord1)
}

func Test_Transaction_insert_should_not_put_a_simple_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error

	simpleRecord1 := simpleRecord{
		PartitionKey: uuid.New().String(),
		SomeValue:    "some value 1",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(simpleRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(simpleRecordsTable.TransactInsert(simpleRecord1)).
		Execute(dynamodbClient)
	assert.ErrorIs(t, err, ErrConditionalCheckFailed)
}

func Test_Transaction_insert_should_not_put_a_composite_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error

	compositeRecord1 := compositeRecord{
		PartitionKey: uuid.New().String(),
		SortKey:      rand.Int(),
		SomeValue:    "some value",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(compositeRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(compositeRecordsTable.TransactInsert(compositeRecord1)).
		Execute(dynamodbClient)
	assert.ErrorIs(t, err, ErrConditionalCheckFailed)

}

func Test_Transaction_upsert_should_put_a_simple_record_into_the_database(t *testing.T) {
	var err error

	simpleRecord1 := simpleRecord{
		PartitionKey: uuid.New().String(),
		SomeValue:    "some value 1",
	}

	err = NewTransaction().
		Include(simpleRecordsTable.TransactUpsert(simpleRecord1)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord1 := simpleRecord{PartitionKey: simpleRecord1.PartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(&actualSimpleRecord1)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord1, actualSimpleRecord1)

}

func Test_Transaction_upsert_should_put_a_composite_record_into_the_database(t *testing.T) {
	var err error

	compositeRecord1 := compositeRecord{
		PartitionKey: uuid.New().String(),
		SortKey:      rand.Int(),
		SomeValue:    "some value",
	}

	err = NewTransaction().
		Include(compositeRecordsTable.TransactUpsert(compositeRecord1)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord1 := compositeRecord{
		PartitionKey: compositeRecord1.PartitionKey,
		SortKey:      compositeRecord1.SortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(&actualCompositeRecord1)
	assert.NoError(t, err)
	assert.Equal(t, compositeRecord1, actualCompositeRecord1)
}

func Test_Transaction_upsert_should_upsert_a_simple_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error

	simpleRecord1 := simpleRecord{
		PartitionKey: uuid.New().String(),
		SomeValue:    "some value 1",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(simpleRecord1)
	assert.NoError(t, err)

	simpleRecord2 := simpleRecord{
		PartitionKey: simpleRecord1.PartitionKey,
		SomeValue:    "some value 2",
	}

	err = NewTransaction().
		Include(simpleRecordsTable.TransactUpsert(simpleRecord2)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord2 := simpleRecord{PartitionKey: simpleRecord2.PartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(&actualSimpleRecord2)
	assert.NoError(t, err)
	assert.Equal(t, simpleRecord2, actualSimpleRecord2)
}

func Test_Transaction_upsert_should_upsert_a_composite_record_into_the_database_if_it_has_been_saved_before(t *testing.T) {
	var err error

	compositeRecord1 := compositeRecord{
		PartitionKey: uuid.New().String(),
		SortKey:      rand.Int(),
		SomeValue:    "some value",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(compositeRecord1)
	assert.NoError(t, err)

	compositeRecord2 := compositeRecord{
		PartitionKey: compositeRecord1.PartitionKey,
		SortKey:      compositeRecord1.SortKey,
		SomeValue:    "some value 2",
	}

	err = NewTransaction().
		Include(compositeRecordsTable.TransactUpsert(compositeRecord2)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord2 := compositeRecord{
		PartitionKey: compositeRecord2.PartitionKey,
		SortKey:      compositeRecord2.SortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(&actualCompositeRecord2)
	assert.NoError(t, err)
	assert.Equal(t, compositeRecord2, actualCompositeRecord2)
}

func Test_Transaction_delete_should_delete_a_simple_record_from_the_database(t *testing.T) {
	var err error

	simpleRecord1 := simpleRecord{
		PartitionKey: uuid.New().String(),
		SomeValue:    "some value 1",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(simpleRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(simpleRecordsTable.TransactDelete(simpleRecord1)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualSimpleRecord1 := simpleRecord{PartitionKey: simpleRecord1.PartitionKey}
	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(&actualSimpleRecord1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func Test_Transaction_delete_should_delete_a_composite_record_from_the_database(t *testing.T) {
	var err error

	compositeRecord1 := compositeRecord{
		PartitionKey: uuid.New().String(),
		SortKey:      rand.Int(),
		SomeValue:    "some value",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(compositeRecord1)
	assert.NoError(t, err)

	err = NewTransaction().
		Include(compositeRecordsTable.TransactDelete(compositeRecord1)).
		Execute(dynamodbClient)
	assert.NoError(t, err)

	actualCompositeRecord1 := compositeRecord{
		PartitionKey: compositeRecord1.PartitionKey,
		SortKey:      compositeRecord1.SortKey,
	}
	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(&actualCompositeRecord1)
	assert.ErrorIs(t, err, ErrNotFound)
}
