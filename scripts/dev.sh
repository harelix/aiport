#!/bin/sh
# Development loop for aiport (make dev). No extra tools needed.
#
#   Go:        rebuilds and restarts `aiport serve` when a .go file or the
#              embedded dashboard (internal/dashboard/dist, after
#              `make web-build`) changes. Not on the config: the gateway
#              applies its own Changes live, and a restart would drop waiting
#              pairing requests and tokens not yet collected. Restart by hand
#              after editing the config file yourself. Polls once a second: Go has no in-process hot
#              reload, so a restart is the reliable equivalent (~1-2 s).
#   Dashboard: Vite on http://127.0.0.1:5173 reloads the page on every edit
#              (CSS updates in place) and proxies /api to the gateway.
#
# Env: AIPORT_CONFIG (default aiport.yaml), AIPORT_DEV_PORT (default 5173).
set -u
cd "$(dirname "$0")/.." || exit 1
CONFIG=${AIPORT_CONFIG:-aiport.yaml}
export CGO_ENABLED=0 GOFLAGS=-mod=vendor
BIN=bin/aiport-dev
STAMP=$(mktemp "${TMPDIR:-/tmp}/aiport-dev.XXXXXX")
SERVER=""
VITE=""

stop_server() {
	[ -n "$SERVER" ] && kill "$SERVER" 2>/dev/null && wait "$SERVER" 2>/dev/null
	SERVER=""
}
cleanup() {
	stop_server
	[ -n "$VITE" ] && kill "$VITE" 2>/dev/null
	rm -f "$STAMP"
	exit 0
}
trap cleanup INT TERM

build_and_start() {
	stop_server
	printf '\n\033[36m[dev] building gateway…\033[0m\n'
	if go build -o "$BIN" ./cmd/aiport; then
		"$BIN" serve --config "$CONFIG" &
		SERVER=$!
	else
		printf '\033[33m[dev] build failed; fix the error and save again\033[0m\n'
	fi
}

if command -v pnpm >/dev/null 2>&1 && [ -d web/node_modules ]; then
	(cd web && exec pnpm exec vite) &
	VITE=$!
	printf '\033[36m[dev] dashboard with live reload: http://127.0.0.1:%s\033[0m\n' "${AIPORT_DEV_PORT:-5173}"
else
	printf '\033[33m[dev] web dependencies not installed (pnpm --dir web install); serving the embedded dashboard only\033[0m\n'
fi

touch "$STAMP"
build_and_start
while :; do
	sleep 1
	changed=$(find cmd internal go.mod -newer "$STAMP" \( -name '*.go' -o -name go.mod -o -path 'internal/dashboard/dist/*' \) 2>/dev/null | head -1)
	if [ -n "$changed" ]; then
		touch "$STAMP"
		printf '\033[36m[dev] %s changed\033[0m\n' "$changed"
		build_and_start
	fi
done
