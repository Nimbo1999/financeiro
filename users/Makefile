PORT=8081
POSTGRES_CONNECTION_STRING=matheus:dev1234@localhost:5432/financeiro_user

PHONY: migrate run

migrate-up:
	migrate -source file://migrations -database pgx://$(POSTGRES_CONNECTION_STRING) up

run:
	PORT=$(PORT) \
	POSTGRES_CONNECTION_STRING=postgres://$(POSTGRES_CONNECTION_STRING) \
	go run cmd/container/main.go
