BINARY_FILE_NAME=app-server
DB_NAME=financeiro_authentication
DB_PORT=25432
PORT=8082
POSTGRES_CONNECTION_STRING=matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)

PHONY: migrate-up run build

migrate-up:
	migrate -source file://migrations -database pgx://$(POSTGRES_CONNECTION_STRING) up

run:
	PORT=$(PORT) \
	POSTGRES_CONNECTION_STRING=postgres://$(POSTGRES_CONNECTION_STRING) \
	go run cmd/server/main.go

build:
	PORT=$(PORT) \
	POSTGRES_CONNECTION_STRING=postgres://$(POSTGRES_CONNECTION_STRING) \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o $(BINARY_FILE_NAME) cmd/server/main.go
