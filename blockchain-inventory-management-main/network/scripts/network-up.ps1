$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location "$scriptDir/.."

try {
    Write-Host "========================================================="
    Write-Host "Starting Hyperledger Fabric Single-Org 4-Peer Network"
    Write-Host "========================================================="

    # 1. Clean previous run state
    if (Test-Path "$scriptDir/network-down.ps1") {
        & "$scriptDir/network-down.ps1"
    }

    # 2. Create directories
    New-Item -ItemType Directory -Force -Path "channel-artifacts" | Out-Null
    New-Item -ItemType Directory -Force -Path "system-genesis-block" | Out-Null

    $pwdPath = (Get-Location).Path

    # 3. Generate crypto material
    Write-Host "Generating cryptographic certificates via Docker..."
    docker run --rm -v "${pwdPath}:/network" -w /network hyperledger/fabric-tools:2.5.15 cryptogen generate --config=./crypto-config.yaml --output="crypto-config"

    # 4. Generate system genesis block and channel configuration transactions
    Write-Host "Generating channel genesis and transactions via Docker..."
    docker run --rm -v "${pwdPath}:/network" -w /network -e FABRIC_CFG_PATH=/network hyperledger/fabric-tools:2.5.15 configtxgen -profile InventoryOrdererGenesis -outputBlock ./system-genesis-block/genesis.block -channelID system-channel
    docker run --rm -v "${pwdPath}:/network" -w /network -e FABRIC_CFG_PATH=/network hyperledger/fabric-tools:2.5.15 configtxgen -profile InventoryOrgChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID inventorychannel

    # 5. Determine Docker Compose command
    if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
        $dockerCompose = "docker-compose"
    } else {
        $dockerCompose = "docker compose"
    }

    Write-Host "Starting containers..."
    if ($dockerCompose -eq "docker-compose") {
        docker-compose -f docker-compose-inventory-net.yaml up -d
    } else {
        docker compose -f docker-compose-inventory-net.yaml up -d
    }

    Write-Host "Waiting for containers to stabilize (10s)..."
    Start-Sleep -Seconds 10

    # 6. Create channel
    Write-Host "Creating inventorychannel..."
    docker exec cli peer channel create `
      -o orderer.inventory.com:7050 `
      -c inventorychannel `
      -f ./network/channel-artifacts/channel.tx `
      --outputBlock ./inventorychannel.block `
      --tls `
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/inventory.com/orderers/orderer.inventory.com/tls/ca.crt

    # 7. Join peers
    Write-Host "Joining peer0.lab to inventorychannel..."
    docker exec cli peer channel join -b ./inventorychannel.block

    Write-Host "Joining peer0.admin to inventorychannel..."
    docker exec -e CORE_PEER_ADDRESS=peer0.admin.inventory.com:8051 -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.admin.inventory.com/tls/ca.crt cli peer channel join -b ./inventorychannel.block

    Write-Host "Joining peer0.store to inventorychannel..."
    docker exec -e CORE_PEER_ADDRESS=peer0.store.inventory.com:9051 -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.store.inventory.com/tls/ca.crt cli peer channel join -b ./inventorychannel.block

    Write-Host "Joining peer0.it to inventorychannel..."
    docker exec -e CORE_PEER_ADDRESS=peer0.it.inventory.com:10051 -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/inventory.com/peers/peer0.it.inventory.com/tls/ca.crt cli peer channel join -b ./inventorychannel.block

    Write-Host "========================================================="
    Write-Host "Blockchain Network started & channel joined successfully!"
    Write-Host "========================================================="
} finally {
    Pop-Location
}
