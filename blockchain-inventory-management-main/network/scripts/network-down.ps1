$ErrorActionPreference = "Continue"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location "$scriptDir/.."

try {
    if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
        $dockerCompose = "docker-compose"
    } else {
        $dockerCompose = "docker compose"
    }

    Write-Host "Stopping Fabric network..."
    if ($dockerCompose -eq "docker-compose") {
        docker-compose -f docker-compose-inventory-net.yaml down -v
    } else {
        docker compose -f docker-compose-inventory-net.yaml down -v
    }

    Write-Host "Cleaning generated config and crypto files..."
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue crypto-config
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue channel-artifacts
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue system-genesis-block
    Remove-Item -Force -ErrorAction SilentlyContinue inventorychannel.block
    Remove-Item -Force -ErrorAction SilentlyContinue assetcc.tar.gz

    Write-Host "Network stopped & cleaned."
} finally {
    Pop-Location
}
