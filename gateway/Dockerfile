FROM golang:1.24-trixie as builder

COPY authentication/ /authentication

WORKDIR /app

COPY gateway/ ./

RUN make build

RUN chmod +x ./app-server

FROM scratch

WORKDIR /app

COPY --from=builder /app/app-server .

CMD [ "/app/app-server" ]
