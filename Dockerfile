FROM golang:alpine3.10 as builder
ARG kubectl_version=v1.15.5
ARG helm_version=v2.13.1
RUN mkdir /build
WORKDIR /build
RUN apk update && apk add curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/${kubectl_version}/bin/linux/amd64/kubectl
RUN curl -LO https://get.helm.sh/helm-${helm_version}-linux-amd64.tar.gz
RUN tar xzf helm*.tar.gz
RUN chmod a+x kubectl linux-amd64/helm
ADD release_cleanup.go /build
RUN go build -o release_cleanup .
FROM alpine:3.10.3
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/release_cleanup /app/
COPY --from=builder /build/kubectl /bin/
COPY --from=builder /build/linux-amd64/helm /bin
WORKDIR /app
ENTRYPOINT ["./release_cleanup"]
