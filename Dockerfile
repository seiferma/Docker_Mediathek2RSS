FROM --platform=$BUILDPLATFORM golang:alpine AS builder

ARG TARGETARCH
ARG GOOS=linux
ARG CGOENABLED=0

RUN apk add --no-cache make ca-certificates git
WORKDIR /go/src/app
COPY . .
RUN export GOARCH=$TARGETARCH && \
    make RELEASE=1 build test

FROM scratch
COPY --from=builder /go/src/app/build/mediathek2rss /opt/mediathek2rss
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENTRYPOINT ["/opt/mediathek2rss"]