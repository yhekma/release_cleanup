FROM golang:alpine3.10 as builder
RUN mkdir /build
WORKDIR /build
RUN apk update && apk add curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.15.5/bin/linux/amd64/kubectl
RUN curl -LO https://get.helm.sh/helm-v2.13.1-linux-amd64.tar.gz
RUN tar xzf helm*.tar.gz
RUN chmod a+x kubectl linux-amd64/helm
ADD cleanup.go /build
RUN go build -o cleanup .
FROM alpine:3.10.3
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/cleanup /app/
COPY --from=builder /build/kubectl /bin/
COPY --from=builder /build/linux-amd64/helm /bin
WORKDIR /app
ENTRYPOINT ["./cleanup"]
