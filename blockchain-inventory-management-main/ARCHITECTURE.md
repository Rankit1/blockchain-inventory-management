# Hybrid Cloud–Blockchain Inventory Management System
## Architectural Design & Engineering Onboarding Guide

---

## 1. Executive Summary & Core Concept

This project is a **Hybrid Cloud-Blockchain Inventory Management Platform** implemented in **Go** with a **React** web dashboard and an interactive **Go CLI**. It addresses enterprise inventory lifecycle governance, supply chain provenance, and proactive maintenance by uniting two core architectural pillars:

1. **Immutable Distributed Ledger (Blockchain)**:
   * Provides a tamper-proof audit trail for all asset lifecycle events (issuance, stock consumption, inter-department transfers, audit results, and decommissioning).
   * Runs on a real **Hyperledger Fabric v2.5** network, reached via the **Fabric Gateway SDK** (`internal/fabricclient`).

2. **Autonomous GenAI Automation Layer (Agentic AI)**:
   * Eliminates passive spreadsheets by deploying autonomous background worker agents.
   * Periodically scans ledger state to calculate 5-factor priority tiers (`P1`, `P2`, `P3`) via a real LLM call (OpenAI/Gemini/Mistral, with a deterministic heuristic fallback), auto-schedules predictive maintenance audits, and parses warranty/service agreements via OCR and LLM adapters.

---

## 2. High-Level System Architecture

```
                                  USER INTERFACES
      ┌─────────────────────────────────┐   ┌─────────────────────────────────┐
      │   React Web Dashboard (Vite)    │   │  Interactive Terminal CLI Client │
      │          (Port 5173)            │   │          (cmd/client)           │
      └────────────────┬────────────────┘   └────────────────┬────────────────┘
                       │                                     │
                       └──────────────────┬──────────────────┘
                                          │ HTTP / REST (JSON)
                                          ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                 API SERVER (Port 8080)                                 │
│                                                                                        │
│   • Chi HTTP Router (internal/api/router.go)                                           │
│   • Role-Based Access Control Middleware (internal/api/rbac.go)                        │
│   • Handlers & Controllers (internal/api/handlers.go)                                  │
│   • Runtime Admin Agent Controller (internal/api/admin.go)                             │
└─────────────────────────────────────────┬──────────────────────────────────────────────┘
                                          │
                                          ▼
                        ┌───────────────────────────────┐
                        │    FabricAdapter (client.go)  │
                        └───────────────┬───────────────┘
                                        │
                                        ▼
                        ┌────────────────────────────────────┐
                        │     ENTERPRISE FABRIC DEPLOYMENT   │
                        │                                    │
                        │ • Fabric Gateway SDK (gRPC)        │
                        │ • Smart Contract: asset_contract.go│
                        │ • State DB: CouchDB (Rich Queries) │
                        │ • Topology: Orderer, Peers, CAs    │
                        └─────────────────▲──────────────────┘
                                          │
                                          │ AutomationDriver Interface
                                          ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                             GENAI AUTOMATION AGENT LAYER                               │
│                                                                                        │
│   • Priority Classification Agent (GenAIService)  -> LLM-scored P1/P2/P3 tiers         │
│   • Predictive Maintenance Agent (PredictiveAgent)-> Auto-schedules audits for P1 items│
│   • Visual Inspection Agent (VisionAgent)         -> Periodic audit heartbeat + OCR    │
│   • Document Intelligence Agent (DocumentAgent)   -> Extracts warranty / AMC metadata  │
│   • Model Provider Pipeline                       -> Dummy <-> OpenAI, Gemini, Mistral, Azure │
│   • Runtime Agent Manager (GlobalManager)         -> Dynamic start/stop & reconfig     │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Repository Directory Structure

```
blockchain-inventory-management-main/
├── main.go                               # Application entrypoint; connects to Fabric, starts agents, and HTTP server
├── go.mod / go.sum                       # Go module dependencies
├── README.md                             # Project overview & basic run instructions
├── TECHNICAL_ADDENDUM.md                 # Extended technical specification
├── ARCHITECTURE.md                       # Comprehensive architecture and onboarding document (this file)
│
├── chaincode/                            # Hyperledger Fabric Smart Contracts (Chaincode)
│   └── assetcc/
│       ├── asset_contract.go             # Fabric contract implementing asset CRUD, classification, and audits
│       ├── asset_contract_test.go        # Unit tests for asset contract functions
│       ├── classification_contract_test.go# Unit tests for chaincode scoring logic
│       └── META-INF/statedb/couchdb/     # CouchDB JSON index definitions for fast lookups
│
├── internal/                             # Core Go Backend Logic
│   ├── api/                              # HTTP REST Server Layer
│   │   ├── router.go                     # Route declarations and Chi middleware setup
│   │   ├── handlers.go                   # HTTP handler functions for asset operations and reports
│   │   ├── client.go                     # BlockchainClient interface; FabricAdapter implementation
│   │   ├── rbac.go                       # Role-Based Access Control definition and validation
│   │   └── admin.go                      # Endpoints for runtime agent management and model switching
│   │
│   ├── genai/                            # Autonomous Agent & AI Subsystem
│   │   ├── adapter.go                    # AutomationDriver interface + FabricDriver (bridges agents to Fabric)
│   │   ├── genai.go                      # GenAIService background scanner for priority classification
│   │   ├── predictive.go                 # PredictiveAgent for audit scheduling on critical assets
│   │   ├── vision.go                     # VisionAgent audit heartbeat + OCR enrichment
│   │   ├── document.go                   # DocumentAgent for document OCR & warranty extraction
│   │   ├── manager.go                    # GlobalManager for starting/stopping agents at runtime
│   │   ├── priority.go                   # PriorityScores type, lifecycle constants, heuristic fallback scorer
│   │   ├── llm_common.go                 # Shared LLM prompt + JSON response parsing
│   │   ├── models.go                     # ModelProvider, OCRModel, and LLMModel interfaces
│   │   ├── models_env.go                 # Provider factory reading environment variables
│   │   ├── models_stub.go                # Deterministic dummy models for local testing
│   │   ├── llm_gemini.go                 # Google Generative Language API client
│   │   ├── llm_openai.go                 # OpenAI Chat Completions API client
│   │   ├── llm_mistral.go                # Mistral AI Chat Completions API client
│   │   ├── ocr_azure.go                  # Azure Cognitive Services OCR client
│   │   └── ocr_tesseract.go              # Local Tesseract OCR client
│   │
│   └── fabricclient/                     # Real Fabric Network Integration
│       └── client.go                     # Fabric Gateway SDK gRPC wrapper, certificate loader, tx submission
│
├── frontend/                             # React Single Page Application (Vite)
│   ├── src/
│   │   ├── App.jsx                       # Main shell, role selector, and route definition
│   │   ├── main.jsx                      # React DOM mounting
│   │   ├── styles.css                    # Modern enterprise styling and design tokens
│   │   ├── api/client.js                 # Axios HTTP client configured with RBAC headers
│   │   └── pages/                        # View components (Dashboard, Assets, Audits, Ledger, Admin, etc.)
│   ├── package.json
│   └── vite.config.js
│
├── cmd/
│   └── client/
│       └── main.go                       # Interactive command-line terminal client
│
└── network/                              # Production Fabric Infrastructure Assets
    ├── docker-compose-inventory-net.yaml # Docker Compose topology for Fabric nodes & CouchDB
    ├── configtx.yaml                     # Channel definition & endorsement policy configuration
    ├── crypto-config.yaml                # Cryptographic material generator configuration
    └── scripts/                          # network-up, deploy-cc, network-down bash & PowerShell scripts
```

---

## 4. Subsystem Deep Dives

### 4.1 Storage & Consensus Subsystem

The application decouples the HTTP handlers from ledger access behind the `BlockchainClient` interface in `internal/api/client.go`, implemented by `FabricAdapter`.

```go
type BlockchainClient interface {
    IssueAsset(deptID, name, category string, qty, threshold int) (txID, assetID string, timestamp time.Time, err error)
    ConsumeStock(deptID, assetID string, qty int, purpose string) (txID string, newQty int, replenishTriggered bool, err error)
    TransferAsset(fromDept, toDept, assetID string, qty int) (txID string, fromQty, toQty int, err error)
    ReadAsset(assetID string) (*ClientAsset, error)
    QueryAssetsByDept(deptID string) ([]*ClientAsset, error)
    GetAssetHistory(assetID string) ([]ClientHistoryEntry, error)
    ClassifyPriority(assetID string, scores genai.PriorityScores) (txID, tier string, score float64, err error)
    UpdatePriorityTier(assetID, tier, reason string) (txID string, err error)
    ScheduleAudit(assetID, auditDate, scope string) (txID string, err error)
    RecordAuditResult(assetID, auditDate, result, notes string) (txID string, err error)
    RetireAsset(assetID, reason string) (txID string, err error)
    GetLedgerBlocks() ([]interface{}, error)
    VerifyLedger() (bool, *int, error)
}
```

#### Enterprise Hyperledger Fabric Engine
* **Connection**: Managed by `internal/fabricclient/client.go` using the official `fabric-gateway-go` SDK.
* **Smart Contract**: `chaincode/assetcc/asset_contract.go` deployed to a channel (default: `inventorychannel`).
* **Rich Queries**: Leverages CouchDB indexes defined in `chaincode/assetcc/META-INF/` for sub-millisecond filtering by department, lifecycle state, and priority tier.
* **Startup requirement**: the server connects to the Fabric peer at process start and exits if it cannot — there is no local fallback mode.

---

### 4.2 Asset Domain Model & Priority Classification

#### Asset Lifecycle State Machine
$$\text{DRAFT} \longrightarrow \text{ACTIVE} \underset{\text{Maintenance Flow}}{\rightleftarrows} \text{UNDER\_MAINTENANCE} \longrightarrow \text{RETIRED}$$

* **`ACTIVE`**: Available for standard consumption, transfer, and operations.
* **`UNDER_MAINTENANCE`**: Temporarily locked for repairs or audit remediation.
* **`RETIRED`**: Read-only archived asset; cannot be issued, transferred, or consumed.

#### Multi-Factor Priority Scoring Engine
Asset criticality is computed using five normalized parameters ($1 \le x_i \le 5$). The five ratings are produced by an LLM call (`GenAIService.classify` in `internal/genai/genai.go`, falling back to a deterministic heuristic in `internal/genai/priority.go` when no LLM is configured); the weighted sum and tier derivation happen on-chain in `chaincode/assetcc/asset_contract.go`'s `ClassifyPriority`:

$$\text{Criticality Score} = \sum_{i=1}^{5} w_i \cdot s_i$$

| Parameter ($s_i$) | Weight ($w_i$) | Business Meaning |
| :--- | :---: | :--- |
| **Business Criticality** | **0.30** | Impact on core operations if unavailable |
| **Replacement Cost** | **0.20** | Financial expenditure required to replace |
| **Replacement Lead Time** | **0.20** | Procurement delay in weeks/months |
| **Safety & Compliance Impact** | **0.15** | Regulatory fines, safety hazards, or SLA violations |
| **Redundancy Availability** | **0.15** | Availability of backup/standby units (inverted scale) |

**Classification Tiers**:
* **`P1` (Critical)**: $\text{Score} \ge 4.0$ $\rightarrow$ Requires mandatory monthly audits & predictive scheduling.
* **`P2` (Moderate)**: $2.5 \le \text{Score} < 4.0$ $\rightarrow$ Quarterly inspection cycle.
* **`P3` (Low)**: $\text{Score} < 2.5$ $\rightarrow$ Standard annual audit cycle.

---

### 4.3 Autonomous GenAI Agent Ecosystem

The agent subsystem runs non-blocking background workers (`goroutines`) that communicate with the ledger through the [`AutomationDriver`](file:///c:/Users/branybuck/code/blockchain%20management/blockchain-inventory-management/blockchain-inventory-management-main/internal/genai/adapter.go) interface:

```
                               ┌───────────────────────────┐
                               │  AutomationDriver Bridge  │
                               └─────────────┬─────────────┘
                                             │
             ┌───────────────────────────────┼───────────────────────────────┐
             ▼                               ▼                               ▼
    ┌─────────────────┐            ┌───────────────────┐           ┌───────────────────┐
    │  GenAIService   │            │  PredictiveAgent  │           │   VisionAgent     │
    │ Scans inventory │            │ Identifies P1     │           │ Runs periodic     │
    │ and LLM-scores  │            │ assets needing    │           │ audit heartbeat & │
    │ unrated assets  │            │ audit checkups    │           │ commits audits    │
    └─────────────────┘            └───────────────────┘           └───────────────────┘
                                             │
                                             ▼
                                   ┌───────────────────┐
                                   │   DocumentAgent   │
                                   │ Extracts warranty │
                                   │ & AMC dates via   │
                                   │ OCR/LLM adapters  │
                                   └───────────────────┘
```

#### Agent Details & Behaviors
1. **`GenAIService` ([`genai.go`](file:///c:/Users/branybuck/code/blockchain%20management/blockchain-inventory-management/blockchain-inventory-management-main/internal/genai/genai.go))**:
   * Inspects all active assets on a 1-second interval ticker.
   * Targets assets with empty `PriorityTier` or stale classifications ($>24\text{ hours}$).
   * Calls the configured LLM (`ScorePriority`) to rate the asset on five criteria, falling back to a deterministic heuristic if no LLM is configured or the call fails, then executes `ClassifyPriority` to assign tier and criticality score on-chain.
2. **`PredictiveAgent` ([`predictive.go`](file:///c:/Users/branybuck/code/blockchain%20management/blockchain-inventory-management/blockchain-inventory-management-main/internal/genai/predictive.go))**:
   * Evaluates high-criticality assets (`PriorityTier == "P1"`).
   * If `LastAuditDate` is blank, automatically submits a `ScheduleAudit` transaction set for $+24\text{ hours}$ with reason `"predictive-maintenance"`.
3. **`VisionAgent` ([`vision.go`](file:///c:/Users/branybuck/code/blockchain%20management/blockchain-inventory-management/blockchain-inventory-management-main/internal/genai/vision.go))**:
   * Runs a periodic audit-heartbeat check for unaudited assets and enriches the note via the configured OCR provider.
   * Submits a `RecordAuditResult` transaction with outcome `"PASS"` and metadata note `"vision:verified"`.
   * No photo-upload endpoint exists yet, so it does not run real vision-LLM image matching — that remains future work (Section 7, Track B).
4. **`DocumentAgent` ([`document.go`](file:///c:/Users/branybuck/code/blockchain%20management/blockchain-inventory-management/blockchain-inventory-management-main/internal/genai/document.go))**:
   * Invokes the OCR pipeline on asset documents to parse `WarrantyExpiry` and `AMCExpiry` dates.
   * Records `"DOC_EXTRACTED"` and `"DOC_SUMMARY"` audit entries and triggers priority re-evaluation.
5. **Dynamic Runtime Agent Manager ([`manager.go`](file:///c:/Users/branybuck/code/blockchain%20management/blockchain-inventory-management/blockchain-inventory-management-main/internal/genai/manager.go))**:
   * Supports hot-toggling agents and switching LLM/OCR models via `POST /api/admin/agents/control` without process restarts.

---

### 4.4 Role-Based Access Control (RBAC)

The REST API enforces security using the `X-User-Role` HTTP request header ([`rbac.go`](file:///c:/Users/branybuck/code/blockchain%20management/blockchain-inventory-management/blockchain-inventory-management-main/internal/api/rbac.go)).

| Role Identifier | Description & Permissions |
| :--- | :--- |
| **`SYSTEM_ADMIN`** | Unrestricted access across all operational, reporting, and management endpoints. |
| **`IT_ADMIN`** | Can retire assets, manually override priority tiers, view compliance reports, and control AI agents. |
| **`ASSET_AUDITOR`** | Can trigger classifications, schedule physical audits, and record audit inspection results. |
| **`AI_OPS`** | Can trigger batch GenAI classification scans and monitor AI model performance. |
| **`STORE_MANAGER`** | Can issue new assets, consume inventory, transfer items, and generate utilization reports. |
| **`DEPARTMENT_USER`** | Read-only asset queries and replenishment request submissions. |

---

## 5. End-to-End Execution Walkthrough

```
[1] User Action
    Store Manager issues a new asset via Frontend / CLI
    └─► POST /api/assets/issue
        Payload: { assetId: "AST-900", deptId: "Lab", name: "Centrifuge", category: "Medical", qty: 5, threshold: 2 }
        Header:  X-User-Role: STORE_MANAGER

[2] API Handling & RBAC Check
    internal/api/router.go matches route -> passes RBAC -> calls handlers.IssueAsset()

[3] Transaction Creation & Consensus (Fabric)
    FabricAdapter submits IssueAsset via the Fabric Gateway SDK
    ├─► Endorsing peers execute chaincode/assetcc/asset_contract.go
    ├─► Endorsement policy satisfied (2-of-4), transaction ordered and committed
    ├─► World state written to CouchDB; block appended to the Fabric ledger
    └─► Commit status returned to the caller

[4] Autonomous GenAI Classification Trigger
    GenAIService scanner detects "AST-900" has no PriorityTier
    ├─► Calls the configured LLM's ScorePriority(name, category) — or the heuristic
    │   fallback if no LLM_PROVIDER is configured — for the five criterion ratings
    ├─► Submits ClassifyPriority to chaincode, which computes the weighted
    │   Criticality Score and derives the tier (e.g. 4.05 -> P1)
    └─► Result is written back to the on-chain asset record

[5] Predictive Maintenance Trigger
    PredictiveAgent scanner detects "AST-900" is P1 and has LastAuditDate == ""
    └─► Submits SCHEDULE_AUDIT transaction for tomorrow (Reason: "predictive-maintenance")

[6] Document & Vision Verification Triggers
    ├─► DocumentAgent extracts warranty date (2027-01-01) -> commits "DOC_EXTRACTED"
    └─► VisionAgent verifies presence -> commits "PASS" audit result
```

---

## 6. Developer Quickstart & Operations

### 6.1 Prerequisites
* **Go**: Version 1.20 or higher
* **Node.js**: Version 18 or higher & npm
* Docker, Docker Compose, Git (to run the Fabric network)

### 6.2 Running in Hyperledger Fabric Mode

```bash
# Step 1: Start Fabric Docker topology (Orderer, Peers, CouchDB, CAs)
cd network/scripts
./network-up.sh

# Step 2: Package, install, approve, and commit the smart contract
./deploy-cc.sh

# Step 3: Run the application pointing to the Fabric network
export FABRIC_PEER_ADDRESS="localhost:7051"
export FABRIC_CRYPTO_DIR="./network/crypto-config"
go run .

# Step 4: In a new terminal, launch the Frontend Web Dashboard (runs on :5173)
cd frontend
npm install
npm run dev

# Step 5: (Optional) Run the Interactive Terminal CLI Client
go run ./cmd/client
```

The server connects to the Fabric peer at startup and exits if the connection fails — there is no local fallback mode.

### 6.3 Environment Variables Reference

| Variable | Values | Default | Purpose |
| :--- | :--- | :--- | :--- |
| `FABRIC_PEER_ADDRESS` | `host:port` | `localhost:7051` | Target Fabric peer gRPC endpoint |
| `FABRIC_PEER_NAME` | string | `peer0.lab.inventory.com` | Target Fabric peer identity name |
| `FABRIC_CRYPTO_DIR` | filepath | `./network/crypto-config` | Directory containing MSP identities & TLS certs |
| `AGENT_VISION_ENABLED` | `true` \| `false` | `true` | Enables or disables the VisionAgent on startup |
| `AGENT_DOCUMENT_ENABLED`| `true` \| `false` | `true` | Enables or disables the DocumentAgent on startup |
| `LLM_PROVIDER` | `dummy` \| `openai` \| `gemini` \| `mistral` | `dummy` | Active LLM integration adapter |
| `OCR_PROVIDER` | `dummy` \| `azure` \| `tesseract` | `dummy` | Active OCR processing adapter |
| `OPENAI_API_KEY` | string | `""` | API key for OpenAI LLM services |
| `GEMINI_API_KEY` | string | `""` | API key for Google Gemini LLM services |
| `AZURE_OCR_ENDPOINT` | URL | `""` | Endpoint for Azure Computer Vision OCR |
| `AZURE_OCR_KEY` | string | `""` | Key for Azure Computer Vision OCR |

### 6.4 Running Tests

```bash
# Run all core application and agent unit tests
go test -v ./...

# Run chaincode smart contract unit tests (separate module)
cd chaincode/assetcc && go test -v ./...
```

---

## 7. Future Work & Engineering Roadmap

> **Design Principle: 100% Pure Software Architecture**  
> The entire platform—including all future roadmap tracks—is designed to be **entirely software-based**. It requires no proprietary physical hardware, on-premise hardware sensors, or custom equipment. All data ingestion, telemetry, visual inspections, and document intelligence operate through software APIs, user file uploads, synthetic/simulated telemetry streams, and cloud AI services.

The forward-looking engineering roadmap is organized into five pure-software evolution tracks:

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        PURE SOFTWARE ENGINEERING ROADMAP                               │
│                                                                                        │
│   Phase 3a (Current Foundation)   Phase 3b (Software AI & Ingestion) Phase 3c (NLP/ESG)│
│   ┌───────────────────────────┐   ┌───────────────────────────┐  ┌───────────────────┐ │
│   │ • 5-Factor Scoring Model  │   │ • Software Telemetry Feeds│  │ • Full RAG NLP    │ │
│   │ • Fabric Smart Contracts  │──►│ • Time-Series ML (LSTM)   │─►│   Assistant Engine│ │
│   │ • Real LLM Classification │   │ • Web/Photo Image Uploads │  │ • Dynamic Safety  │ │
│   │ • OCR/LLM Model Adapters  │   │ • Kafka / Pulsar Event Bus│  │   Reorder Models  │ │
│   │ • React + CLI Interfaces  │   │ • Vector DB Off-Chain Sync│  │ • Software ESG /  │ │
│   └───────────────────────────┘   └───────────────────────────┘  │   Compliance Docs │ │
│                                                                  └───────────────────┘ │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

### 7.1 Track A: Predictive Maintenance via Software Telemetry (Phase 3b)
* **Software-Driven Telemetry Ingestion**: Ingest synthetic or system-generated usage metrics (e.g. software runtime hours, utilization rates, error frequency logs, operating cycles) submitted via standard REST/JSON APIs.
* **Time-Series Machine Learning (LSTM / Prophet / AutoRegressive)**: Train purely algorithmic software models on historical asset usage data to compute failure probability curves ($P_{\text{fail}} > \theta$) and auto-trigger preventative maintenance schedules.
* **Algorithmic Anomaly Detection**: Statistical software detection of abnormal usage patterns without relying on physical sensors.

### 7.2 Track B: Web-Based Computer Vision & Document Intelligence (Phase 3b)
* **Web & Mobile Image Upload Pipeline**: Allow auditors to upload standard photos of assets (JPEG/PNG) via the web dashboard. Software vision models (e.g., Vision Transformers, Google Cloud Vision, or Azure Vision APIs) automatically detect asset condition, verify barcode/QR labels from images, and flag visual defects.
* **Cloud OCR & Document Intelligence**: Deep parsing of uploaded PDF invoices, purchase orders, and warranty contracts using cloud OCR models (Azure Document Intelligence, Tesseract, or multimodal LLMs) to extract expiration dates, SLA clauses, and vendor metadata.

### 7.3 Track C: Event-Driven Microservices & Vector Database Sync
* **Message Broker Decoupling (Kafka / Pulsar)**: Stream committed blockchain block events directly to independent software microservices, ensuring horizontal scalability without impacting consensus nodes.
* **Semantic Search & Vector Database Sync**: Mirror committed asset records into a software Vector Database (e.g. pgvector, Pinecone, or Qdrant) along with Elasticsearch for sub-second natural language search and semantic asset queries.

### 7.4 Track D: Conversational AI & Natural Language Operations (Phase 3c)
* **RAG-Powered Inventory Assistant**: An interactive software assistant that retrieves live ledger state to answer natural language questions (e.g., *"Which P1 equipment in the Lab department has had no maintenance in the last 6 months?"*).
* **Text-to-Transaction (NL2Tx) Pipeline**: Allow authorized managers to formulate operations in natural language (e.g., *"Transfer 3 laptops from IT Store to Radiology"*), with software parsing, human-in-the-loop review, and cryptographic signature before on-chain execution.

### 7.5 Track E: Supply Chain Optimization & Automated ESG Compliance (Phase 3c)
* **Dynamic Safety Stock & Reorder Modeling**: Pure software analytics models that continuously recalculate optimal reorder points and safety stock levels based on past consumption spikes and supplier lead times.
* **Automated ESG & Regulatory Compliance Reports**: Software-generated PDF/JSON compliance packages adhering to **E-Waste Management Rules**, **ISO 55000 Asset Management**, and internal governance standards.
* **Zero-Knowledge Proofs (ZKPs) for Audits**: Software cryptographic proofs enabling external third-party auditors to verify compliance without exposing proprietary organizational inventory counts or internal cost figures.


