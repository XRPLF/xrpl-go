.PHONY: lint test benchmark

EXCLUDED_TEST_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces)
EXCLUDED_COVERAGE_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces)

INTEGRATION_TEST_PACKAGES = ./xrpl/transaction/integration

PARALLEL_TESTS = 4
TEST_TIMEOUT = 5m

GOLANGCI_LINT_MAJOR_VERSION = 2
GOLANGCI_LINT_VERSION = v2.11.3

RIPPLED_IMAGE = rippleci/rippled:2.3.0

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
	@go test $(EXCLUDED_TEST_PACKAGES)
	@echo "Tests complete!"

test-binary-codec:
	@echo "Running Go tests for binary codec package..."
	@go test ./binary-codec/...
	@echo "Tests complete!"

test-address-codec:
	@echo "Running Go tests for address codec package..."
	@go test ./address-codec/...
	@echo "Tests complete!"

test-keypairs:
	@echo "Running Go tests for keypairs package..."
	@go test ./keypairs/...
	@echo "Tests complete!"

test-xrpl:
	@echo "Running Go tests for xrpl package..."
	@go test ./xrpl/...
	@echo "Tests complete!"

test-ci:
	@echo "Running Go tests..."
	@go clean -testcache
	@go test $(EXCLUDED_TEST_PACKAGES) -parallel $(PARALLEL_TESTS) -timeout $(TEST_TIMEOUT)
	@echo "Tests complete!"

run-localnet-linux/arm64:
	@echo "Running localnet..."
	@docker run -p 6006:6006 --rm -it -d --platform linux/arm64 --name rippled_standalone --volume $(PWD)/.ci-config:/etc/opt/ripple/ --entrypoint bash $(RIPPLED_IMAGE) -c 'rippled -a'
	@echo "Localnet running!"

run-localnet-linux/amd64:
	@echo "Running localnet..."
	@docker run -p 6006:6006 --rm -it -d --platform linux/amd64 --name rippled_standalone --volume $(PWD)/.ci-config:/etc/opt/ripple/ --entrypoint bash $(RIPPLED_IMAGE) -c 'rippled -a'
	@echo "Localnet running!"

test-integration-localnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@INTEGRATION=localnet go test $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-devnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@INTEGRATION=devnet go test $(INTEGRATION_TEST_PACKAGES)  -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-testnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@INTEGRATION=testnet go test $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

coverage-unit:
	@echo "Generating unit test coverage report..."
	@go test -coverprofile=coverage.out $(EXCLUDED_COVERAGE_PACKAGES)
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

benchmark:
	@echo "Running Go benchmarks..."
	@go test -bench=. ./...
	@echo "Benchmarks complete!"
