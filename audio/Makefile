GO       ?= go
# tools/go.mod pins the developer tools (tool directives); both go.mod files require go 1.27.1, so GOTOOLCHAIN=auto
# switches older local toolchains automatically.
TOOLS     = $(GO) tool -modfile=tools/go.mod
COVERDIR ?= reports/coverage
FUZZTIME ?= 30s

.PHONY: all check ci lint fmt fmt-check vet test test-unit test-integration cover fuzz vuln tidy tidy-check bench bench-smoke build generate diagrams

all: check

## check: the fast local gate (tidiness, vet, lint, race tests)
check: tidy-check vet lint test

## ci: everything the CI workflow runs (../.github/workflows/audio.yml)
ci: tidy-check fmt-check vet lint test test-integration bench-smoke cover vuln

## lint: golangci-lint v2, pinned in tools/go.mod
lint:
	$(TOOLS) golangci-lint run ./...

## fmt: apply the configured formatters (gci, gofumpt)
fmt:
	$(TOOLS) golangci-lint fmt ./...

## fmt-check: show formatting diffs without rewriting files
fmt-check:
	$(TOOLS) golangci-lint fmt --diff ./...

vet:
	$(GO) vet ./...
	$(GO) vet -tags=integration ./...

## test: unit tests with the race detector, in random order
test:
	$(GO) test -race -shuffle=on ./...

# test-unit is kept as an alias of test for existing muscle memory.
test-unit: test

## test-integration: integration-tagged tests (//go:build integration), with their own coverage profile
test-integration:
	mkdir -p $(COVERDIR)
	$(GO) test -race -shuffle=on -tags=integration -coverpkg=./... -coverprofile=$(COVERDIR)/integration.out ./...

## cover: unit-test coverage across packages (-coverpkg), total printed last
cover:
	mkdir -p $(COVERDIR)
	$(GO) test -coverpkg=./... -coverprofile=$(COVERDIR)/unit.out ./...
	$(GO) tool cover -func=$(COVERDIR)/unit.out | tail -1

## fuzz: run every fuzz target for FUZZTIME each. Minimisation is capped at 1s: with the default (60s per new input) the
## large fixture seeds stall the fuzzer at 0 execs/sec.
fuzz:
	@for pkg in $$($(GO) list ./...); do \
		for target in $$($(GO) test -list '^Fuzz' $$pkg | grep '^Fuzz'); do \
			echo "== $$pkg $$target"; \
			$(GO) test -run '^$$' -fuzz "^$$target$$" -fuzztime $(FUZZTIME) -fuzzminimizetime 1s $$pkg || exit 1; \
		done; \
	done

vuln:
	$(TOOLS) govulncheck ./...

tidy:
	$(GO) mod tidy
	cd tools && $(GO) mod tidy

tidy-check:
	$(GO) mod tidy -diff
	cd tools && $(GO) mod tidy -diff

bench:
	$(GO) test -run '^$$' -bench . -benchmem ./...

## bench-smoke: run every benchmark once, so they keep compiling and passing their own checks
bench-smoke:
	$(GO) test -run '^$$' -bench . -benchtime 1x ./...

build:
	$(GO) build ./...

generate:
	$(GO) generate ./...

## diagrams: regenerate diagrams/*.dot and render them to PNG (needs docker)
diagrams:
	$(GO) run ./scripts/diagrams generate
	docker run --rm -v ./diagrams:/diagrams -w /diagrams docker.io/zalgonoise/graphviz:ubuntu sh -c \
		'dot -Tpng sdk.dot > sdk.png && dot -Tpng encoding.dot > encoding.png'
