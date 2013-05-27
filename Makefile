SOURCE = $(wildcard *.go)
TAG = $(shell git describe --tags)
# $(tag) here will contain either `-1.0-` or just `-`
ALL = \
	$(foreach arch,32 64,\
    $(foreach tag,-$(TAG)- -,\
	$(foreach suffix,win.exe linux osx,\
		build/gostatic$(tag)$(arch)-$(suffix))))

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
build/gostatic-$(TAG)-64-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 go build -o $@

build/gostatic-$(TAG)-32-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 go build -o $@

build/gostatic-%: build/gostatic-$(TAG)-%
	@mkdir -p $(@D)
	cd $(@D) && ln -sf $(<F) $(@F)

upload: $(ALL)
ifndef UPLOAD_PATH
	@echo "Define UPLOAD_PATH to determine where files should be uploaded"
else
	rsync -l -P $(ALL) $(UPLOAD_PATH)
endif
