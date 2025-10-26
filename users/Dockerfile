FROM golang:1.24-trixie as builder

COPY modules/migrator/ /modules/migrator/
COPY commons/ /commons/

WORKDIR /app

COPY users/ ./

RUN make OUTPUT_PATH="./bin" SERVICE_NAME="users" build

RUN chmod +x ./bin/users
RUN chmod +x ./bin/users-health-checker

FROM scratch

WORKDIR /app

COPY --from=builder /app/bin/users-health-checker bin/users-health-checker
COPY --from=builder /app/bin/users bin/users
COPY users/migrations ./migrations

CMD [ "/app/bin/users" ]
