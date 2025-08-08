# Makefile for GitLotto Common

.PHONY: all deps test setup-python setup-aws setup-samlocal docker-up docker-down

MODULES = api batcher database env_var logging notification outboxer queue workflows zulu

# Install Go dependencies for all modules
deps:
	@echo "Installing Go dependencies for all modules..."
	@go env -w GOPRIVATE=github.com/gitlotto
	@for dir in $(MODULES); do \
		cd $$dir && go get ./... && go mod tidy && cd ..; \
	done

# Start local development infrastructure
dev_up:
	@echo "🚀 Starting local development infrastructure..."
	@docker compose -f .dev/docker-compose_local.yaml --env-file .dev/.env -p gitlotto up -d
	@echo "✅ Docker containers started in detached mode"

# Start CI/CD testing infrastructure
ci_up:
	@echo "🚀 Starting CI/CD testing infrastructure..."
	@docker compose -f .dev/docker-compose.yaml --env-file .dev/.env -p gitlotto up -d
	@echo "✅ Docker containers started in detached mode"

# Setup local AWS infrastructure (requires dev-up to be running)
setup:
	@echo "🔧 Setting up local AWS infrastructure..."
	@aws --endpoint-url http://localhost:4566 s3api create-bucket --bucket gitlotto
	@echo "📦 S3 bucket created"
	@samlocal deploy --template-file database/.dev/db.yaml --stack-name database --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=database
	@echo "💾 database stack deployed"

	@samlocal deploy --template-file outboxer/.dev/db.yaml --stack-name outboxer_dynamodb --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=outboxer_dynamodb
	@samlocal deploy --template-file outboxer/.dev/random_queues.yaml --stack-name outboxer_random_queues --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=outboxer_random_queues
	@samlocal deploy --template-file outboxer/.dev/notification.yaml --stack-name outboxer_notification --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=outboxer_notification
	@echo "💾 outboxer stack deployed"

	@samlocal deploy --template-file workflows/.dev/db.yaml --stack-name workflows --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=workflows
	@echo "💾 workflows stack deployed"

	@echo "✅ Local AWS infrastructure setup complete!"
# Stop and remove local development infrastructure
dev_down:
	@echo "🛑 Stopping local development infrastructure..."
	@docker compose -f ./.dev/docker-compose_local.yaml --env-file ./.dev/.env -p divanbashi down
	@echo "✅ Docker containers stopped and removed"

ci_down:
	@echo "🛑 Stopping CI/CD testing infrastructure..."
	@docker compose -f .dev/docker-compose.yaml --env-file .dev/.env -p gitlotto down
	@echo "✅ Docker containers stopped and removed"

# Full reset: stop everything, start, and setup
dev_reset: dev_down dev_up dev_setup

ci_reset: ci_down ci_up ci_setup

# Test and deploy for all modules
test:
	cd api && go test ./... -v -count=1 -p 1 && cd ..
	cd batcher && go test ./... -v -count=1 -p 1 && cd ..
	cd database && go test ./... -v -count=1 -p 1 && cd ..
	cd env_var && go test ./... -v -count=1 -p 1 && cd ..
	cd workflows && go test ./... -v -count=1 -p 1 && cd ..
	cd outboxer && go test ./... -v -count=1 -p 1 && cd ..
	cd zulu && go test ./... -v -count=1 -p 1 && cd ..
