.ONESHELL:
.DEFAULT: help

.PHONY: help
help:
	@grep -E '^[a-z-]+:.*#' Makefile | \
		sort | \
		while read -r l; do printf "\033[1;32m$$(echo $$l | \
		cut -d':' -f1)\033[00m:$$(echo $$l | cut -d'#' -f2-)\n"; \
	done

.PHONY: test
test: # Run unit test suite
	go test -race -coverprofile=c.out .
	go tool cover -html=c.out -o=coverage.html

.PHONY: bench
bench: # Run benchmark test suite
	go test -bench .

.PHONY: format
format: # Run linter and formatters
	goimports -w -local github.com/miniscruff/scopie-go .
	golangci-lint run ./...

.PHONY: gen
gen:
	go test -bench . > BENCHMARKS.txt
