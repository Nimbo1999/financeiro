SERVICE_NAME=${SERVICE_NAME:=notification}
OUTPUT_PATH=${OUTPUT_PATH:=./bin}
DB_NAME=financeiro_notification
DB_PORT=35432
POSTGRES_CONNECTION_STRING=postgres://matheus:dev1234@localhost:$(DB_PORT)/$(DB_NAME)
SMTP_USERNAME=${EMAIL}
SMTP_PASSWORD=${EMAIL_PASSWORD}
RABBITMQ_URL=amqp://admin:admin123@localhost:5672/ # Using weak password configured in docker for local development

.PHONY: run build build-health-checker create-bin-dir test unit-test

run:
	HTTP_PORT=8084 \
	POSTGRES_CONNECTION_STRING=$(POSTGRES_CONNECTION_STRING) \
	RABBITMQ_URL=$(RABBITMQ_URL) \
	SMTP_USERNAME=$(SMTP_USERNAME) \
	SMTP_PASSWORD=$(SMTP_PASSWORD) \
	SMTP_PORT=465 \
	SMTP_FROM=$(SMTP_USERNAME) \
	go run cmd/server/main.go

build: build-health-checker
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME) cmd/server/main.go

build-health-checker:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-w -s" -o $(OUTPUT_PATH)/$(SERVICE_NAME)-health-checker cmd/health_checker/main.go

create-bin-dir:
	mkdir -p $(OUTPUT_PATH)

unit-test:
	go test ./... -v -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
