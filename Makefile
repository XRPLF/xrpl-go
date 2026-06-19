.PHONY: lint test benchmark

EXCLUDED_TEST_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces)
EXCLUDED_COVERAGE_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces)

INTEGRATION_TEST_PACKAGES = ./xrpl/transaction/integration/...

PARALLEL_TESTS = 4
TEST_TIMEOUT = 5m

GOTEST := $(shell command -v gotest 2>/dev/null || echo "go test")

GOLANGCI_LINT_MAJOR_VERSION = 2
GOLANGCI_LINT_VERSION = v2.11.3

RIPPLED_IMAGE ?= rippleci/xrpld:develop
LOCALNET_CONTAINER ?= rippled_standalone

################################################################################
############################### LINTING ########################################
################################################################################

lint:
	@echo "Linting Go code..."
	@go install github.com/golangci/golangci-lint/v$(GOLANGCI_LINT_MAJOR_VERSION)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@golangci-lint run
	@echo "Linting complete!"

lint-fix:
	@echo "Fixing Go code..."
	@go install github.com/golangci/golangci-lint/v$(GOLANGCI_LINT_MAJOR_VERSION)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@golangci-lint run --fix
	@echo "Fixing complete!"

################################################################################
############################### TESTING ########################################
################################################################################

test-all:
	@echo "Running Go tests..."
	@$(GOTEST) $(EXCLUDED_TEST_PACKAGES)
	@echo "Tests complete!"

test-binary-codec:
	@echo "Running Go tests for binary codec package..."
	@$(GOTEST) ./binary-codec/...
	@echo "Tests complete!"

test-address-codec:
	@echo "Running Go tests for address codec package..."
	@$(GOTEST) ./address-codec/...
	@echo "Tests complete!"

test-keypairs:
	@echo "Running Go tests for keypairs package..."
	@$(GOTEST) ./keypairs/...
	@echo "Tests complete!"

test-xrpl:
	@echo "Running Go tests for xrpl package..."
	@$(GOTEST) ./xrpl/...
	@echo "Tests complete!"

test-ci:
	@echo "Running Go tests..."
	@go clean -testcache
	@$(GOTEST) $(EXCLUDED_TEST_PACKAGES) -parallel $(PARALLEL_TESTS) -timeout $(TEST_TIMEOUT)
	@echo "Tests complete!"

run-localnet: run-localnet-linux/amd64

run-localnet-linux/amd64:
	@echo "Running localnet..."
	@docker run --rm -d --platform linux/amd64 -p 5005:5005 -p 6006:6006 --name $(LOCALNET_CONTAINER) --volume $(PWD)/.ci-config:/etc/opt/ripple/ --entrypoint bash $(RIPPLED_IMAGE) -c 'mkdir -p /var/lib/xrpld/db/ && rippled -a --start & sleep 5 && while true; do rippled ledger_accept; sleep 1; done'
	@echo "Localnet running!"

run-localnet-linux/arm64:
	@echo "Running localnet..."
	@docker run --rm -d --platform linux/arm64 -p 5005:5005 -p 6006:6006 --name $(LOCALNET_CONTAINER) --volume $(PWD)/.ci-config:/etc/opt/ripple/ --entrypoint bash $(RIPPLED_IMAGE) -c 'mkdir -p /var/lib/xrpld/db/ && rippled -a --start & sleep 5 && while true; do rippled ledger_accept; sleep 1; done'
	@echo "Localnet running!"

stop-localnet:
	@docker stop $(LOCALNET_CONTAINER) >/dev/null 2>&1 || true

integration-localnet:
	@./scripts/localnet-integration.sh

test-integration-localnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@INTEGRATION=localnet $(GOTEST) -p 1 $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-devnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@INTEGRATION=devnet $(GOTEST) $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-testnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@INTEGRATION=testnet $(GOTEST) $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

coverage-unit:
	@echo "Generating unit test coverage report..."
	@$(GOTEST) -coverprofile=coverage.out $(EXCLUDED_COVERAGE_PACKAGES)
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

benchmark:
	@echo "Running Go benchmarks..."
	@$(GOTEST) -bench=. ./...
	@echo "Benchmarks complete!"
