FROM golang:alpine3.10 as builder
RUN mkdir /build
ADD cleanup.go /build
WORKDIR /build
RUN go build -o cleanup .
RUN apk update && apk add curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.11.5/bin/linux/amd64/kubectl
RUN chmod a+x kubectl
FROM alpine:3.10.3
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/cleanup /app/
COPY --from=builder /build/kubectl /bin/
WORKDIR /app
ENTRYPOINT ["./cleanup"]
