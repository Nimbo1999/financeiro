BINARY_FILE_NAME=app-server
USER_SERVICE_URL=http://localhost:8081
AUTH_SERVICE_URL=http://localhost:8082
AUTH_SERVICE_GRPC_URL=localhost:9092
NOTIFICATION_SERVICE_URL=http://localhost:8084
GATEWAY_PORT=8083

PHONY: run build unit-test

run:
	GATEWAY_PORT=$(GATEWAY_PORT) \
	AUTH_SERVICE_GRPC_URL=$(AUTH_SERVICE_GRPC_URL) \
	USER_SERVICE_URL=$(USER_SERVICE_URL) \
	AUTH_SERVICE_URL=$(AUTH_SERVICE_URL) \
	NOTIFICATION_SERVICE_URL=$(NOTIFICATION_SERVICE_URL) \
	go run cmd/server/main.go

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o $(BINARY_FILE_NAME) cmd/server/main.go

unit-test:
	go test -v -cover ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
