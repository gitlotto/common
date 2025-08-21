package database

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Record interface {
	PartitionKey() types.AttributeValue
	SortKey() *types.AttributeValue
}

type Table[R Record] struct {
	Name         string
	PartitionKey string
	SortKey      *string
}

func (t Table[R]) PrimaryKey(record R) (keys map[string]types.AttributeValue, err error) {
	keys = map[string]types.AttributeValue{
		t.PartitionKey: record.PartitionKey(),
	}
	if t.SortKey != nil {
		maybeSortKey := record.SortKey()
		if maybeSortKey == nil {
			err = ErrSortKeyIsMissing(t.Name, *t.SortKey)
			return
		}
		keys[*t.SortKey] = *maybeSortKey
	}
	return
}

func NewTable[R Record](name string, partitionKey string, sortKey *string) Table[R] {
	return Table[R]{
		Name:         name,
		PartitionKey: partitionKey,
		SortKey:      sortKey,
	}
}
