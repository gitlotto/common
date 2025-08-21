package notification

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
)

const onlyOneMessageGroupId string = "0"

type Postman struct {
	SnsClient *sns.Client
	TopicArn  string
}

func (postman *Postman) SendNotification(ctx context.Context, requestId string, message string) (err error) {
	notification := JobNotification{
		RequestId: requestId,
		Message:   message,
	}

	messageBytes, err := json.Marshal(notification)
	if err != nil {
		return
	}

	messageBody := string(messageBytes)

	deduplicationId := uuid.New().String()

	publishInput := &sns.PublishInput{
		Message:                aws.String(messageBody),
		TopicArn:               &postman.TopicArn,
		MessageGroupId:         aws.String(onlyOneMessageGroupId),
		MessageDeduplicationId: aws.String(deduplicationId),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"NotificationType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Job"),
			},
		},
	}

	_, err = postman.SnsClient.Publish(ctx, publishInput)
	return
}
