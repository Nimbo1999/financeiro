FROM golang:1.24-trixie as builder

COPY modules/migrator/ /modules/migrator/

WORKDIR /app

COPY users/ ./

RUN make build
RUN make build-health-checker

RUN chmod +x ./app-server

FROM scratch

WORKDIR /app

COPY --from=builder /app/health-checker .
COPY --from=builder /app/app-server .
COPY users/migrations migrations/

CMD [ "/app/app-server" ]
