.PHONY: build install clean test run-example check-deps

BINARY=corecut
VERSION=1.0.0

build:
	go build -o $(BINARY) -ldflags "-X main.Version=$(VERSION)" .

install: build
	sudo cp $(BINARY) /usr/local/bin/

clean:
	rm -f $(BINARY)
	rm -rf reports/

test:
	go test -v ./...

run-example: build
	chmod +x examples/*.sh examples/*.py
	./$(BINARY) run \
		--baseline ./examples/baseline.sh \
		--optimized ./examples/optimized.sh \
		--runs 5 \
		--warmup 1 \
		--alternate

run-throughput: build
	chmod +x examples/*.py
	./$(BINARY) run \
		--baseline ./examples/throughput_baseline.py \
		--optimized ./examples/throughput_optimized.py \
		--mode throughput \
		--runs 5 \
		--warmup 1

check-deps: build
	./$(BINARY) check-deps

aggregate: build
	./$(BINARY) aggregate ./reports/

# Development helpers
dev:
	go run . run --baseline ./examples/baseline.sh --optimized ./examples/optimized.sh --runs 3 --warmup 1

fmt:
	go fmt ./...

lint:
	golangci-lint run
