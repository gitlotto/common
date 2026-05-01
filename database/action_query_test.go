package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Querying_in_descending_order_should_fetch_composite_records_from_the_beginning_if_no_start_key_is_provided(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKeyValue := "2231bbd9-8247-4115-941a-cf1dd87a6f1a"

	firstRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      1,
		SomeValue:       "some value 1",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, firstRecord)
	assert.NoError(t, err)

	secondRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      2,
		SomeValue:       "some value 2",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, secondRecord)
	assert.NoError(t, err)

	thirdRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      3,
		SomeValue:       "some value 3",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, thirdRecord)
	assert.NoError(t, err)

	fourthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      4,
		SomeValue:       "some value 4",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fourthRecord)
	assert.NoError(t, err)

	fifthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      5,
		SomeValue:       "some value 5",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fifthRecord)
	assert.NoError(t, err)

	limit := 2
	keyRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
	}
	actualRecords, err := compositeRecordsTable.
		Action(dynamodbClient).
		Query(ctx, keyRecord.PartitionKey(), nil, false, limit)
	assert.NoError(t, err)

	assert.Equal(t, actualRecords[1], fourthRecord)
	assert.Equal(t, actualRecords[0], fifthRecord)

	assert.Len(t, actualRecords, 2)
}

func Test_Querying_in_descending_order_should_fetch_composite_records_from_the_given_start_key(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKeyValue := "bc8f6d9b-cc47-46e7-a18b-489b63d8dfc4"

	firstRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      1,
		SomeValue:       "some value 1",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, firstRecord)
	assert.NoError(t, err)

	secondRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      2,
		SomeValue:       "some value 2",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, secondRecord)
	assert.NoError(t, err)

	thirdRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      3,
		SomeValue:       "some value 3",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, thirdRecord)
	assert.NoError(t, err)

	fourthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      4,
		SomeValue:       "some value 4",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fourthRecord)
	assert.NoError(t, err)

	fifthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      5,
		SomeValue:       "some value 5",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fifthRecord)
	assert.NoError(t, err)

	keyRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      4,
	}
	limit := 2
	actualRecords, err := compositeRecordsTable.
		Action(dynamodbClient).
		Query(ctx, keyRecord.PartitionKey(), keyRecord.SortKey(), false, limit)
	assert.NoError(t, err)
	assert.Equal(t, actualRecords[1], secondRecord)
	assert.Equal(t, actualRecords[0], thirdRecord)

	assert.Len(t, actualRecords, 2)
}

func Test_Querying_in_ascending_order_should_fetch_composite_records_from_the_beginning_if_no_start_key_is_provided(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKeyValue := "2231bbd9-8247-4115-941a-cf1dd87a6f1a"

	firstRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      1,
		SomeValue:       "some value 1",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, firstRecord)
	assert.NoError(t, err)

	secondRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      2,
		SomeValue:       "some value 2",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, secondRecord)
	assert.NoError(t, err)

	thirdRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      3,
		SomeValue:       "some value 3",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, thirdRecord)
	assert.NoError(t, err)

	fourthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      4,
		SomeValue:       "some value 4",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fourthRecord)
	assert.NoError(t, err)

	fifthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      5,
		SomeValue:       "some value 5",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fifthRecord)
	assert.NoError(t, err)

	limit := 2
	keyRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
	}
	actualRecords, err := compositeRecordsTable.
		Action(dynamodbClient).
		Query(ctx, keyRecord.PartitionKey(), nil, true, limit)
	assert.NoError(t, err)
	assert.Equal(t, actualRecords[0], firstRecord)
	assert.Equal(t, actualRecords[1], secondRecord)

}

func Test_Querying_in_ascending_order_should_fetch_composite_records_from_the_given_start_key(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKeyValue := "bc8f6d9b-cc47-46e7-a18b-489b63d8dfc4"

	firstRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      1,
		SomeValue:       "some value 1",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, firstRecord)
	assert.NoError(t, err)

	secondRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      2,
		SomeValue:       "some value 2",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, secondRecord)
	assert.NoError(t, err)

	thirdRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      3,
		SomeValue:       "some value 3",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, thirdRecord)
	assert.NoError(t, err)

	fourthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      4,
		SomeValue:       "some value 4",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fourthRecord)
	assert.NoError(t, err)

	fifthRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      5,
		SomeValue:       "some value 5",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, fifthRecord)
	assert.NoError(t, err)

	limit := 2
	keyRecord := compositeRecord{
		ThePartitionKey: partitionKeyValue,
		TheSortKey:      2,
	}
	actualRecords, err := compositeRecordsTable.
		Action(dynamodbClient).
		Query(ctx, keyRecord.PartitionKey(), keyRecord.SortKey(), true, limit)
	assert.NoError(t, err)
	assert.Equal(t, actualRecords[0], thirdRecord)
	assert.Equal(t, actualRecords[1], fourthRecord)

	assert.Len(t, actualRecords, 2)
}
