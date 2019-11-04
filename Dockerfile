FROM golang:alpine as builder
RUN mkdir /build
ADD . /build
WORKDIR /build
RUN go build -o cleanup .
FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/cleanup /app/
WORKDIR /app
ENTRYPOINT ["./cleanup"]
