BINARY_FILE_NAME=app-server
DB_NAME=financeiro_notification
DB_PORT=35432
POSTGRES_CONNECTION_STRING=matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)
SMTP_EMAIL_FROM=noreply@localhost.com

.PHONY: run build test migrate-up migrate-down

run:
	HTTP_PORT=8084 \
	POSTGRES_CONNECTION_STRING=postgres://$(POSTGRES_CONNECTION_STRING) \
	RABBITMQ_URL=amqp://guest:guest@localhost:5672/ \
	WELCOME_QUEUE_NAME=notification.welcome \
	OTP_QUEUE_NAME=notification.otp \
	DLQ_NAME=notification.dlq \
	CONSUMER_PREFETCH_COUNT=10 \
	SMTP_HOST=smtp.gmail.com \
	SMTP_PORT=587 \
	SMTP_FROM=$(SMTP_EMAIL_FROM) \
	TEMPLATE_DIR=./internal/templates \
	go run cmd/server/main.go

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o $(BINARY_FILE_NAME) cmd/server/main.go

migrate-up:
	migrate -path migrations -database "pgx://$(POSTGRES_CONNECTION_STRING)" up

migrate-down:
	migrate -path migrations -database "pgx://$(POSTGRES_CONNECTION_STRING)" down

unit-test:
	go test ./... -v -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
