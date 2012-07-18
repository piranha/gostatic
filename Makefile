run:
	go run *.go test/config.json

fmt:
	gofmt -w=true *.go
