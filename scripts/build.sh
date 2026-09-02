#!/usr/bin/env bash
# Pull latest code (optional) and compile production binaries.
#
# VPS usage (after push to main):
#   cd /opt/ws && ./scripts/build.sh
#
# Environment:
#   GIT_BRANCH=main       branch to pull (default: main)
#   SKIP_GIT_PULL=1       skip git fetch/pull (CI or local compile only)
#   BIN_DIR=bin           output directory (default: bin)
#   GOOS=linux            target OS (default: linux)
#   GOARCH=amd64          target arch (default: amd64)

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# shellcheck source=scripts/lib/common.sh
source "${ROOT_DIR}/scripts/lib/common.sh"

require_cmd git
require_cmd go

BIN_DIR="${BIN_DIR:-bin}"
GIT_BRANCH="${GIT_BRANCH:-main}"

if [[ "${SKIP_GIT_PULL:-0}" != "1" ]]; then
	log "Pulling origin/${GIT_BRANCH}..."
	git fetch origin
	git checkout "$GIT_BRANCH"
	git pull --ff-only origin "$GIT_BRANCH"
else
	log "SKIP_GIT_PULL=1 — using current checkout"
fi

VERSION="$(git describe --tags --always --dirty 2>/dev/null || git rev-parse --short HEAD)"
LDFLAGS="-s -w"

log "Building api and migrate (version=${VERSION})..."
mkdir -p "$BIN_DIR"

CGO_ENABLED=0 GOOS="${GOOS:-linux}" GOARCH="${GOARCH:-amd64}" \
	go build -trimpath -ldflags "$LDFLAGS" -o "$BIN_DIR/api" ./cmd/api

CGO_ENABLED=0 GOOS="${GOOS:-linux}" GOARCH="${GOARCH:-amd64}" \
	go build -trimpath -ldflags "$LDFLAGS" -o "$BIN_DIR/migrate" ./cmd/migrate

log "Binaries written to ${BIN_DIR}/api and ${BIN_DIR}/migrate"
