#!/usr/bin/env bash
# Full production deploy on a VPS: pull, build, migrate, restart service.
#
# Typical flow:
#   1. Merge to main and push
#   2. SSH to the server
#   3. ./scripts/deploy.sh
#
# Environment:
#   ENV_FILE=/etc/ws/ws.env   production secrets (DATABASE_URL, HTTP_PORT, …)
#   SYSTEMD_SERVICE=ws-api    systemd unit to restart (optional)
#   GIT_BRANCH=main           passed through to build.sh
#   BIN_DIR=bin                 passed through to build.sh
#
# First-time server setup: see scripts/README.md

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# shellcheck source=scripts/lib/common.sh
source "${ROOT_DIR}/scripts/lib/common.sh"

ENV_FILE="${ENV_FILE:-/etc/ws/ws.env}"
BIN_DIR="${BIN_DIR:-bin}"

if [[ -f "$ENV_FILE" ]]; then
	log "Loading env from ${ENV_FILE}"
	set -a
	# shellcheck disable=SC1090
	source "$ENV_FILE"
	set +a
elif [[ -f "${ROOT_DIR}/.env" ]]; then
	log "WARN: ${ENV_FILE} not found — falling back to repo .env (not for production)"
	set -a
	# shellcheck disable=SC1091
	source "${ROOT_DIR}/.env"
	set +a
else
	die "No env file found. Create ${ENV_FILE} with DATABASE_URL and other vars."
fi

require_env DATABASE_URL

"${ROOT_DIR}/scripts/build.sh"

log "Running migrations..."
(cd "$ROOT_DIR" && DATABASE_URL="$DATABASE_URL" "${ROOT_DIR}/${BIN_DIR}/migrate" up)

if [[ -n "${SYSTEMD_SERVICE:-}" ]]; then
	require_cmd systemctl
	log "Restarting ${SYSTEMD_SERVICE}..."
	sudo systemctl restart "$SYSTEMD_SERVICE"
	sudo systemctl --no-pager status "$SYSTEMD_SERVICE"
else
	log "SYSTEMD_SERVICE not set — start or restart ${BIN_DIR}/api manually"
fi

log "Deploy complete."
