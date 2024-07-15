run:
	SLACK_BOT_TOKEN=... SLACK_USER_TOKEN=... go run main.go version

test:
	SLACK_BOT_TOKEN=test-bot-token SLACK_USER_TOKEN=test-user-token go test ./...

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/celebrations-macos-arm64 main.go

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/celebrations-linux-amd64 main.go

all: test build-macos-arm64 build-linux-amd64
