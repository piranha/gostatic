run:
	go run *.go test/config --summary

render:
	go run *.go test/config

config:
	go run *.go test/config --show-config

fmt:
	gofmt -w=true *.go
