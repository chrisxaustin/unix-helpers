.DEFAULT_GOAL := build
.PHONY:fmt vet build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

check: vet
	staticcheck ./...

build: check
	go build -o bin/ ./...

clean:
	go clean ./...

test:
	go test -race ./...

doc:
	pkgsite