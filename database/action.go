package database

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrNotFound = fmt.Errorf("record not found")

func (table Table[R]) Action(dynamodbClient *dynamodb.Client) TableAction[R] {
	return TableAction[R]{
		Table:          table,
		DynamodbClient: dynamodbClient,
	}
}

type TableAction[R Record] struct {
	Table[R]
	DynamodbClient *dynamodb.Client
}

func (table TableAction[R]) Reconstitute(ctx context.Context, recordWithKey *R) (err error) {

	if recordWithKey == nil {
		return
	}
	keys, err := table.PrimaryKey(*recordWithKey)
	if err != nil {
		return
	}
	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(table.Table.Name),
		Key:       keys,
	}
	result, err := table.DynamodbClient.GetItem(ctx, getItemInput)
	if err != nil {
		return
	}
	if len(result.Item) == 0 {
		return ErrNotFound
	}

	err = attributevalue.UnmarshalMap(result.Item, recordWithKey)
	if err != nil {
		return
	}
	return
}

func (table TableAction[R]) Persist(ctx context.Context, record R) (err error) {

	items, err := attributevalue.MarshalMap(record)
	if err != nil {
		return
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(table.Table.Name),
		Item:      items,
	}

	_, err = table.DynamodbClient.PutItem(ctx, putItemInput)
	return
}

func (table TableAction[R]) Query(ctx context.Context, record R, cursor *string, limit int) (records []R, nextCursor *string, err error) {

	partitionKeyName := table.PartitionKey
	parititonKeyValue := record.PartitionKey()

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(table.Name),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :the_partition_key", partitionKeyName)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":the_partition_key": parititonKeyValue,
		},
		ScanIndexForward: aws.Bool(false),
	}

	if cursor != nil {
		exclusiveStartKey, errOfDecoding := decodeCursor(*cursor)
		if errOfDecoding != nil {
			err = errOfDecoding
			return
		}
		queryInput.ExclusiveStartKey = exclusiveStartKey
	}

	queryInput.Limit = aws.Int32(int32(limit))

	items, err := table.DynamodbClient.Query(ctx, queryInput)
	if err != nil {
		return
	}

	records = make([]R, len(items.Items))
	err = attributevalue.UnmarshalListOfMaps(items.Items, &records)
	if err != nil {
		return
	}

	nextCursor, err = encodeCursor(items.LastEvaluatedKey)
	if err != nil {
		return
	}
	return
}

func (table TableAction[R]) QueryAsc(ctx context.Context, record R, cursor *string, limit int) (records []R, nextCursor *string, err error) {

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(table.Name),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :the_partition_key", table.PartitionKey)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":the_partition_key": record.PartitionKey(),
		},
		ScanIndexForward: aws.Bool(true),
	}

	if cursor != nil {
		exclusiveStartKey, errOfDecoding := decodeCursor(*cursor)
		if errOfDecoding != nil {
			err = errOfDecoding
			return
		}
		queryInput.ExclusiveStartKey = exclusiveStartKey
	}

	queryInput.Limit = aws.Int32(int32(limit))

	items, err := table.DynamodbClient.Query(ctx, queryInput)
	if err != nil {
		return
	}

	records = make([]R, len(items.Items))
	err = attributevalue.UnmarshalListOfMaps(items.Items, &records)
	if err != nil {
		return
	}

	nextCursor, err = encodeCursor(items.LastEvaluatedKey)
	if err != nil {
		return
	}
	return
}

func decodeCursor(cursor string) (exclusiveStartKey map[string]types.AttributeValue, err error) {
	var decodedCursor []byte
	decodedCursor, err = base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return
	}
	cursorAttributes := map[string]AttributeValueWrapper{}
	err = json.Unmarshal(decodedCursor, &cursorAttributes)
	if err != nil {
		return
	}

	exclusiveStartKey = map[string]types.AttributeValue{}
	for key, value := range cursorAttributes {
		exclusiveStartKey[key] = value.AttributeValue
	}
	return
}

func encodeCursor(exclusiveStartKey map[string]types.AttributeValue) (cursor *string, err error) {
	if len(exclusiveStartKey) == 0 {
		return
	}
	cursorAttributes := map[string]*AttributeValueWrapper{}
	for key, value := range exclusiveStartKey {
		cursorAttributes[key] = &AttributeValueWrapper{value}
	}
	var cursorBytes []byte
	cursorBytes, err = json.Marshal(cursorAttributes)
	if err != nil {
		return
	}
	cursorCandidate := base64.StdEncoding.EncodeToString(cursorBytes)
	cursor = &cursorCandidate
	return
}

type AttributeValueWrapper struct {
	types.AttributeValue
}

func (avw *AttributeValueWrapper) MarshalJSON() ([]byte, error) {
	jsonAV := toJson(avw.AttributeValue)
	return json.Marshal(jsonAV)
}

func (avw *AttributeValueWrapper) UnmarshalJSON(data []byte) error {
	var jsonAV AttributeValueJSON
	if err := json.Unmarshal(data, &jsonAV); err != nil {
		return err
	}
	avw.AttributeValue = fromJson(&jsonAV)
	return nil
}

type AttributeValueJSON struct {
	B    []byte                         `json:"B,omitempty"`
	BOOL *bool                          `json:"BOOL,omitempty"`
	BS   [][]byte                       `json:"BS,omitempty"`
	L    []*AttributeValueJSON          `json:"L,omitempty"`
	M    map[string]*AttributeValueJSON `json:"M,omitempty"`
	N    *string                        `json:"N,omitempty"`
	NS   []string                       `json:"NS,omitempty"`
	NULL *bool                          `json:"NULL,omitempty"`
	S    *string                        `json:"S,omitempty"`
	SS   []string                       `json:"SS,omitempty"`
}

func toJson(av types.AttributeValue) *AttributeValueJSON {
	if av == nil {
		return nil
	}

	jsonAV := &AttributeValueJSON{}

	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		jsonAV.S = &v.Value
	case *types.AttributeValueMemberN:
		jsonAV.N = &v.Value
	case *types.AttributeValueMemberB:
		jsonAV.B = v.Value
	case *types.AttributeValueMemberBOOL:
		jsonAV.BOOL = &v.Value
	case *types.AttributeValueMemberBS:
		jsonAV.BS = v.Value
	case *types.AttributeValueMemberL:
		jsonAV.L = listToJson(v.Value)
	case *types.AttributeValueMemberM:
		jsonAV.M = mapToJson(v.Value)
	case *types.AttributeValueMemberNS:
		jsonAV.NS = v.Value
	case *types.AttributeValueMemberNULL:
		jsonAV.NULL = &v.Value
	case *types.AttributeValueMemberSS:
		jsonAV.SS = v.Value
	}

	return jsonAV
}

func listToJson(list []types.AttributeValue) []*AttributeValueJSON {
	if list == nil {
		return nil
	}
	result := make([]*AttributeValueJSON, len(list))
	for i, item := range list {
		result[i] = toJson(item)
	}
	return result
}

func mapToJson(m map[string]types.AttributeValue) map[string]*AttributeValueJSON {
	if m == nil {
		return nil
	}
	result := make(map[string]*AttributeValueJSON)
	for key, value := range m {
		result[key] = toJson(value)
	}
	return result
}

func fromJson(jsonAV *AttributeValueJSON) types.AttributeValue {
	if jsonAV == nil {
		return nil
	}

	// Check each field and return the appropriate concrete type
	if jsonAV.S != nil {
		return &types.AttributeValueMemberS{Value: *jsonAV.S}
	}
	if jsonAV.N != nil {
		return &types.AttributeValueMemberN{Value: *jsonAV.N}
	}
	if jsonAV.B != nil {
		return &types.AttributeValueMemberB{Value: jsonAV.B}
	}
	if jsonAV.BOOL != nil {
		return &types.AttributeValueMemberBOOL{Value: *jsonAV.BOOL}
	}
	if jsonAV.BS != nil {
		return &types.AttributeValueMemberBS{Value: jsonAV.BS}
	}
	if jsonAV.L != nil {
		return &types.AttributeValueMemberL{Value: listFromJSON(jsonAV.L)}
	}
	if jsonAV.M != nil {
		return &types.AttributeValueMemberM{Value: mapFromJSON(jsonAV.M)}
	}
	if jsonAV.NS != nil {
		return &types.AttributeValueMemberNS{Value: jsonAV.NS}
	}
	if jsonAV.NULL != nil {
		return &types.AttributeValueMemberNULL{Value: *jsonAV.NULL}
	}
	if jsonAV.SS != nil {
		return &types.AttributeValueMemberSS{Value: jsonAV.SS}
	}

	return nil
}

func listFromJSON(list []*AttributeValueJSON) []types.AttributeValue {
	if list == nil {
		return nil
	}
	result := make([]types.AttributeValue, len(list))
	for i, item := range list {
		result[i] = fromJson(item)
	}
	return result
}

func mapFromJSON(m map[string]*AttributeValueJSON) map[string]types.AttributeValue {
	if m == nil {
		return nil
	}
	result := make(map[string]types.AttributeValue)
	for key, value := range m {
		result[key] = fromJson(value)
	}
	return result
}
