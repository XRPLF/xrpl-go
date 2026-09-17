#!/bin/sh

set -eu

CONTAINER_NAME="${LOCALNET_CONTAINER:-xrpld_standalone}"
MAX_WAIT_SECONDS="${MAX_WAIT_SECONDS:-60}"
WAIT_ATTEMPTS_PER_SECOND=4
RPC_URL="${RPC_URL:-http://127.0.0.1:5005}"
STARTED_LOCALNET=0

cleanup() {
	if [ "$STARTED_LOCALNET" -eq 1 ] && [ "${KEEP_LOCALNET:-0}" != "1" ]; then
		echo "Stopping localnet..."
		docker rm --force "$CONTAINER_NAME" >/dev/null 2>&1 || true
	fi
}

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "Missing required command: $1" >&2
		exit 1
	fi
}

wait_for_localnet() {
	attempt=0
	max_attempts=$((MAX_WAIT_SECONDS * WAIT_ATTEMPTS_PER_SECOND))
	while [ "$attempt" -lt "$max_attempts" ]; do
		if curl -fsS \
			-H "Content-Type: application/json" \
			-d '{"method":"server_info","params":[{}]}' \
			"$RPC_URL" >/dev/null 2>&1; then
			return 0
		fi

		sleep 0.25
		attempt=$((attempt + 1))
	done

	echo "Localnet RPC did not become ready within ${MAX_WAIT_SECONDS}s." >&2
	docker logs "$CONTAINER_NAME" 2>&1 || true
	return 1
}

require_command curl
require_command docker
require_command make

trap cleanup EXIT INT TERM

if docker ps -q -f "name=^/${CONTAINER_NAME}$" | grep -q .; then
	echo "Using existing localnet container: $CONTAINER_NAME"
else
	echo "Starting localnet..."
	make run-localnet
	STARTED_LOCALNET=1
fi

echo "Waiting for localnet RPC at $RPC_URL..."
wait_for_localnet

echo "Running localnet integration tests..."
if [ -n "${INTEGRATION_TEST_REPORT:-}" ]; then
	make test-integration-localnet-ci
else
	make test-integration-localnet
fi
