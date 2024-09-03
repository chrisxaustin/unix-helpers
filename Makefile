.DEFAULT_GOAL := build
.PHONY:fmt vet build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

check: vet
	staticcheck ./...

build: check
	go build ./...

clean:
	go clean ./...

test:
	go test ./...

doc:
	godoc -http=:8080