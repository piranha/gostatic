SOURCE = $(shell find . -name '*.go')
TAG ?= $(shell git describe --tags)
GOBUILD = go build -ldflags '-w'

ALL = \
	$(foreach suffix,linux osx win.exe,\
		build/gostatic-64-$(suffix))

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

# NOTE: first push a tag, then make release!
ifndef desc
release:
	@echo "You forgot description! Run it as 'make release desc=tralala'"
else
release: $(ALL)
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
