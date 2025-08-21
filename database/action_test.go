package database

import (
	"context"
	"strconv"
	"testing"

	"math/rand/v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_SimpleRecord_should_be_stored_in_correct_form(t *testing.T) {

	var err error
	ctx := context.TODO()

	partitionKey := uuid.New().String()

	record := simpleRecord{
		ThePartitionKey: partitionKey,
		SomeValue:       "some value",
	}

	err = simpleRecordsTable.Action(dynamodbClient).Persist(ctx, record)
	assert.NoError(t, err)

	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(simpleRecordsTableName),
		Key: map[string]types.AttributeValue{
			"partition_key": &types.AttributeValueMemberS{
				Value: partitionKey,
			},
		},
	}
	getEventOutput, err := dynamodbClient.GetItem(ctx, getItemInput)

	assert.NoError(t, err)
	actualItems := getEventOutput.Item

	expectedItems := map[string]types.AttributeValue{
		"partition_key": &types.AttributeValueMemberS{
			Value: partitionKey,
		},
		"some_value": &types.AttributeValueMemberS{
			Value: "some value",
		},
	}

	assert.Equal(t, expectedItems, actualItems)

	expectedRecord := record

	actualRecord := simpleRecord{}
	err = attributevalue.UnmarshalMap(getEventOutput.Item, &actualRecord)

	assert.NoError(t, err)
	assert.Equal(t, expectedRecord, actualRecord)

}

func Test_SimpleRecord_should_be_reconstituted_in_correct_form(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKey := uuid.New().String()

	record := simpleRecord{
		ThePartitionKey: partitionKey,
		SomeValue:       "some value",
	}

	expectedItems := map[string]types.AttributeValue{
		"partition_key": &types.AttributeValueMemberS{
			Value: partitionKey,
		},
		"some_value": &types.AttributeValueMemberS{
			Value: "some value",
		},
	}

	actualItems, err := attributevalue.MarshalMap(record)
	assert.NoError(t, err)
	assert.Equal(t, expectedItems, actualItems)

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(simpleRecordsTableName),
		Item:      actualItems,
	}
	_, err = dynamodbClient.PutItem(ctx, putItemInput)
	assert.NoError(t, err)

	recordWithOnlyKey := simpleRecord{
		ThePartitionKey: partitionKey,
	}

	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &recordWithOnlyKey)
	assert.NoError(t, err)

	expectedRecord := record
	actualRecord := recordWithOnlyKey
	assert.Equal(t, expectedRecord, actualRecord)

}

func Test_Reconstitut_should_return_error_if_the_simple_record_does_not_exist(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKey := uuid.New().String()

	record := simpleRecord{
		ThePartitionKey: partitionKey,
	}

	err = simpleRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &record)
	assert.ErrorIs(t, err, ErrNotFound)

}

func Test_CompositeRecord_should_be_stored_in_correct_form(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKey := uuid.New().String()
	sortKey := rand.Int()

	record := compositeRecord{
		ThePartitionKey: partitionKey,
		TheSortKey:      sortKey,
		SomeValue:       "some value",
	}

	err = compositeRecordsTable.Action(dynamodbClient).Persist(ctx, record)
	assert.NoError(t, err)

	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(compositeRecordsTableName),
		Key: map[string]types.AttributeValue{
			"partition_key": &types.AttributeValueMemberS{
				Value: partitionKey,
			},
			"sort_key": &types.AttributeValueMemberN{
				Value: strconv.Itoa(sortKey),
			},
		},
	}
	getEventOutput, err := dynamodbClient.GetItem(ctx, getItemInput)

	assert.NoError(t, err)
	actualItems := getEventOutput.Item

	expectedItems := map[string]types.AttributeValue{
		"partition_key": &types.AttributeValueMemberS{
			Value: partitionKey,
		},
		"sort_key": &types.AttributeValueMemberN{
			Value: strconv.Itoa(sortKey),
		},
		"some_value": &types.AttributeValueMemberS{
			Value: "some value",
		},
	}

	assert.Equal(t, expectedItems, actualItems)

	expectedRecord := record

	actualRecord := compositeRecord{}
	err = attributevalue.UnmarshalMap(getEventOutput.Item, &actualRecord)

	assert.NoError(t, err)
	assert.Equal(t, expectedRecord, actualRecord)

}

func Test_CompositeRecord_should_be_reconstituted_in_correct_form(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKey := uuid.New().String()
	sortKey := rand.Int()

	record := compositeRecord{
		ThePartitionKey: partitionKey,
		TheSortKey:      sortKey,
		SomeValue:       "some value",
	}

	expectedItems := map[string]types.AttributeValue{
		"partition_key": &types.AttributeValueMemberS{
			Value: partitionKey,
		},
		"sort_key": &types.AttributeValueMemberN{
			Value: strconv.Itoa(sortKey),
		},
		"some_value": &types.AttributeValueMemberS{
			Value: "some value",
		},
	}

	actualItems, err := attributevalue.MarshalMap(record)
	assert.NoError(t, err)
	assert.Equal(t, expectedItems, actualItems)

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(compositeRecordsTableName),
		Item:      actualItems,
	}
	_, err = dynamodbClient.PutItem(ctx, putItemInput)
	assert.NoError(t, err)

	recordWithOnlyKey := compositeRecord{
		ThePartitionKey: partitionKey,
		TheSortKey:      sortKey,
	}

	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &recordWithOnlyKey)
	assert.NoError(t, err)

	expectedRecord := record
	actualRecord := recordWithOnlyKey
	assert.Equal(t, expectedRecord, actualRecord)

}

func Test_Reconstitute_should_return_error_if_the_composite_record_does_not_exist(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKey := uuid.New().String()
	sortKey := rand.Int()

	record := compositeRecord{
		ThePartitionKey: partitionKey,
		TheSortKey:      sortKey,
	}

	err = compositeRecordsTable.Action(dynamodbClient).Reconstitute(ctx, &record)
	assert.ErrorIs(t, err, ErrNotFound)

}

func Test_Querying_should_fetch_composite_records_from_the_beginning_if_no_cursor_is_provided(t *testing.T) {
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
	actualRecords, nextCursor, err := compositeRecordsTable.Action(dynamodbClient).Query(ctx, firstRecord, nil, limit)
	assert.NoError(t, err)
	assert.Contains(t, actualRecords, fifthRecord)
	assert.Contains(t, actualRecords, fourthRecord)

	expectedCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiIyMjMxYmJkOS04MjQ3LTQxMTUtOTQxYS1jZjFkZDg3YTZmMWEifSwic29ydF9rZXkiOnsiTiI6IjQifX0="
	assert.NotNil(t, nextCursor)
	assert.Equal(t, expectedCursor, *nextCursor)

}

func Test_Querying_should_fetch_composite_records_from_the_given_cursor(t *testing.T) {
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
	startingCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiJiYzhmNmQ5Yi1jYzQ3LTQ2ZTctYTE4Yi00ODliNjNkOGRmYzQifSwic29ydF9rZXkiOnsiTiI6IjQifX0="

	actualRecords, nextCursor, err := compositeRecordsTable.Action(dynamodbClient).Query(ctx, firstRecord, &startingCursor, limit)
	assert.NoError(t, err)
	assert.Contains(t, actualRecords, thirdRecord)
	assert.Contains(t, actualRecords, secondRecord)

	expectedCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiJiYzhmNmQ5Yi1jYzQ3LTQ2ZTctYTE4Yi00ODliNjNkOGRmYzQifSwic29ydF9rZXkiOnsiTiI6IjIifX0="
	assert.NotNil(t, nextCursor)
	assert.Equal(t, expectedCursor, *nextCursor)

}

func Test_Querying_should_fetch_the_last_composite_records_and_return_nil_as_a_cursor(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKeyValue := "80f8f38f-8f63-4340-850d-fcbd6b95d826"

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

	limit := 10
	startingCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiI4MGY4ZjM4Zi04ZjYzLTQzNDAtODUwZC1mY2JkNmI5NWQ4MjYifSwic29ydF9rZXkiOnsiTiI6IjQifX0="

	actualRecords, nextCursor, err := compositeRecordsTable.Action(dynamodbClient).Query(ctx, firstRecord, &startingCursor, limit)
	assert.NoError(t, err)
	assert.Contains(t, actualRecords, thirdRecord)
	assert.Contains(t, actualRecords, secondRecord)
	assert.Contains(t, actualRecords, firstRecord)

	assert.Nil(t, nextCursor)

}

func Test_QueryingAsc_should_fetch_composite_records_from_the_beginning_if_no_cursor_is_provided(t *testing.T) {
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
	actualRecords, nextCursor, err := compositeRecordsTable.Action(dynamodbClient).QueryAsc(ctx, firstRecord, nil, limit)
	assert.NoError(t, err)
	assert.Contains(t, actualRecords, firstRecord)
	assert.Contains(t, actualRecords, secondRecord)

	expectedCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiIyMjMxYmJkOS04MjQ3LTQxMTUtOTQxYS1jZjFkZDg3YTZmMWEifSwic29ydF9rZXkiOnsiTiI6IjIifX0="
	assert.NotNil(t, nextCursor)
	assert.Equal(t, expectedCursor, *nextCursor)

}

func Test_QueryingAsc_should_fetch_composite_records_from_the_given_cursor(t *testing.T) {
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
	startingCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiJiYzhmNmQ5Yi1jYzQ3LTQ2ZTctYTE4Yi00ODliNjNkOGRmYzQifSwic29ydF9rZXkiOnsiTiI6IjIifX0="

	actualRecords, nextCursor, err := compositeRecordsTable.Action(dynamodbClient).QueryAsc(ctx, firstRecord, &startingCursor, limit)
	assert.NoError(t, err)
	assert.Contains(t, actualRecords, thirdRecord)
	assert.Contains(t, actualRecords, fourthRecord)

	expectedCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiJiYzhmNmQ5Yi1jYzQ3LTQ2ZTctYTE4Yi00ODliNjNkOGRmYzQifSwic29ydF9rZXkiOnsiTiI6IjQifX0="
	assert.NotNil(t, nextCursor)
	assert.Equal(t, expectedCursor, *nextCursor)

}

func Test_QueryingAsc_should_fetch_the_last_composite_records_and_return_nil_as_a_cursor(t *testing.T) {
	var err error
	ctx := context.TODO()

	partitionKeyValue := "80f8f38f-8f63-4340-850d-fcbd6b95d826"

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

	limit := 10
	startingCursor := "eyJwYXJ0aXRpb25fa2V5Ijp7IlMiOiI4MGY4ZjM4Zi04ZjYzLTQzNDAtODUwZC1mY2JkNmI5NWQ4MjYifSwic29ydF9rZXkiOnsiTiI6IjIifX0="

	actualRecords, nextCursor, err := compositeRecordsTable.Action(dynamodbClient).QueryAsc(ctx, firstRecord, &startingCursor, limit)
	assert.NoError(t, err)
	assert.Contains(t, actualRecords, fifthRecord)
	assert.Contains(t, actualRecords, fourthRecord)
	assert.Contains(t, actualRecords, thirdRecord)

	assert.Nil(t, nextCursor)

}
