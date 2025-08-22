module github.com/gitlotto/common/direct_pass

go 1.24.6

require github.com/google/uuid v1.6.0

require github.com/davecgh/go-spew v1.1.1 // indirect

require github.com/stretchr/testify v1.9.0

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.29.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.28.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.33.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.37.1 // indirect
	github.com/aws/smithy-go v1.22.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.38.0
	github.com/aws/aws-sdk-go-v2/config v1.31.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.49.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.37.1
	github.com/aws/aws-sdk-go-v2/service/sqs v1.41.1
	github.com/gitlotto/common/database v0.0.0-00010101000000-000000000000
	github.com/gitlotto/common/env_var v0.0.0-00010101000000-000000000000
	github.com/gitlotto/common/logging v0.0.0-00010101000000-000000000000
	github.com/gitlotto/common/notification v0.0.0-00010101000000-000000000000
	github.com/gitlotto/common/queue v0.0.0-00010101000000-000000000000
	github.com/gitlotto/common/workflows v0.0.0-00010101000000-000000000000
	github.com/gitlotto/common/zulu v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
)

replace github.com/gitlotto/common/workflows => ../workflows

replace github.com/gitlotto/common/database => ../database

replace github.com/gitlotto/common/env_var => ../env_var

replace github.com/gitlotto/common/logging => ../logging

replace github.com/gitlotto/common/notification => ../notification

replace github.com/gitlotto/common/queue => ../queue

replace github.com/gitlotto/common/zulu => ../zulu
