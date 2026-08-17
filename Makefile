MODULE  := github.com/jaredwarren/SubGame
BINARY  := game
MAIN    := ./cmd/game
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null)

.PHONY: all build run test test-v vet lint clean tidy check audio gen-audio

all: check lint build

## build: compile the game binary
build:
	go build -o $(BINARY) $(MAIN)

## audio: generate or regenerate all procedural DSP audio assets
audio:
	go run ./cmd/gen_audio

## gen-audio: alias for audio target
gen-audio: audio

## run: build and launch the game (must be run from repo root so assets resolve)
run: build
	./$(BINARY)

## test: run all tests
test:
	go test ./...

## test-v: run all tests with verbose output
test-v:
	go test -v ./...

## vet: run go vet
vet:
	go vet ./...

## lint: run golangci-lint (must be installed: https://golangci-lint.run/usage/install/)
lint:
ifndef GOLANGCI_LINT
	$(error golangci-lint not found — install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
endif
	golangci-lint run ./...

## check: vet + test (CI-style gate)
check: vet test

## tidy: tidy and verify module dependencies
tidy:
	go mod tidy
	go mod verify

## clean: remove the compiled binary
clean:
	rm -f $(BINARY)

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
