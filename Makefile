BINARY=openmarkers
MODULE=github.com/openmarkers/openmarkers-cli

.PHONY: build test lint install clean

build:
	go build -o $(BINARY) .

test:
	go test ./...

lint:
	go vet ./...

install:
	go install .

clean:
	rm -f $(BINARY)
