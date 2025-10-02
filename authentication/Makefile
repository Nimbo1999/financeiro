BINARY_FILE_NAME=app-server
DB_NAME=financeiro_authentication
DB_PORT=25432
HTTP_PORT=8082
GRPC_PORT=9092
POSTGRES_CONNECTION_STRING=matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)
RABBITMQ_URL=amqp://admin:admin123@localhost:5672/ # Using weak password configured in docker for local development
USER_GRPC_ADDRESS=localhost:9091

PHONY: migrate-up run build unit-test g-cert-files

migrate-up:
	migrate -source file://migrations -database pgx://$(POSTGRES_CONNECTION_STRING) up

run: g-cert-files migrate-up
	HTTP_PORT=$(HTTP_PORT) \
	GRPC_PORT=$(GRPC_PORT) \
	RABBITMQ_URL=$(RABBITMQ_URL) \
	USER_GRPC_ADDRESS=$(USER_GRPC_ADDRESS) \
	POSTGRES_CONNECTION_STRING=postgres://$(POSTGRES_CONNECTION_STRING) \
	go run cmd/server/main.go

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o $(BINARY_FILE_NAME) cmd/server/main.go

unit-test:
	go test -v -cover ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

# Generate JWT RSA256 cert files
g-cert-files:
	go run cmd/generate_certificates/main.go
