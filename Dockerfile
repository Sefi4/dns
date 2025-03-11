ARG IG_VERSION=v0.38.0

FROM ghcr.io/inspektor-gadget/ig:${IG_VERSION} AS ig
FROM golang:1.23 AS builder
ARG IG_VERSION

WORKDIR /app
COPY . .
COPY --from=ig /usr/bin/ig /usr/bin/ig

RUN ig image pull ghcr.io/inspektor-gadget/gadget/trace_dns:${IG_VERSION} && \
    ig image export trace_dns:${IG_VERSION} trace_dns.tar && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/trace_dns.tar .
COPY --from=builder /app/myapp .

EXPOSE 8080

CMD ["./myapp"]
