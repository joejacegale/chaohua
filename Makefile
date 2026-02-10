include .env
export

.PHONY: run
run:
	@go run main.go

.PHONY: test
test:
	@go test -v ./...