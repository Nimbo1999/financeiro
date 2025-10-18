DB_NAME=financeiro_transaction
DB_PORT=45432
POSTGRES_CONNECTION_STRING=postgres://matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)

PHONY: run unit-test

run:
	HTTP_PORT=8084 \
	POSTGRES_CONNECTION_STRING=$(POSTGRES_CONNECTION_STRING) \
	go run cmd/server/main.go

unit-test:
	go test -v -cover ./internal/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
