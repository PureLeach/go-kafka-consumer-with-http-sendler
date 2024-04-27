.PHONY: build
build:
	go build -v ./cmd/events_consumer


.PHONY: start
start:
	./events_consumer


.PHONY: run
run:
	go run -v ./cmd/events_consumer/main.go


.PHONY: test
test:
	go test -v -race -count=1 -timeout 5s ./...

.DEFAULT_GOAL := build