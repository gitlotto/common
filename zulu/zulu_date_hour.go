package zulu

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DateHour struct {
	Date
	hour int
}

func (z DateHour) ToTime() time.Time {
	return time.Date(z.year, z.month, z.day, z.hour, 0, 0, 0, time.UTC)
}

func DateHourFromTime(t time.Time) DateHour {
	t = t.UTC()
	return DateHour{
		Date: DateFromTime(t),
		hour: t.Hour(),
	}
}

func DateHourFromString(s string) (DateHour, error) {
	t, err := time.Parse("2006-01-02T15Z", s)
	if err != nil {
		return DateHour{}, err
	}
	return DateHourFromTime(t), nil
}

func (z DateHour) String() string {
	return z.ToTime().Format("2006-01-02T15Z")
}

func (z DateHour) ToDate() Date {
	return z.Date
}

// MarshalDynamoDBAttributeValue implements attributevalue.Marshaler for AWS SDK v2
func (e DateHour) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: e.String()}, nil
}

// UnmarshalDynamoDBAttributeValue implements attributevalue.Unmarshaler for AWS SDK v2
func (e *DateHour) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		if v.Value == "" {
			return nil
		}
		t, err := time.Parse("2006-01-02T15Z", v.Value)
		if err != nil {
			return err
		}
		t = t.UTC()
		e.Date = DateFromTime(t)
		e.hour = t.Hour()
		return nil
	}
	return nil
}
