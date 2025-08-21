package zulu

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Date struct {
	year  int
	month time.Month
	day   int
}

func (z Date) ToTime() time.Time {
	return time.Date(z.year, z.month, z.day, 0, 0, 0, 0, time.UTC)
}

func DateFromTime(t time.Time) Date {
	t = t.UTC()
	return Date{year: t.Year(), month: t.Month(), day: t.Day()}
}

func DateFromString(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return Date{}, err
	}
	return DateFromTime(t), nil
}

func (z Date) String() string {
	return z.ToTime().Format("2006-01-02")
}

// MarshalDynamoDBAttributeValue implements attributevalue.Marshaler for AWS SDK v2
func (e Date) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: e.String()}, nil
}

// UnmarshalDynamoDBAttributeValue implements attributevalue.Unmarshaler for AWS SDK v2
func (e *Date) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		if v.Value == "" {
			return nil
		}
		t, err := time.Parse("2006-01-02", v.Value)
		if err != nil {
			return err
		}
		t = t.UTC()
		e.year = t.Year()
		e.month = t.Month()
		e.day = t.Day()
		return nil
	}
	return nil
}
