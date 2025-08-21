package queue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// this is a test util. But it is accessible from production code. Figure out how to create shared test utils in go
func GetLastNCommands(ctx context.Context, svc *sqs.Client, queueUrl string, numberOfMessages int) (lastNMessages []types.Message, err error) {

	var lastMessages []types.Message
	supposedlyHasMessages := true

	for {
		if !supposedlyHasMessages {
			break
		}
		var resp *sqs.ReceiveMessageOutput
		receiveMessageInput := &sqs.ReceiveMessageInput{
			QueueUrl:            &queueUrl,
			MaxNumberOfMessages: int32(10), // You can adjust this number
			VisibilityTimeout:   int32(10),
			WaitTimeSeconds:     int32(0),
		}
		resp, err = svc.ReceiveMessage(ctx, receiveMessageInput)
		if err != nil {
			fmt.Printf("Failed to fetch message with error%v", err)
			return
		}

		if (len(resp.Messages)) == 0 {
			supposedlyHasMessages = false
			break
		}

		for _, message := range resp.Messages {
			deleteMessageInput := &sqs.DeleteMessageInput{
				QueueUrl:      &queueUrl,
				ReceiptHandle: message.ReceiptHandle,
			}
			_, err = svc.DeleteMessage(ctx, deleteMessageInput)
			if err != nil {
				fmt.Printf("Failed to delete message with error%v", err)
				return
			}
			lastMessages = append(lastMessages, message)
		}
	}

	if len(lastMessages) > numberOfMessages {
		lastNMessages = lastMessages[len(lastMessages)-numberOfMessages:]
	} else {
		lastNMessages = lastMessages
	}

	return
}
