lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

test-unit: lint
	mkdir -p reports/coverage
	go test ./... -coverprofile=reports/coverage/coverage.out

test-integration:
	mkdir -p reports/coverage
	go test ./... -race -tags=integration -coverprofile=reports/coverage/coverage.out

generate:
	go generate ./...
	rm -rf ./diagrams && \
		mkdir -p ./diagrams && \
		go run ./scripts/diagrams generate && \
		cd ./go-diagrams && \
		dot -Tpng sdk.dot > sdk.png && \
		mv sdk.png ../diagrams && \
		cd .. && \
		rm -rf ./go-diagrams