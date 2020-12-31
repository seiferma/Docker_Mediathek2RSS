FROM golang:alpine AS builder
RUN apk add --no-cache make
WORKDIR /go/src/app
COPY . .
RUN make RELEASE=1 build test

FROM scratch
COPY --from=builder /go/src/app/build/ard2rss /opt/ard2rss
EXPOSE 8080
ENTRYPOINT ["/opt/ard2rss"]