# Hybrid Cloud–Blockchain Inventory Management — Go Prototype & Hyperledger Fabric

This repository is an inventory management solution implemented in Go on top of a real Hyperledger Fabric v2.5 network, connected via the Fabric Gateway SDK.

The system tracks multi-department assets, manages stock transactions, computes priority tiers, records audit lifecycle events, and supports runtime GenAI automation agents for classification, predictive audits, vision checks, and document intelligence.

---

## What the Project Does Today

This project implements:

- Asset issue/consume/transfer workflows submitted as Fabric chaincode transactions.
- Asset priority scoring and tier classification based on business criticality, replacement cost, lead time, safety compliance, and redundancy availability — scored by an LLM (OpenAI/Gemini/Mistral) with a deterministic heuristic fallback when no provider is configured.
- Manual priority override with justification.
- Audit scheduling and audit result recording flows.
- Asset retirement tracking.
- Runtime GenAI automation for:
  - automatic priority classification (`GenAIService`), backed by a real LLM call,
  - predictive audit scheduling for P1 assets (`PredictiveAgent`),
  - periodic audit-heartbeat checks (`VisionAgent`) — enriched via OCR when a provider is configured; a real photo-upload pipeline is not wired up yet,
  - document metadata extraction (`DocumentAgent`) via the configured OCR/LLM providers; a real document-upload pipeline is not wired up yet.
- Runtime admin control over agent enable/disable and provider model selection via `/api/admin/agents/control`.
- A web API server on port `8080` and an optional interactive CLI client in `cmd/client`.

---

## How to Run

This application requires a running Hyperledger Fabric network (see `network/`) with the `assetcc` chaincode deployed to the `inventorychannel` channel.

```bash
FABRIC_CRYPTO_DIR=./network/crypto-config FABRIC_PEER_ADDRESS=localhost:7051 FABRIC_PEER_NAME=peer0.lab.inventory.com go run .
```

If `FABRIC_CRYPTO_DIR`/`FABRIC_PEER_ADDRESS`/`FABRIC_PEER_NAME` are unset, they default to `./network/crypto-config`, `localhost:7051`, and `peer0.lab.inventory.com` respectively. The server exits at startup if it cannot connect to the peer.

### Model and Agent Configuration

- `AGENT_VISION_ENABLED=true|false`
- `AGENT_DOCUMENT_ENABLED=true|false`
- `OCR_PROVIDER=tesseract|azure|dummy`
- `LLM_PROVIDER=openai|mistral|gemini|dummy`
- `OPENAI_API_KEY`, `MISTRAL_API_KEY`, `GEMINI_API_KEY`
- `AZURE_OCR_ENDPOINT`, `AZURE_OCR_KEY`

Without a real `LLM_PROVIDER`/API key configured, priority classification falls back to a deterministic category-based heuristic instead of an LLM call.

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
  - Returns raw ledger blocks (not yet wired to Fabric's qscc query; returns empty).
- `GET /api/ledger/verify`
  - Verifies ledger chain integrity (Fabric's consensus already guarantees this; returns `true`).

---

## File Overview

### Root
- `main.go`
  - Application entrypoint. Connects to the Fabric Gateway, starts GenAI agents, registers the runtime manager, and launches the HTTP server.
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
  - `FabricAdapter`: the `BlockchainClient` implementation that wraps the Fabric Gateway SDK.
- `admin.go`
  - Admin endpoint that toggles runtime GenAI agents and updates model providers.
- `rbac.go`
  - Role-based access control middleware enforcing `X-User-Role` permissions.

### `internal/genai`
- `adapter.go`
  - Defines `AutomationDriver` and `FabricDriver`, bridging agent actions to Fabric transaction submission.
- `genai.go`
  - `GenAIService` scans assets and auto-classifies missing or stale priority tiers using an LLM call (with heuristic fallback).
- `predictive.go`
  - `PredictiveAgent` schedules audits for high-priority assets with missing audit history.
- `vision.go`
  - `VisionAgent` runs a periodic audit heartbeat and records audit results.
- `document.go`
  - `DocumentAgent` extracts warranty/AMC notes via OCR/LLM and triggers reclassification.
- `manager.go`
  - Global runtime manager for enabling/disabling agents and updating active model providers.
- `priority.go`
  - `PriorityScores` type, asset lifecycle constants, and the deterministic heuristic fallback scorer.
- `llm_common.go`
  - Shared LLM prompt and JSON-response parsing used by all providers.
- `models_env.go`
  - Env-driven model provider selection for OCR and LLM adapters.
- `models.go`
  - GenAI model interface definitions.
- `models_stub.go`
  - Dummy OCR/LLM implementations used when real providers are not configured.
- `llm_openai.go`, `llm_mistral.go`, `llm_gemini.go`
  - Real LLM adapters (OpenAI Chat Completions, Mistral Chat Completions, Google Generative Language API) for document summarization and priority scoring.
- `ocr_azure.go`, `ocr_tesseract.go`
  - OCR adapters for Azure Computer Vision Read API and local Tesseract.

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

- Fabric mode is the only runtime mode; the server requires a live Fabric peer to connect at startup.
- Priority classification calls a real LLM provider (OpenAI/Mistral/Gemini) when configured, falling back to a deterministic heuristic otherwise.
- The CLI client exercises the same HTTP API as the web server.
- RBAC is enforced by `X-User-Role` headers for protected endpoints.

## Notes

- The assistant queries are currently a stub and return guidance rather than a production conversational AI answer.
- Real OCR/LLM provider support depends on valid provider credentials and environment configuration.
- The Vision and Document agents call the configured OCR provider, but no photo/document upload endpoint exists yet, so they run without real image/document bytes until that pipeline is built.
