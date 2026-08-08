#!/bin/bash
set -e

# Make sure we are in the network directory
CDIR="$(cd "$(dirname "$0")" && pwd)"
cd "$CDIR/.."

# Prevent Git Bash from converting path arguments to Windows format on docker run
export MSYS_NO_PATHCONV=1

# Determine user flag for Docker (on Linux, use id -u:id -g to preserve file ownership; omit on Windows)
USER_FLAG=""
if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" && "$OSTYPE" != "mingw"* && "$OSTYPE" != "win32" ]]; then
    if command -v id >/dev/null 2>&1; then
        USER_FLAG="--user $(id -u):$(id -g)"
    fi
fi

# Determine the Docker Compose command to use
if command -v docker-compose >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    echo "Error: Neither 'docker-compose' nor 'docker compose' was found. Please install Docker Compose." >&2
    exit 1
fi

echo "========================================================="
echo "Starting Hyperledger Fabric Single-Org 4-Peer Network"
echo "========================================================="

# 1. Clean previous run state
./scripts/network-down.sh || true

# 2. Create directories
mkdir -p channel-artifacts
mkdir -p system-genesis-block

# 3. Generate crypto material
echo "Generating cryptographic certificates via Docker..."
docker run --rm $USER_FLAG -v "$(pwd):/network" -w /network hyperledger/fabric-tools:2.5.15 cryptogen generate --config=./crypto-config.yaml --output="crypto-config"

# 4. Generate system genesis block and channel configuration transactions via Docker
echo "Generating channel genesis and transactions via Docker..."
docker run --rm $USER_FLAG -v "$(pwd):/network" -w /network -e FABRIC_CFG_PATH=/network hyperledger/fabric-tools:2.5.15 configtxgen -profile InventoryOrdererGenesis -outputBlock ./system-genesis-block/genesis.block -channelID system-channel
docker run --rm $USER_FLAG -v "$(pwd):/network" -w /network -e FABRIC_CFG_PATH=/network hyperledger/fabric-tools:2.5.15 configtxgen -profile InventoryOrgChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID inventorychannel

# 5. Bring up docker containers
echo "Starting containers..."
$DOCKER_COMPOSE -f docker-compose-inventory-net.yaml up -d

echo "Waiting for containers to stabilize (10s)..."
sleep 10

# 6. Create the application channel
echo "Creating inventorychannel..."
docker exec cli peer channel create \
  -o orderer.inventory.com:7050 \
  -c inventorychannel \
  -f ./network/channel-artifacts/channel.tx \
  --outputBlock ./inventorychannel.block \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/inventory.com/orderers/orderer.inventory.com/tls/ca.crt

# 7. Join Lab Peer
echo "Joining peer0.lab to inventorychannel..."
docker exec cli peer channel join -b ./inventorychannel.block

# 8. Join Admin Peer
echo "Joining peer0.admin to inventorychannel..."
docker exec \
  -e CORE_PEER_ADDRESS=peer0.admin.inventory.com:8051 \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.admin.inventory.com/tls/ca.crt \
  cli peer channel join -b ./inventorychannel.block

# 9. Join Store Peer
echo "Joining peer0.store to inventorychannel..."
docker exec \
  -e CORE_PEER_ADDRESS=peer0.store.inventory.com:9051 \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.store.inventory.com/tls/ca.crt \
  cli peer channel join -b ./inventorychannel.block

# 10. Join IT Peer
echo "Joining peer0.it to inventorychannel..."
docker exec \
  -e CORE_PEER_ADDRESS=peer0.it.inventory.com:10051 \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.it.inventory.com/tls/ca.crt \
  cli peer channel join -b ./inventorychannel.block

echo "========================================================="
echo "Blockchain Network started & channel joined successfully!"
echo "========================================================="
