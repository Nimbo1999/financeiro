PORT=8081

PHONY: migrate run

migrate-up:
	migrate -source file://migrations -database pgx://matheus:dev1234@localhost:5432/financeiro_user up

run:
	PORT=$(PORT) go run cmd/container/main.go
