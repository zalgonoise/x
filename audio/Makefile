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
		docker run --rm -ti -v ./diagrams:/diagrams -w /diagrams docker.io/zalgonoise/graphviz:ubuntu sh -c 'dot -Tpng sdk.dot > sdk.png'