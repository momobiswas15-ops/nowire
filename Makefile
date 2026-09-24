.DEFAULT_GOAL := test
BINARY := nowire

.PHONY: fmt vet test race build install clean
fmt:
	gofmt -w cmd/ internal/

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -trimpath -ldflags='-s -w' -o $(BINARY) ./cmd/nowire

install:
	go install ./cmd/nowire

clean:
	rm -f $(BINARY)
