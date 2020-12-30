APPNAME:=ard2rss
RELEASE?=0

ifeq ($(RELEASE), 1)
	# Strip debug information from the binary
	GO_LDFLAGS+=-s -w
endif
GO_LDFLAGS:=-ldflags="$(GO_LDFLAGS)"

.PHONY: default
default: test

.PHONY: build
build:
	go build $(GO_LDFLAGS) -o ./build/$(APPNAME) -v cmd/main.go

.PHONY: test
test: build
	go test -v -race ./...

.PHONY: clean
clean:
	rm -rf ./build