SERVICE_NAME=${SERVICE_NAME:=authentication}
OUTPUT_PATH=${OUTPUT_PATH:=./bin}
DB_NAME=financeiro_authentication
DB_PORT=25432
HTTP_PORT=8082
GRPC_PORT=9092
POSTGRES_CONNECTION_STRING=postgres://matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)
RABBITMQ_URL=amqp://admin:admin123@localhost:5672/ # Using weak password configured in docker for local development
USER_GRPC_ADDRESS=localhost:9091

PHONY: run build build-health-checker create-bin-dir unit-test g-cert-files

run: g-cert-files
	HTTP_PORT=$(HTTP_PORT) \
	GRPC_PORT=$(GRPC_PORT) \
	RABBITMQ_URL=$(RABBITMQ_URL) \
	USER_GRPC_ADDRESS=$(USER_GRPC_ADDRESS) \
	POSTGRES_CONNECTION_STRING=$(POSTGRES_CONNECTION_STRING) \
	go run cmd/server/main.go

build: g-cert-files build-health-checker
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME) cmd/server/main.go

build-health-checker: create-bin-dir
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME)-health-checker cmd/health_checker/main.go

create-bin-dir:
	mkdir -p $(OUTPUT_PATH)

unit-test:
	go test -v -cover ./internal/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

# Generate JWT RSA256 cert files
g-cert-files:
	go run cmd/generate_certificates/main.go
