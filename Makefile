# Makefile for GitLotto Common

.PHONY: deps dev_up ci_up setup dev_down ci_down dev_reset ci_reset test tag-create tag-push tag-all

MODULES = api batcher database direct_pass env_var logging notification outboxer queue workflows zulu

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
dev_setup:
	@echo "🔧 Setting up local AWS infrastructure..."
	@aws --endpoint-url http://localhost:4566 s3api create-bucket --bucket gitlotto
	@echo "📦 S3 bucket created"
	@samlocal deploy --template-file database/.dev/db.yaml --stack-name database --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=database
	@echo "💾 database stack deployed"

	@samlocal deploy --template-file outboxer/.dev/db.yaml --stack-name outboxer_dynamodb --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=outboxer_dynamodb
	@samlocal deploy --template-file outboxer/.dev/random_queues.yaml --stack-name outboxer_random_queues --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=outboxer_random_queues
	@samlocal deploy --template-file outboxer/.dev/notification.yaml --stack-name outboxer_notification --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=outboxer_notification
	@echo "💾 outboxer stack deployed"

	@samlocal deploy --template-file direct_pass/.dev/db.yaml --stack-name direct_passer_dynamodb --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=direct_passer_dynamodb
	@samlocal deploy --template-file direct_pass/.dev/queues.yaml --stack-name direct_passer_queues --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=direct_passer_queues
	@samlocal deploy --template-file direct_pass/.dev/notification.yaml --stack-name direct_passer_notification --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --s3-bucket gitlotto --parameter-overrides TheStackName=direct_passer_notification
	@echo "💾 direct_passer stack deployed"

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
	cd direct_pass && go test ./... -v -count=1 -p 1 && cd ..
	cd zulu && go test ./... -v -count=1 -p 1 && cd ..

# Create tags for all modules with the same version (local only)
# Usage: make tag-create VERSION=v0.17.0
tag-create:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required. Usage: make tag-create VERSION=v0.17.0"; \
		exit 1; \
	fi
	@echo "🏷️  Creating tags for all modules with version $(VERSION)..."
	@for module in $(MODULES); do \
		echo "  📦 Creating tag $$module/$(VERSION)"; \
		git tag $$module/$(VERSION); \
	done
	@echo "✅ All module tags created locally with $(VERSION)!"

# Push all module tags for a specific version to remote
# Usage: make tag-push VERSION=v0.17.0
tag-push:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required. Usage: make tag-push VERSION=v0.17.0"; \
		exit 1; \
	fi
	@echo "🚀 Pushing module tags for version $(VERSION) to remote..."
	@for module in $(MODULES); do \
		echo "  🚀 Pushing $$module/$(VERSION)"; \
		git push origin $$module/$(VERSION); \
	done
	@echo "✅ All module tags for $(VERSION) pushed to remote!"

# Create and push tags for all modules (combines tag-create and tag-push)
# Usage: make tag-all VERSION=v0.17.0
tag-all: tag-create tag-push
