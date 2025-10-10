SERVICE_NAME=${SERVICE_NAME:=users}
OUTPUT_PATH=${OUTPUT_PATH:=./bin}
DB_NAME=financeiro_user
HTTP_PORT=8081
GRPC_PORT=9091
DB_PORT=15432
POSTGRES_CONNECTION_STRING=postgres://matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)
RABBITMQ_URL=amqp://admin:admin123@localhost:5672

PHONY: run build build-health-checker create-bin-dir generate-proto unit-test

run:
	HTTP_PORT=$(HTTP_PORT) \
	GRPC_PORT=$(GRPC_PORT) \
	POSTGRES_CONNECTION_STRING=$(POSTGRES_CONNECTION_STRING) \
	RABBITMQ_URL=$(RABBITMQ_URL) \
	go run cmd/server/main.go

build: build-health-checker
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME) cmd/server/main.go

build-health-checker: create-bin-dir
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME)-health-checker cmd/health_checker/main.go

create-bin-dir:
	mkdir -p $(OUTPUT_PATH)

generate-proto:
	./scripts/generate_proto.sh

unit-test:
	go test -v -cover ./internal/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
