FROM golang:alpine3.10 as builder
RUN mkdir /build
ADD cleanup.go /build
WORKDIR /build
RUN go build -o cleanup .
FROM alpine:3.10.3
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/cleanup /app/
WORKDIR /app
ENTRYPOINT ["./cleanup"]
