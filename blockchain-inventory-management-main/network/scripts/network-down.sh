#!/bin/bash
CDIR="$(cd "$(dirname "$0")" && pwd)"
cd "$CDIR/.."

# Prevent Git Bash from converting path arguments to Windows format on docker commands
export MSYS_NO_PATHCONV=1

# Determine the Docker Compose command to use
if command -v docker-compose >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    echo "Error: Neither 'docker-compose' nor 'docker compose' was found. Please install Docker Compose." >&2
    exit 1
fi

echo "Stopping Fabric network..."
$DOCKER_COMPOSE -f docker-compose-inventory-net.yaml down -v

echo "Cleaning generated config and crypto files..."
rm -rf crypto-config
rm -rf channel-artifacts
rm -rf system-genesis-block
rm -f inventorychannel.block
rm -f assetcc.tar.gz

echo "Network stopped & cleaned."
