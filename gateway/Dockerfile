FROM golang:1.24-trixie as builder

COPY authentication/ /authentication

WORKDIR /app

COPY gateway/ ./

RUN make OUTPUT_PATH="./bin" SERVICE_NAME="gateway" build

RUN chmod +x ./bin/gateway
RUN chmod +x ./bin/gateway-health-checker

FROM scratch

WORKDIR /app

COPY --from=builder /app/bin/gateway-health-checker .
COPY --from=builder /app/bin/gateway .

CMD [ "/app/bin/gateway" ]
