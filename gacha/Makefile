.PHONY: all test test-race test-integration cover bench bench-integration vet fmt lint tidy clean

GO ?= go
GOLANGCI_LINT ?= $(GO) tool golangci-lint
COVER_PROFILE ?= coverage.out

all: fmt vet lint test

test:
	$(GO) test -v ./...

test-race:
	$(GO) test -race -v ./...

test-integration:
	$(GO) test -race -tags integration -v ./...

cover:
	$(GO) test -race -coverprofile=$(COVER_PROFILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVER_PROFILE)

bench:
	$(GO) test -bench=. -benchmem -run=^$$ ./...

bench-integration:
	$(GO) test -tags integration -bench=. -benchmem -run=^$$ ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

lint:
	$(GOLANGCI_LINT) run ./...

tidy:
	$(GO) mod tidy

clean:
	$(GO) clean -testcache
	rm -f $(COVER_PROFILE)
