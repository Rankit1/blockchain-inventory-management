$ErrorActionPreference = "Stop"

Write-Host "========================================================="
Write-Host "Running End-to-End Fabric Blockchain + GenAI Test"
Write-Host "========================================================="

# 1. Health Check
Write-Host "1. Testing Healthz..."
$health = Invoke-RestMethod -Uri "http://localhost:8080/healthz" -Method Get
Write-Host "Health status: $($health.status), mode: $($health.mode)"

# 2. Issue Asset
Write-Host "`n2. Issuing Asset (Category: server, Qty: 10, Threshold: 3)..."
$issuePayload = @{
    deptID    = "Lab"
    name      = "E2E Quantum Server X9"
    category  = "server"
    qty       = 10
    threshold = 3
} | ConvertTo-Json

$issueRes = Invoke-RestMethod -Uri "http://localhost:8080/api/assets/issue" -Method Post -ContentType "application/json" -Body $issuePayload
$assetID = $issueRes.assetID
Write-Host "Asset Issued Successfully! AssetID: $assetID, TxID: $($issueRes.txID)"

# 3. Wait for GenAI automation agents to process the asset
Write-Host "`n3. Waiting 5s for GenAI Agents (Classification, Predictive, Document, Vision)..."
Start-Sleep -Seconds 5

# 4. Fetch Asset details
Write-Host "`n4. Reading Asset State from Fabric..."
$asset = Invoke-RestMethod -Uri "http://localhost:8080/api/assets/$assetID" -Method Get
Write-Host "Asset Name: $($asset.name)"
Write-Host "Category: $($asset.category)"
Write-Host "Quantity: $($asset.qty)"
Write-Host "Priority Tier: $($asset.priorityTier)"
Write-Host "Criticality Score: $($asset.criticalityScore)"
Write-Host "Lifecycle State: $($asset.lifecycleState)"
Write-Host "Last Audit Date: $($asset.lastAuditDate)"
Write-Host "Warranty Expiry: $($asset.warrantyExpiry)"
Write-Host "AMC Expiry: $($asset.amcExpiry)"

# 5. Consume Stock to trigger replenishment
Write-Host "`n5. Consuming Stock (Qty: 8)..."
$consumePayload = @{
    deptID  = "Lab"
    assetID = $assetID
    qty     = 8
    purpose = "E2E lab workload provisioning"
} | ConvertTo-Json

$consumeRes = Invoke-RestMethod -Uri "http://localhost:8080/api/assets/consume" -Method Post -ContentType "application/json" -Body $consumePayload
Write-Host "Consume TxID: $($consumeRes.txID), New Qty: $($consumeRes.newQty), Replenish Triggered: $($consumeRes.replenishTriggered)"

# 6. Transfer 1 unit to IT department
Write-Host "`n6. Transferring 1 Unit to IT Department..."
$transferPayload = @{
    fromDept = "Lab"
    toDept   = "IT"
    assetID  = $assetID
    qty      = 1
} | ConvertTo-Json

$transferRes = Invoke-RestMethod -Uri "http://localhost:8080/api/assets/transfer" -Method Post -ContentType "application/json" -Body $transferPayload
Write-Host "Transfer TxID: $($transferRes.txID), Source Qty: $($transferRes.sourceQty), Dest Qty: $($transferRes.destQty)"

# 7. Fetch Transaction History
Write-Host "`n7. Fetching Provenance / Transaction History from Fabric Ledger..."
$history = Invoke-RestMethod -Uri "http://localhost:8080/api/assets/$assetID/history" -Method Get
Write-Host "Total Transactions Recorded in History: $($history.Count)"
foreach ($entry in $history) {
    Write-Host "  - TxID: $($entry.txId) | IsDelete: $($entry.isDelete) | Timestamp: $($entry.timestamp) | Qty: $($entry.value.qty)"
}

# 8. Query Compliance Report (RBAC: IT_ADMIN)
Write-Host "`n8. Querying Compliance Report (Role: IT_ADMIN)..."
$headers = @{ "X-User-Role" = "IT_ADMIN" }
$compliance = Invoke-RestMethod -Uri "http://localhost:8080/api/reports/compliance" -Method Get -Headers $headers
Write-Host "Compliant Assets: $($compliance.compliantAssets.Count), Non-Compliant Assets: $($compliance.nonCompliantAssets.Count)"

# 9. Query GenAI Assistant (RBAC: ASSET_AUDITOR)
Write-Host "`n9. Querying GenAI Assistant (Role: ASSET_AUDITOR)..."
$assistantHeaders = @{ "X-User-Role" = "ASSET_AUDITOR"; "Content-Type" = "application/json" }
$assistantPayload = @{ query = "Provide an operational summary of the newly deployed server assets." } | ConvertTo-Json
$assistantRes = Invoke-RestMethod -Uri "http://localhost:8080/api/assistant/query" -Method Post -Headers $assistantHeaders -Body $assistantPayload
Write-Host "Assistant Response:"
Write-Host $assistantRes.response

Write-Host "`n========================================================="
Write-Host "End-to-End Test Completed Successfully!"
Write-Host "========================================================="
