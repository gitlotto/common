package zulu

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DateTimeMillis struct {
	DateTime
	milli int
}

func (z DateTimeMillis) ToTime() time.Time {
	return time.Date(z.year, z.month, z.day, z.hour, z.minute, z.second, z.milli*1000000, time.UTC)
}

func DateTimeMillisFromTime(t time.Time) DateTimeMillis {
	t = t.UTC()
	return DateTimeMillis{
		DateTime: DateTimeFromTime(t),
		milli:    t.Nanosecond() / 1000000,
	}
}

func DateTimeMillisFromString(s string) (DateTimeMillis, error) {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", s)
	t = t.UTC()
	if err != nil {
		return DateTimeMillis{}, err
	}
	return DateTimeMillisFromTime(t), nil
}

func (z DateTimeMillis) String() string {
	return z.ToTime().Format("2006-01-02T15:04:05.000Z")
}

func (z DateTimeMillis) ToDate() Date {
	return z.Date
}

// MarshalDynamoDBAttributeValue implements attributevalue.Marshaler for AWS SDK v2
func (e DateTimeMillis) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: e.String()}, nil
}

// UnmarshalDynamoDBAttributeValue implements attributevalue.Unmarshaler for AWS SDK v2
func (e *DateTimeMillis) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		if v.Value == "" {
			return nil
		}
		t, err := time.Parse("2006-01-02T15:04:05.000Z", v.Value)
		if err != nil {
			return err
		}
		t = t.UTC()
		e.DateTime = DateTimeFromTime(t)
		e.milli = t.Nanosecond() / 1000000
		return nil
	}
	return nil
}
