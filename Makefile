SOURCE = $(wildcard *.go)
ALL = \
	$(foreach arch,32 64,\
	$(foreach suffix,win.exe linux osx,\
		gostatic-$(arch)-$(suffix)))

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
win.exe = windows
osx = darwin
gostatic-64-%: $(SOURCE)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 go build -o $@

gostatic-32-%: $(SOURCE)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 go build -o $@

upload: $(ALL)
ifndef UPLOAD_PATH
	@echo "Define UPLOAD_PATH to determine where files should be uploaded"
else
	rsync -P $(ALL) $(UPLOAD_PATH)
endif
