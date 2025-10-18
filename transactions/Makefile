SERVICE_NAME=${SERVICE_NAME:=transactions}
OUTPUT_PATH=${OUTPUT_PATH:=./bin}
DB_NAME=financeiro_transaction
DB_PORT=45432
POSTGRES_CONNECTION_STRING=postgres://matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)

PHONY: run unit-test build build-health-checker

run:
	HTTP_PORT=8084 \
	POSTGRES_CONNECTION_STRING=$(POSTGRES_CONNECTION_STRING) \
	go run cmd/server/main.go

unit-test:
	go test -v -cover ./internal/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

build: build-health-checker
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME) cmd/server/main.go

build-health-checker:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME)-health-checker cmd/health_checker/main.go
