FROM golang:1.24-trixie as builder

COPY modules/migrator/ /modules/migrator/

WORKDIR /app

COPY notification/ ./

RUN make build
RUN make build-health-checker

RUN chmod +x ./app-server
RUN chmod +x ./health-checker

FROM scratch

WORKDIR /app

# Necessary files for network operations in a scratch container:
#
# CA Certificates: To establish a secure TLS connection with smtp.gmail.com, your Go
#   binary needs to verify Gmail's certificate. This requires a bundle of trusted
#   Certificate Authority (CA) certificates, which are not present in scratch.
#
# DNS Resolution (when CGO is disabled): Go binaries built with CGO_ENABLED=0 (common for
#   scratch images) use Go's internal DNS resolver. This resolver relies on the
#   /etc/resolv.conf file to know which DNS servers to query. The scratch image lacks this
#   file.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/resolv.conf /etc/resolv.conf

COPY --from=builder /app/app-server .
COPY --from=builder /app/health-checker .
COPY notification/migrations migrations/
COPY notification/internal/templates internal/templates/

CMD [ "/app/app-server" ]
