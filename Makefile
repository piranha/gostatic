SOURCE = $(wildcard *.go)
ALL = $(foreach os,windows linux darwin,gostatic-$(os))

all: $(ALL)

run:
	go run *.go test/config --summary

render:
	go run *.go test/config

config:
	go run *.go test/config --show-config

fmt:
	gofmt -w=true *.go

gostatic-%: $(SOURCE)
	CGO_ENABLED=0 GOOS=$* GOARCH=amd64 go build -o $@
