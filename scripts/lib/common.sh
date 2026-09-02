#!/usr/bin/env bash
# Shared helpers for production scripts.

log() {
	printf '[%s] %s\n' "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" "$*"
}

die() {
	log "ERROR: $*"
	exit 1
}

require_env() {
	local name="$1"
	if [[ -z "${!name:-}" ]]; then
		die "$name is required (set in ${ENV_FILE:-/etc/ws/ws.env})"
	fi
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is not installed"
}
