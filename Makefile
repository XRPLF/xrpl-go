.PHONY: workspace lint lint-fix lint-confidential
.PHONY: test-all test-binary-codec test-address-codec test-keypairs test-xrpl test-ci
.PHONY: run-localnet run-localnet-linux/amd64 run-localnet-linux/arm64 stop-localnet integration-localnet
.PHONY: test-integration-localnet test-integration-localnet-ci test-integration-devnet test-integration-testnet
.PHONY: test-integration-confidential-localnet test-integration-confidential-devnet
.PHONY: coverage-unit coverage-unit-ci test-report-summary benchmark
.PHONY: test-confidential test-confidential-nocgo update-mpt-crypto
.PHONY: update-definitions

# Root ./... does not cross the confidential module boundary.
UNIT_TEST_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces) ./xrpl/testutil/integration/...
EXCLUDED_TEST_PACKAGES = $(shell go list ./... | grep -v /faucet | grep -v /examples | grep -v /testutil | grep -v /interfaces)

INTEGRATION_TEST_PACKAGES = ./xrpl/transaction/integration/...
# Localnet runs both modules. Public networks keep the slower confidential suite
# separate. Confidential paths below are relative to its module directory.
CONFIDENTIAL_INTEGRATION_TEST_PACKAGES = ./integration/...
INTEGRATION_MODULE ?= all
ifneq ($(INTEGRATION_MODULE),all)
ifneq ($(INTEGRATION_MODULE),core)
ifneq ($(INTEGRATION_MODULE),confidential)
$(error INTEGRATION_MODULE must be all, core, or confidential)
endif
endif
endif

PARALLEL_TESTS = 4
TEST_TIMEOUT = 5m
# Public networks advance ledger time in real time, including vault subscription waits.
PUBLICNET_INTEGRATION_TEST_TIMEOUT ?= 20m
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

# 3.4.0 development build, commit 21890d9d. Supports LendingProtocolV1_1 and fixCleanup3_4_0.
XRPLD_IMAGE ?= rippleci/xrpld@sha256:1f62f82d87794614881d7900748890ce60fb14576c3534227926a5dc3add250e
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

lint-confidential:
	@go install github.com/golangci/golangci-lint/v$(GOLANGCI_LINT_MAJOR_VERSION)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@cd confidential && golangci-lint run --config ../.golangci.yml

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

# Run make workspace first to test the two checked-out modules together.
# CGO_ENABLED=1 prevents the confidential scenarios from being silently omitted.
test-integration-localnet:
	@echo "Running localnet integration tests ($(INTEGRATION_MODULE))..."
	@go clean -testcache
ifneq ($(INTEGRATION_MODULE),confidential)
	@env INTEGRATION=localnet $(GOTEST) -tags integration_localnet -p 1 $(INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
endif
ifneq ($(INTEGRATION_MODULE),core)
	@cd confidential && env INTEGRATION=localnet CGO_ENABLED=1 $(GOTEST) -tags integration_localnet -p 1 $(CONFIDENTIAL_INTEGRATION_TEST_PACKAGES) -timeout $(TEST_TIMEOUT) -v
endif
	@echo "Tests complete!"

test-integration-localnet-ci:
	@echo "Running localnet integration tests ($(INTEGRATION_MODULE)) with structured output..."
	@go clean -testcache
	@status=0; : > "$(INTEGRATION_TEST_REPORT)"; \
		if [ "$(INTEGRATION_MODULE)" != confidential ]; then \
			env INTEGRATION=localnet go test -json -tags integration_localnet -p 1 -timeout $(TEST_TIMEOUT) $(INTEGRATION_TEST_PACKAGES) >> "$(INTEGRATION_TEST_REPORT)" || status=1; \
		fi; \
		if [ "$(INTEGRATION_MODULE)" != core ]; then \
			(cd confidential && env INTEGRATION=localnet CGO_ENABLED=1 go test -json -tags integration_localnet -p 1 -timeout $(TEST_TIMEOUT) $(CONFIDENTIAL_INTEGRATION_TEST_PACKAGES)) >> "$(INTEGRATION_TEST_REPORT)" || status=1; \
		fi; \
		cat "$(INTEGRATION_TEST_REPORT)"; exit $$status

test-integration-devnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@env INTEGRATION=devnet $(GOTEST) $(INTEGRATION_TEST_PACKAGES) -timeout $(PUBLICNET_INTEGRATION_TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-testnet:
	@echo "Running Go tests for integration package..."
	@go clean -testcache
	@env INTEGRATION=testnet $(GOTEST) $(INTEGRATION_TEST_PACKAGES) -timeout $(PUBLICNET_INTEGRATION_TEST_TIMEOUT) -v
	@echo "Tests complete!"

# Confidential-only integration targets run from the optional module.
test-integration-confidential-localnet:
	@echo "Running confidential MPT integration tests on localnet (CGo required)..."
	@go clean -testcache
	@cd confidential && env INTEGRATION=localnet CGO_ENABLED=1 $(GOTEST) -tags integration_localnet -p 1 $(CONFIDENTIAL_INTEGRATION_TEST_PACKAGES) -timeout $(CONFIDENTIAL_TEST_TIMEOUT) -v
	@echo "Tests complete!"

test-integration-confidential-devnet:
	@echo "Running confidential MPT integration tests on devnet (CGo required)..."
	@go clean -testcache
	@cd confidential && env INTEGRATION=devnet CGO_ENABLED=1 $(GOTEST) -p 1 $(CONFIDENTIAL_INTEGRATION_TEST_PACKAGES) -timeout $(CONFIDENTIAL_TEST_TIMEOUT) -v
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

# The workspace is local-only. The version-specific replacement also lets Go load
# the dependency graph before the declared minimum core version is published.
workspace:
	@if [ ! -f go.work ]; then GOWORK=off go work init . ./confidential; fi
	@GOWORK="$(CURDIR)/go.work" go work use . ./confidential
	@core_version=$$(awk '$$1 == "github.com/Peersyst/xrpl-go" { print $$2; exit }' confidential/go.mod); \
		GOWORK="$(CURDIR)/go.work" go work edit -replace="github.com/Peersyst/xrpl-go@$$core_version=."

test-confidential:
	@echo "Running confidential MPT tests (CGo required)..."
	@cd confidential && CGO_ENABLED=1 go test ./... -v -timeout $(TEST_TIMEOUT)
	@echo "Confidential tests complete!"

test-confidential-nocgo:
	@cd confidential && CGO_ENABLED=0 go build ./...
	@cd confidential && CGO_ENABLED=0 go test ./mptcrypto -timeout $(TEST_TIMEOUT)

update-mpt-crypto:
	@bash confidential/deps/update.sh

################################################################################
######################### PROTOCOL DEFINITIONS #################################
################################################################################

# Refreshes binary-codec/definitions/definitions.json from a node's
# server_definitions response. Override the node with NODE_URL=<url>.
update-definitions:
	@bash scripts/update-definitions.sh $(if $(NODE_URL),--node "$(NODE_URL)")
