SOURCE = $(wildcard *.go)
TAG ?= $(shell git describe --tags)
GOBUILD = go build -ldflags '-w'

# $(tag) here will contain either `-1.0-` or just `-`
ALL = \
	$(foreach arch,64 32,\
	$(foreach suffix,linux osx win.exe,\
		build/gostatic-$(arch)-$(suffix)))

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
build/gostatic-64-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD) -o $@

build/gostatic-32-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD) -o $@

release: $(ALL)
ifndef desc
	@echo "Run it as 'make release desc=tralala'"
else
	github-release release -u piranha -r gostatic -t "$(TAG)" -n "$(TAG)" --description '$(desc)'
	@for x in $(ALL); do \
		github-release upload -u piranha \
                              -r gostatic \
                              -t $(TAG) \
                              -f "$$x" \
                              -n "$$(basename $$x)" \
		&& echo "Uploaded $$x"; \
	done
endif
