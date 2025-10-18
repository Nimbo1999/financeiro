FROM golang:1.24-trixie as builder

COPY modules/migrator/ /modules/migrator/
COPY users/ /users
COPY commons /commons

WORKDIR /app

COPY transactions/ ./

RUN make OUTPUT_PATH="./bin" SERVICE_NAME="transactions" build

RUN chmod +x ./bin/transactions
RUN chmod +x ./bin/transactions-health-checker

FROM scratch

WORKDIR /app

COPY --from=builder /app/bin/transactions bin/transactions
COPY --from=builder /app/bin/transactions-health-checker bin/transactions-health-checker
COPY transactions/migrations ./migrations

CMD [ "/app/bin/transactions" ]
