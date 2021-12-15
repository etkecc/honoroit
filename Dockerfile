FROM registry.gitlab.com/etke.cc/base AS builder

WORKDIR /honoroit
COPY . .
RUN make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && update-ca-certificates && \
    adduser -D -g '' honoroit && \
    mkdir /data && chown -R honoroit /data

COPY --from=builder /honoroit/honoroit /opt/honoroit/honoroit

WORKDIR /opt/honoroit
USER honoroit

ENTRYPOINT ["/opt/honoroit/honoroit"]

