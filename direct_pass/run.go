package direct_pass

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"

	"github.com/gitlotto/common/env_var"
	"github.com/gitlotto/common/logging"
	"github.com/gitlotto/common/notification"
)

const nextStartIn = time.Minute * 10

func Run() {

	ctx := context.Background()

	logger := logging.MustCreateZuluTimeLogger()
	defer logger.Sync()

	envVarReader := env_var.EnvVarReader{
		Logger: logger,
	}

	workflowsTableName := envVarReader.MustFind("WORKFLOWS_TABLE_NAME")
	notificationTopicArn := envVarReader.MustFind("NOTIFICATION_TOPIC_ARN")
	awsRegion := envVarReader.MustFind("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		logger.Error("impossible to load AWS config", zap.Error(err))
		return
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)
	snsClient := sns.NewFromConfig(cfg)
	postman := notification.Postman{
		SnsClient: snsClient,
		TopicArn:  notificationTopicArn,
	}

	passer := DirectPasser{
		workflowsTableName:   workflowsTableName,
		notificationTopicArn: notificationTopicArn,
		nextStartIn:          nextStartIn,
		logger:               logger,
		dynamodbClient:       dynamodbClient,
		sqsClient:            sqsClient,
		postman:              postman,
	}

	handler := func(ctx context.Context, event events.DynamoDBEvent) {
		_ = passer.Pass(ctx, event)
	}

	lambda.Start(handler)

}
