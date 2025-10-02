BINARY_FILE_NAME=app-server
DB_NAME=financeiro_user
HTTP_PORT=8081
GRPC_PORT=9091
DB_PORT=15432
POSTGRES_CONNECTION_STRING=matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)
RABBITMQ_URL=amqp://admin:admin123@localhost:5672/

PHONY: migrate-up run build generate-proto

migrate-up:
	migrate -source file://migrations -database pgx://$(POSTGRES_CONNECTION_STRING) up

run: migrate-up
	HTTP_PORT=$(HTTP_PORT) \
	GRPC_PORT=$(GRPC_PORT) \
	POSTGRES_CONNECTION_STRING=postgres://$(POSTGRES_CONNECTION_STRING) \
	RABBITMQ_URL=$(RABBITMQ_URL) \
	go run cmd/container/main.go

build:
	HTTP_PORT=$(HTTP_PORT) \
	GRPC_PORT=$(GRPC_PORT) \
	POSTGRES_CONNECTION_STRING=postgres://$(POSTGRES_CONNECTION_STRING) \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o $(BINARY_FILE_NAME) cmd/container/main.go

generate-proto:
	./scripts/generate_proto.sh
