FROM ghcr.io/etkecc/base/build AS builder

WORKDIR /app
COPY . .
RUN just build

FROM scratch

ENV HONOROIT_DB_DSN /data/honoroit.db

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/honoroit /bin/honoroit

USER app

ENTRYPOINT ["/bin/honoroit"]

