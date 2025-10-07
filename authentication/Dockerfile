FROM golang:1.24-trixie as builder

COPY modules/migrator/ /modules/migrator/
COPY users/ /users

WORKDIR /app

COPY authentication/ ./

RUN make OUTPUT_PATH="./bin" SERVICE_NAME="authentication" build

RUN chmod +x ./bin/authentication
RUN chmod +x ./bin/authentication-health-checker

FROM scratch

WORKDIR /app

COPY --from=builder /app/bin/authentication-health-checker bin/authentication-health-checker
COPY --from=builder /app/bin/authentication bin/authentication
# These files are copied via k8s secrets mounted as volume
COPY --from=builder /app/private_key.pem /etc/jwt-keys/private_key.pem
COPY --from=builder /app/public_key.pem /etc/jwt-keys/public_key.pem
COPY authentication/migrations ./migrations

CMD [ "/app/bin/authentication" ]
