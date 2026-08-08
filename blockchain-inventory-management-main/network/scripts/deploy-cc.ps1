$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location "$scriptDir/.."

try {
    Write-Host "========================================================="
    Write-Host "Packaging and Deploying assetcc Go Chaincode"
    Write-Host "========================================================="

    # 1. Package
    Write-Host "Packaging chaincode..."
    docker exec cli peer lifecycle chaincode package assetcc.tar.gz `
      --path /opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/assetcc `
      --lang golang `
      --label assetcc_1.0

    # 2. Install on Lab peer
    Write-Host "Installing on peer0.lab..."
    docker exec cli peer lifecycle chaincode install assetcc.tar.gz

    # 3. Install on Admin peer
    Write-Host "Installing on peer0.admin..."
    docker exec -e CORE_PEER_ADDRESS=peer0.admin.inventory.com:8051 -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.admin.inventory.com/tls/ca.crt cli peer lifecycle chaincode install assetcc.tar.gz

    # 4. Install on Store peer
    Write-Host "Installing on peer0.store..."
    docker exec -e CORE_PEER_ADDRESS=peer0.store.inventory.com:9051 -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.store.inventory.com/tls/ca.crt cli peer lifecycle chaincode install assetcc.tar.gz

    # 5. Install on IT peer
    Write-Host "Installing on peer0.it..."
    docker exec -e CORE_PEER_ADDRESS=peer0.it.inventory.com:10051 -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.it.inventory.com/tls/ca.crt cli peer lifecycle chaincode install assetcc.tar.gz

    # 6. Retrieve package ID
    Write-Host "Retrieving Package ID..."
    $packageId = (docker exec cli peer lifecycle chaincode calculatepackageid assetcc.tar.gz).Trim()
    Write-Host "Chaincode Package ID: $packageId"

    # 7. Approve for Org
    Write-Host "Approving for InventoryOrg..."
    docker exec cli peer lifecycle chaincode approveformyorg `
      -o orderer.inventory.com:7050 `
      --channelID inventorychannel `
      --name assetcc `
      --version 1.0 `
      --package-id "$packageId" `
      --sequence 1 `
      --tls `
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/inventory.com/orderers/orderer.inventory.com/tls/ca.crt

    # 8. Commit chaincode
    Write-Host "Committing chaincode..."
    docker exec cli peer lifecycle chaincode commit `
      -o orderer.inventory.com:7050 `
      --channelID inventorychannel `
      --name assetcc `
      --version 1.0 `
      --sequence 1 `
      --tls `
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/inventory.com/orderers/orderer.inventory.com/tls/ca.crt `
      --peerAddresses peer0.lab.inventory.com:7051 `
      --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.lab.inventory.com/tls/ca.crt `
      --peerAddresses peer0.admin.inventory.com:8051 `
      --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.admin.inventory.com/tls/ca.crt

    Write-Host "========================================================="
    Write-Host "assetcc Go Chaincode deployed successfully!"
    Write-Host "========================================================="
} finally {
    Pop-Location
}
