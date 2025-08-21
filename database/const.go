package database

import (
	"fmt"
)

func ErrSortKeyIsMissing(tableName string, sortKey string) error {
	return fmt.Errorf("sort key %s is missing from record for table %s", sortKey, tableName)
}
