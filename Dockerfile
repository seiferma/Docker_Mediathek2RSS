FROM golang:alpine AS builder
RUN apk add --no-cache make gcc musl-dev
WORKDIR /go/src/app
COPY . .
RUN make build test

FROM golang:alpine
WORKDIR /opt
COPY --from=builder /go/src/app/build/ard2rss /opt/ard2rss
EXPOSE 8080
CMD ["./ard2rss"]