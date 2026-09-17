.PHONY: lint lint-fix
.PHONY: test-all test-binary-codec test-address-codec test-keypairs test-xrpl test-ci
.PHONY: run-localnet run-localnet-linux/amd64 run-localnet-linux/arm64 stop-localnet integration-localnet
.PHONY: test-integration-localnet test-integration-localnet-ci test-integration-devnet test-integration-testnet
.PHONY: test-integration-confidential-localnet test-integration-confidential-devnet
.PHONY: coverage-unit coverage-unit-ci test-report-summary benchmark
.PHONY: test-confidential update-mpt-crypto
.PHONY: update-definitions

UNIT_TEST_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces | grep -v /confidential) ./xrpl/testutil/integration/...
EXCLUDED_TEST_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces | grep -v /confidential)

# Devnet and testnet leave the confidential suite out because it funds a fresh set of
# accounts per scenario through a shared faucet and takes an order of magnitude longer
# there. It has its own devnet target for the occasions that are worth the wait.
INTEGRATION_TEST_PACKAGES = $(shell go list ./xrpl/transaction/integration/... | grep -v /confidential)
# Localnet closes ledgers on demand and funds from the genesis account, so it runs the
# whole suite, confidential included. That is the run CI makes on every change, and it
# is what keeps the CGo transaction path covered end to end.
LOCALNET_INTEGRATION_TEST_PACKAGES = ./xrpl/transaction/integration/...
CONFIDENTIAL_INTEGRATION_TEST_PACKAGES = ./xrpl/transaction/integration/confidential/...

PARALLEL_TESTS = 4
TEST_TIMEOUT = 5m
CONFIDENTIAL_TEST_TIMEOUT ?= 60m
UNIT_TEST_REPORT ?= unit-test-results.json
INTEGRATION_TEST_REPORT ?= localnet-test-results.json
COVERAGE_PROFILE ?= coverage.out
COVERAGE_HTML ?= coverage.html
TEST_REPORT ?=
TEST_REPORT_TITLE ?= Test report

GOTEST := $(shell command -v gotest 2>/dev/null || echo "go test")

GOLANGCI_LINT_MAJOR_VERSION = 2
GOLANGCI_LINT_VERSION = v2.11.3

XRPLD_IMAGE ?= rippleci/xrpld:develop
XRPLD_CONFIG ?= /etc/xrpld/xrpld.cfg
LOCALNET_CONTAINER ?= xrpld_standalone
LOCALNET_LEDGER_INTERVAL ?= 0.1

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
	@$(GOTEST) $(UNIT_TEST_PACKAGES)
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
	@$(GOTEST) $(UNIT_TEST_PACKAGES) -parallel $(PARALLEL_TESTS) -timeout $(TEST_TIMEOUT)
	@echo "Tests complete!"

run-localnet: run-localnet-linux/amd64

run-localnet-linux/amd64:
	@echo "Running localnet..."
	@docker run --rm -d --platform linux/amd64 -p 5005:5005 -p 6006:6006 --name $(LOCALNET_CONTAINER) --volume $(PWD)/.ci-config/xrpld.cfg:$(XRPLD_CONFIG):ro --entrypoint bash $(XRPLD_IMAGE) -c 'mkdir -p /var/lib/xrpld/db/ && xrpld --conf $(XRPLD_CONFIG) -a --start & while true; do xrpld --conf $(XRPLD_CONFIG) ledger_accept; sleep $(LOCALNET_LEDGER_INTERVAL); done'
	@echo "Localnet running!"

run-localnet-linux/arm64:
	@echo "Running localnet..."
	@docker run --rm -d --platform linux/arm64 -p 5005:5005 -p 6006:6006 --name $(LOCALNET_CONTAINER) --volume $(PWD)/.ci-config/xrpld.cfg:$(XRPLD_CONFIG):ro --entrypoint bash $(XRPLD_IMAGE) -c 'mkdir -p /var/lib/xrpld/db/ && xrpld --conf $(XRPLD_CONFIG) -a --start & while true; do xrpld --conf $(XRPLD_CONFIG) ledger_accept; sleep $(LOCALNET_LEDGER_INTERVAL); done'
	@echo "Localnet running!"

stop-localnet:
	@docker rm --force $(LOCALNET_CONTAINER) >/dev/null 2>&1 || true

integration-localnet:
	@./scripts/localnet-integration.sh

# CGO_ENABLED is explicit on both localnet targets because the confidential scenarios
# are compiled out rather than failed when it is off, which would drop them silently.
test-integration-localnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@env INTEGRATION=localnet CGO_ENABLED=1 $(GOTEST) -tags integration_localnet -p 1 $(LOCALNET_INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-localnet-ci:
	@echo "Running Go localnet integration tests with structured output..."
	@go clean -testcache
	@env INTEGRATION=localnet CGO_ENABLED=1 go test -json -tags integration_localnet -p 1 -timeout $(TEST_TIMEOUT) $(LOCALNET_INTEGRATION_TEST_PACKAGES) > "$(INTEGRATION_TEST_REPORT)" || { cat "$(INTEGRATION_TEST_REPORT)"; false; }
	@cat "$(INTEGRATION_TEST_REPORT)"

test-integration-devnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@env INTEGRATION=devnet $(GOTEST) $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-testnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@env INTEGRATION=testnet $(GOTEST) $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
	@echo "Tests complete!"

# The confidential MPT integration tests are kept out of the targets above because they
# link CGo and run roughly an order of magnitude longer.
test-integration-confidential-localnet:
	@echo "Running confidential MPT integration tests on localnet (CGo required)..."
	@go clean -testcache
	@env INTEGRATION=localnet CGO_ENABLED=1 $(GOTEST) -tags integration_localnet -p 1 $(CONFIDENTIAL_INTEGRATION_TEST_PACKAGES) -timeout $(CONFIDENTIAL_TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-confidential-devnet:
	@echo "Running confidential MPT integration tests on devnet (CGo required)..."
	@go clean -testcache
	@env INTEGRATION=devnet CGO_ENABLED=1 $(GOTEST) -p 1 $(CONFIDENTIAL_INTEGRATION_TEST_PACKAGES) -timeout $(CONFIDENTIAL_TEST_TIMEOUT) -v
	@echo "Tests complete!"

coverage-unit:
	@echo "Generating unit test coverage report..."
	@$(GOTEST) -coverprofile=coverage.out $(UNIT_TEST_PACKAGES)
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Multi-package JSON tests can emit duplicate set-mode blocks, so this target
# merges each block by its highest covered value before generating reports.
coverage-unit-ci:
	@echo "Generating unit test coverage with structured output..."
	@go clean -testcache
	@go test -json -covermode=set -coverprofile="$(COVERAGE_PROFILE)" -timeout $(TEST_TIMEOUT) $(UNIT_TEST_PACKAGES) > "$(UNIT_TEST_REPORT)" || { cat "$(UNIT_TEST_REPORT)"; false; }
	@cat "$(UNIT_TEST_REPORT)"
	@awk 'NR == 1 { print; next } { key = $$1 " " $$2; if (!(key in seen)) { order[++count] = key; seen[key] = 1; covered[key] = $$3 } else if ($$3 > covered[key]) { covered[key] = $$3 } } END { for (i = 1; i <= count; i++) print order[i], covered[order[i]] }' "$(COVERAGE_PROFILE)" > "$(COVERAGE_PROFILE).tmp"
	@mv "$(COVERAGE_PROFILE).tmp" "$(COVERAGE_PROFILE)"
	@go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_HTML)"
	@echo "Coverage report generated at $(COVERAGE_HTML)"

test-report-summary:
	@sh scripts/summarize-go-test-report.sh "$(TEST_REPORT)" "$(TEST_REPORT_TITLE)"

benchmark:
	@echo "Running Go benchmarks..."
	@$(GOTEST) -bench=. $(EXCLUDED_TEST_PACKAGES)
	@echo "Benchmarks complete!"

################################################################################
######################### CONFIDENTIAL MPT #####################################
################################################################################

test-confidential:
	@echo "Running confidential MPT tests (CGo required)..."
	@CGO_ENABLED=1 go test ./confidential/... -v -timeout $(TEST_TIMEOUT)
	@echo "Confidential tests complete!"

update-mpt-crypto:
	@bash confidential/deps/update.sh

################################################################################
######################### PROTOCOL DEFINITIONS #################################
################################################################################

# Refreshes binary-codec/definitions/definitions.json from a node's
# server_definitions response. Override the node with NODE_URL=<url>.
update-definitions:
	@bash scripts/update-definitions.sh $(if $(NODE_URL),--node "$(NODE_URL)")
