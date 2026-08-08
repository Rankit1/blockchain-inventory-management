# Hybrid Cloud–Blockchain Inventory Management — Go Prototype & Hyperledger Fabric

This repository is a hybrid inventory management solution implemented in Go, supporting both:

- a self-contained local blockchain simulation mode with persistent `bbolt` storage, and
- a real Hyperledger Fabric integration mode using the Fabric Gateway SDK.

The system tracks multi-department assets, manages stock transactions, computes priority tiers, records audit lifecycle events, and supports runtime GenAI automation agents for classification, predictive audits, vision checks, and document intelligence.

---

## What the Project Does Today

This project now implements:

- Asset issue/consume/transfer workflows with blockchain-style transaction recording.
- Persistent simulated ledger + world state using `bbolt` in `data/inventory.db`.
- Fabric integration mode that connects to a real Fabric peer and submits transactions through `fabricclient`.
- Asset priority scoring and tier classification based on business criticality, replacement cost, lead time, safety compliance, and redundancy availability.
- Manual priority override with justification.
- Audit scheduling and audit result recording flows.
- Asset retirement tracking.
- Runtime GenAI automation for:
  - automatic priority classification (`GenAIService`),
  - predictive audit scheduling for P1 assets (`PredictiveAgent`),
  - simulated vision audits (`VisionAgent`),
  - simulated document extraction and warranty enrichment (`DocumentAgent`).
- Runtime admin control over agent enable/disable and provider model selection via `/api/admin/agents/control`.
- A web API server on port `8080` and an optional interactive CLI client in `cmd/client`.

---

## How to Run

### Simulation Mode

This is the default mode if `RUN_MODE` is unset or if Fabric connection fails.

```bash
# Run the API server in local simulation mode
go run .
```

The local simulation stores data in `./data/inventory.db`.

### Fabric Mode

Set the runtime mode to Fabric and provide peer connection details:

```bash
RUN_MODE=fabric FABRIC_CRYPTO_DIR=./network/crypto-config FABRIC_PEER_ADDRESS=localhost:7051 FABRIC_PEER_NAME=peer0.lab.inventory.com go run .
```

If Fabric is not available, the code automatically falls back to simulation mode.

### Model and Agent Configuration

- `AGENT_VISION_ENABLED=true|false`
- `AGENT_DOCUMENT_ENABLED=true|false`
- `OCR_PROVIDER=tesseract|azure|dummy`
- `LLM_PROVIDER=openai|mistral|gemini|dummy`
- `OPENAI_API_KEY`, `MISTRAL_API_KEY`, `GEMINI_API_KEY`
- `AZURE_OCR_ENDPOINT`, `AZURE_OCR_KEY`

---

## API Endpoints

### Health
- `GET /healthz`
  - Returns service health.

### Asset Management
- `POST /api/assets/issue`
  - Issue a new inventory asset.
- `POST /api/assets/consume`
  - Consume stock from an asset.
- `POST /api/assets/transfer`
  - Transfer stock across departments.
- `GET /api/assets/{id}`
  - Read a single asset.
- `GET /api/assets/{id}/history`
  - Fetch transaction history for an asset.
- `GET /api/assets`
  - Query all assets, optional `?dept=` filter.

### GenAI-augmented Asset Operations
- `POST /api/assets/classify`
  - Classify asset priority using five criteria.
  - RBAC: `AI_OPS`, `ASSET_AUDITOR`, or `SYSTEM_ADMIN`.
- `POST /api/assets/update-priority`
  - Manually override asset priority tier with a reason.
  - RBAC: `ASSET_AUDITOR`, `IT_ADMIN`, or `SYSTEM_ADMIN`.
- `POST /api/assets/schedule-audit`
  - Schedule an asset audit.
  - RBAC: `ASSET_AUDITOR` or `SYSTEM_ADMIN`.
- `POST /api/assets/record-audit`
  - Record audit results and update audit metadata.
  - RBAC: `ASSET_AUDITOR` or `SYSTEM_ADMIN`.
- `POST /api/assets/retire`
  - Retire an asset permanently.
  - RBAC: `IT_ADMIN` or `SYSTEM_ADMIN`.

### Replenishment
- `POST /api/replenish/request`
  - Request replenishment for a low-stock asset.

### Reports
- `GET /api/reports/utilization`
  - Returns asset utilization metrics.
  - RBAC: `STORE_MANAGER`, `IT_ADMIN`, or `SYSTEM_ADMIN`.
- `GET /api/reports/compliance`
  - Returns compliance audit summary and overdue assets.
  - RBAC: `IT_ADMIN` or `SYSTEM_ADMIN`.

### Assistant and Admin
- `POST /api/assistant/query`
  - Simple stub assistant for guidance on asset, audit, and priority operations.
- `POST /api/admin/agents/control`
  - Runtime toggle for `genai`, `predictive`, `vision`, and `document` agents.
  - Also accepts model provider updates.
  - RBAC: `IT_ADMIN` or `SYSTEM_ADMIN`.

### Ledger Inspection
- `GET /api/ledger/blocks`
  - Returns stored blockchain blocks in simulation mode.
- `GET /api/ledger/verify`
  - Verifies ledger chain integrity.

---

## File Overview

### Root
- `main.go`
  - Application entrypoint.
  - Chooses Fabric or simulation mode, starts GenAI agents, registers runtime manager, and launches the HTTP server.
- `go.mod`, `go.sum`
  - Go module and dependency management.
- `README.md`
  - Project documentation.
- `TECHNICAL_ADDENDUM.md`
  - Additional design notes and system addenda.

### `cmd/client`
- `main.go`
  - Interactive CLI client that calls the REST API endpoints for asset operations, classification, audits, reports, and ledger verification.

### `internal/api`
- `router.go`
  - Sets up the Chi router, middleware, and all HTTP routes.
- `handlers.go`
  - Implements each REST API endpoint and request/response payloads.
- `client.go`
  - Blockchain client abstraction used by the API handlers.
  - Provides both `FabricAdapter` and `SimulationAdapter` implementations.
- `admin.go`
  - Admin endpoint that toggles runtime GenAI agents and updates model providers.
- `rbac.go`
  - Role-based access control middleware enforcing `X-User-Role` permissions.

### `internal/genai`
- `adapter.go`
  - Defines `AutomationDriver` for both simulated and Fabric-backed automation.
  - `SimulationDriver` and `FabricDriver` bridge agent actions to underlying transaction submission.
- `genai.go`
  - `GenAIService` scans assets and auto-classifies missing or stale priority tiers.
- `predictive.go`
  - `PredictiveAgent` schedules audits for high-priority assets with missing audit history.
- `vision.go`
  - `VisionAgent` simulates visual asset auditing and records audit results.
- `document.go`
  - `DocumentAgent` simulates document intelligence, attaches warranty/A MC notes, and triggers reclassification.
- `manager.go`
  - Global runtime manager for enabling/disabling agents and updating active model providers.
- `models_env.go`
  - Env-driven model provider selection for OCR and LLM adapters.
- `models.go`
  - GenAI model interface definitions.
- `models_stub.go`
  - Dummy OCR/LLM implementations used when real providers are not configured.
- `llm_openai.go`, `llm_mistral.go`, `llm_gemini.go`
  - LLM adapter stubs for provider support.
- `ocr_azure.go`, `ocr_tesseract.go`
  - OCR adapter stubs for Azure and Tesseract.

### `internal/network`
- `network.go`
  - Builds the in-memory blockchain network, the simulated orderer, and peer endorsements.
  - Implements `ProposeAndCommit` to simulate endorsement, ordering, and state updates.

### `internal/ledger`
- `ledger.go`
  - `Ledger` persistence over `bbolt`, block append, chain retrieval, integrity verification, and history scan.

### `internal/worldstate`
- `store.go`
  - `worldstate.Store` persistence for asset records and list/query operations.

### `internal/fabricclient`
- `client.go`
  - Fabric Gateway connection helper and REST-style wrappers to submit/evaluate Fabric chaincode transactions.
  - Supports issue, consume, transfer, read, query, classification, audit, and retirement operations.

### `chaincode/assetcc`
- `asset_contract.go`
  - Fabric smart contract for asset lifecycle management and priority classification.
  - Implements issuing, consuming, transferring, auditing, and retirement.
- `asset_contract_test.go`
  - Chaincode unit tests.
- `META-INF/statedb/couchdb/indexes/`
  - CouchDB index definitions for efficient Rich Query asset lookups.

### `network`
- `docker-compose-inventory-net.yaml`
  - Fabric network composition for orderer, peers, CouchDB, and certificate authorities.
- `crypto-config.yaml`, `configtx.yaml`
  - Fabric cryptogen and channel configuration.
- `scripts/`
  - `network-up.sh`, `network-up.ps1` — start the Fabric network.
  - `deploy-cc.sh`, `deploy-cc.ps1` — package and deploy chaincode.
  - `network-down.sh`, `network-down.ps1` — tear down the Fabric network.

---

## Current Behavior Summary

- Simulation mode is fully functional and persists state locally.
- Fabric mode is wired to the Gateway SDK and can operate against a real Fabric peer.
- GenAI automation is available both in simulation and Fabric mode, with runtime toggles and model provider selection.
- The CLI client exercises the same HTTP API as the web server.
- RBAC is enforced by `X-User-Role` headers for protected endpoints.

## Notes

- The assistant queries are currently a stub and return guidance rather than a production conversational AI answer.
- Real OCR/LLM provider support depends on valid provider credentials and environment configuration.
- The simulated agents use local heuristics and dummy model adapters for deterministic behavior.

