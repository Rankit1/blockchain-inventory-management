#!/bin/bash
set -e

# Prevent Git Bash from converting path arguments to Windows format on docker commands
export MSYS_NO_PATHCONV=1

VERSION=${1:-"1.0"}
SEQUENCE=${2:-"1"}

CDIR="$(cd "$(dirname "$0")" && pwd)"
cd "$CDIR/.."

echo "========================================================="
echo "Packaging and Deploying assetcc Go Chaincode (Version: $VERSION, Sequence: $SEQUENCE)"
echo "========================================================="

# 1. Package the chaincode
echo "Packaging chaincode..."
docker exec cli peer lifecycle chaincode package assetcc.tar.gz \
  --path /opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/assetcc \
  --lang golang \
  --label "assetcc_${VERSION}"

# 2. Install on Lab peer
echo "Installing on peer0.lab..."
docker exec cli peer lifecycle chaincode install assetcc.tar.gz

# 3. Install on Admin peer
echo "Installing on peer0.admin..."
docker exec \
  -e CORE_PEER_ADDRESS=peer0.admin.inventory.com:8051 \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.admin.inventory.com/tls/ca.crt \
  cli peer lifecycle chaincode install assetcc.tar.gz

# 4. Install on Store peer
echo "Installing on peer0.store..."
docker exec \
  -e CORE_PEER_ADDRESS=peer0.store.inventory.com:9051 \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.store.inventory.com/tls/ca.crt \
  cli peer lifecycle chaincode install assetcc.tar.gz

# 5. Install on IT peer
echo "Installing on peer0.it..."
docker exec \
  -e CORE_PEER_ADDRESS=peer0.it.inventory.com:10051 \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.it.inventory.com/tls/ca.crt \
  cli peer lifecycle chaincode install assetcc.tar.gz

# 6. Retrieve package ID
echo "Retrieving Package ID..."
PACKAGE_ID=$(docker exec cli peer lifecycle chaincode calculatepackageid assetcc.tar.gz)
echo "Chaincode Package ID: $PACKAGE_ID"

# 7. Approve for Org
echo "Approving for InventoryOrg..."
docker exec cli peer lifecycle chaincode approveformyorg \
  -o orderer.inventory.com:7050 \
  --channelID inventorychannel \
  --name assetcc \
  --version "$VERSION" \
  --package-id "$PACKAGE_ID" \
  --sequence "$SEQUENCE" \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/inventory.com/orderers/orderer.inventory.com/tls/ca.crt

# 8. Commit chaincode definition (specifying at least two peers to submit endorsement proofs)
echo "Committing chaincode..."
docker exec cli peer lifecycle chaincode commit \
  -o orderer.inventory.com:7050 \
  --channelID inventorychannel \
  --name assetcc \
  --version "$VERSION" \
  --sequence "$SEQUENCE" \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/inventory.com/orderers/orderer.inventory.com/tls/ca.crt \
  --peerAddresses peer0.lab.inventory.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.lab.inventory.com/tls/ca.crt \
  --peerAddresses peer0.admin.inventory.com:8051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.admin.inventory.com/tls/ca.crt

echo "========================================================="
echo "assetcc Go Chaincode deployed successfully!"
echo "========================================================="
