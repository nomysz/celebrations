help:  ## Display help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "\nUsage:\n  make \033[32m<target>\033[0m\n\nTargets:\n"}; {printf "\t\033[32m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL = run

run: ## Run development version (display available commands)
	SLACK_BOT_TOKEN=... SLACK_USER_TOKEN=... go run main.go

version: ## Display current version
	SLACK_BOT_TOKEN=... SLACK_USER_TOKEN=... go run main.go version

test: ## Run test suite
	SLACK_BOT_TOKEN=test-bot-token SLACK_USER_TOKEN=test-user-token go test ./...

build-macos-arm64: ## Build for MacOS
	GOOS=darwin GOARCH=arm64 go build -o bin/celebrations-macos-arm64 main.go

build-linux-amd64: ## Build for Linux
	GOOS=linux GOARCH=amd64 go build -o bin/celebrations-linux-amd64 main.go

all: test build-macos-arm64 build-linux-amd64 ## Run tests, then build for Linux and MacOS
