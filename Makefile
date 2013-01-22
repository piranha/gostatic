SOURCE = $(wildcard *.go)
ALL = $(foreach suffix,windows.exe linux osx,gostatic-$(suffix))

all: $(ALL)

run:
	go run *.go test/config --summary

render:
	go run *.go test/config

config:
	go run *.go test/config --show-config

fmt:
	gofmt -w=true *.go

# os is determined as thus: if variable of suffix exists, it's taken, if not, then
# suffix itself is taken
windows.exe = windows
osx = darwin
gostatic-%: $(SOURCE)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 go build -o $@

upload: $(ALL)
	rsync -P $(ALL) $(UPLOAD_PATH)
