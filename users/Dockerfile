FROM golang:1.24-alpine as builder

WORKDIR /app

COPY . .

RUN make build

RUN chmod +x ./app-server

FROM scratch

WORKDIR /app

COPY --from=builder /app/app-server .

CMD [ "/app/app-server" ]
