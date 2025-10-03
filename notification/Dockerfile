FROM golang:1.24-trixie as builder

COPY modules/migrator/ /modules/migrator/

WORKDIR /app

COPY notification/ ./

RUN make build

RUN chmod +x ./app-server

FROM scratch

WORKDIR /app

COPY --from=builder /app/app-server .
COPY notification/migrations migrations/
COPY notification/internal/templates internal/templates/

CMD [ "/app/app-server" ]
