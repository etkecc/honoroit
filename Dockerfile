FROM registry.gitlab.com/etke.cc/base AS builder

WORKDIR /honoroit
COPY . .
RUN make build

FROM alpine:latest

ENV HONOROIT_DB_DSN /data/honoroit.db

RUN apk --no-cache add ca-certificates tzdata olm && update-ca-certificates && \
    adduser -D -g '' honoroit && \
    mkdir /data && chown -R honoroit /data

COPY --from=builder /honoroit/honoroit /opt/honoroit/honoroit

WORKDIR /opt/honoroit
USER honoroit

ENTRYPOINT ["/opt/honoroit/honoroit"]

