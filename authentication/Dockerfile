FROM golang:1.24-trixie as builder

COPY modules/migrator/ /modules/migrator/
COPY users/ /users

WORKDIR /app

COPY authentication/ ./

RUN make build
RUN make build-health-checker

RUN chmod +x ./app-server

FROM scratch

WORKDIR /app

COPY --from=builder /app/health-checker .
COPY --from=builder /app/app-server .
COPY --from=builder /app/private_key.pem .
COPY --from=builder /app/public_key.pem .
COPY authentication/migrations migrations/

CMD [ "/app/app-server" ]
