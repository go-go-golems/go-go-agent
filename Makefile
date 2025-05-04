.PHONY: gifs generate-proto install-proto-tools clean

all: gifs generate-proto

VERSION=v0.1.14

TAPES=$(shell ls doc/vhs/*tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v

lint:
	golangci-lint run -v

test:
	go test ./...

build:
	go generate ./...
	go build ./...

goreleaser:
	goreleaser release --skip=sign --snapshot --clean

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/go-go-agent@$(shell svu current)

bump-glazed:
	go get github.com/go-go-golems/glazed@latest
	go get github.com/go-go-golems/clay@latest
	go get github.com/go-go-golems/geppetto@latest
	go get github.com/go-go-golems/go-emrichen@latest
	go get github.com/go-go-golems/pinocchio@latest
	go get github.com/go-go-golems/bobatea@latest
	go mod tidy

AGENT_BINARY=$(shell which agent)
install:
	go build -o ./dist/agent ./cmd/agent && \
		cp ./dist/agent $(AGENT_BINARY)

# Install required protobuf tools
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from proto files
generate-proto:
	mkdir -p pkg/events
	protoc --go_out=. --go_opt=paths=source_relative \
		proto/events.proto

# Clean generated files
clean:
	rm -rf pkg/events/*.pb.go
