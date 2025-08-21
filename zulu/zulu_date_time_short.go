package zulu

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DateTime struct {
	Date
	hour   int
	minute int
	second int
}

func (z DateTime) ToTime() time.Time {
	return time.Date(z.year, z.month, z.day, z.hour, z.minute, z.second, 0, time.UTC)
}

func DateTimeFromTime(t time.Time) DateTime {
	t = t.UTC()
	return DateTime{
		Date:   DateFromTime(t),
		hour:   t.Hour(),
		minute: t.Minute(),
		second: t.Second(),
	}
}

func DateTimeFromString(s string) (DateTime, error) {
	t, err := time.Parse("2006-01-02T15:04:05Z", s)
	if err != nil {
		return DateTime{}, err
	}
	return DateTimeFromTime(t), nil
}

func (z DateTime) String() string {
	return z.ToTime().Format("2006-01-02T15:04:05Z")
}

func (z DateTime) ToDate() Date {
	return z.Date
}

// MarshalDynamoDBAttributeValue implements attributevalue.Marshaler for AWS SDK v2
func (e DateTime) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: e.String()}, nil
}

// UnmarshalDynamoDBAttributeValue implements attributevalue.Unmarshaler for AWS SDK v2
func (e *DateTime) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		if v.Value == "" {
			return nil
		}
		t, err := time.Parse("2006-01-02T15:04:05Z", v.Value)
		if err != nil {
			return err
		}
		t = t.UTC()
		e.Date = DateFromTime(t)
		e.hour = t.Hour()
		e.minute = t.Minute()
		e.second = t.Second()
		return nil
	}
	return nil
}
