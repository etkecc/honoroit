FROM registry.gitlab.com/etke.cc/base/build AS builder

WORKDIR /honoroit
COPY . .
RUN just build

FROM registry.gitlab.com/etke.cc/base/app

ENV HONOROIT_DB_DSN /data/honoroit.db

COPY --from=builder /honoroit/honoroit /bin/honoroit

USER app

ENTRYPOINT ["/bin/honoroit"]

