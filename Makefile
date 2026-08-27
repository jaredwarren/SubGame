MODULE  := github.com/jaredwarren/SubGame
BINARY  := game
MAIN    := ./cmd/game
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null)

.PHONY: all build run test test-v vet lint clean tidy check audio gen-audio tools wasm serve pages-check

all: check lint build

## build: compile the game binary
build:
	go build -o $(BINARY) $(MAIN)

## audio: generate or regenerate all procedural DSP audio assets
audio:
	go run ./cmd/gen_audio

## gen-audio: alias for audio target
gen-audio: audio

## tools: launch the local web devtools hub (world / caves / audio)
tools:
	go run ./cmd/devtools

## run: build and launch the game (must be run from repo root so assets resolve)
run: build
	./$(BINARY)

## wasm: build the WebAssembly bundle into web/ (game.wasm + wasm_exec.js)
wasm:
	GOOS=js GOARCH=wasm go build -o web/game.wasm $(MAIN)
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
	cp StartBackground.jpeg web/ 2>/dev/null || true

## serve: build wasm and serve web/ on :8080 for browser/phone testing
serve: wasm
	go run ./cmd/webserve

## pages-check: verify web/ has everything GitHub Pages needs after `make wasm`
pages-check: wasm
	@test -f web/index.html
	@test -f web/game.wasm
	@test -f web/wasm_exec.js
	@echo "OK — push to main (with Pages source = GitHub Actions) deploys to:"
	@echo "     https://jaredwarren.github.io/SubGame/"

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

## clean: remove the compiled binary and wasm bundle
clean:
	rm -f $(BINARY) web/game.wasm web/wasm_exec.js web/StartBackground.jpeg

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
