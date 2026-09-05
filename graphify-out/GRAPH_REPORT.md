# Graph Report - blockchain-inventory-management  (2026-09-05)

## Corpus Check
- Corpus is ~30,503 words - fits in a single context window. You may not need a graph.

## Summary
- 479 nodes · 919 edges · 30 communities (13 shown, 10 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 44 edges (avg confidence: 0.86)
- Token cost: 1,250 input · 850 output

## Community Hubs (Navigation)
- Frontend React State & UI
- GenAI Document & Predictive Agents
- Fabric Gateway gRPC Client
- Backend Blockchain Adapter & Services
- CLI Client Workflow & Operations
- API Request & Response Models
- Smart Contract & Ledger Core
- Multi-Provider LLM Integration
- Fabric Chaincode Mock Testing
- REST API Handlers & Endpoints
- Frontend Dependencies & Tooling
- Priority Scoring & Heuristics
- Routing & Role-Based Access Control
- OCR Tesseract Integration
- Chaincode Deployment Scripts (Bash)
- Network Teardown Scripts (Bash)
- Network Startup Scripts (Bash)
- Channel Configuration & Network Profile
- Fabric MSP Topology & Docker Compose
- E-Waste Regulatory & Retirement Compliance
- CouchDB Rich Query State
- Chaincode Module Definition
- Root Project Module Definition

## God Nodes (most connected - your core abstractions)
1. `FabricClient` - 25 edges
2. `useRole()` - 23 edges
3. `Handlers` - 21 edges
4. `writeJSONError()` - 21 edges
5. `writeJSONResponse()` - 21 edges
6. `FabricAdapter` - 20 edges
7. `main()` - 18 edges
8. `stubBlockchainClient` - 18 edges
9. `AssetContract` - 17 edges
10. `Asset` - 14 edges

## Surprising Connections (you probably didn't know these)
- `Predictive Maintenance Automation Agent` --semantically_similar_to--> `PredictiveAgent`  [INFERRED] [semantically similar]
  blockchain-inventory-management-main/GenAI_Asset_Prioritization_Addendum.pdf → blockchain-inventory-management-main/internal/genai/predictive.go
- `Vision-Based Audit Verification Agent` --semantically_similar_to--> `VisionAgent`  [INFERRED] [semantically similar]
  blockchain-inventory-management-main/GenAI_Asset_Prioritization_Addendum.pdf → blockchain-inventory-management-main/internal/genai/vision.go
- `Hybrid Cloud-Blockchain Architecture` --implements--> `AssetContract`  [EXTRACTED]
  blockchain-inventory-management-main/ARCHITECTURE.md → blockchain-inventory-management-main/chaincode/assetcc/asset_contract.go
- `5-Factor Criticality Scoring Formula` --implements--> `scorePriority()`  [EXTRACTED]
  blockchain-inventory-management-main/TECHNICAL_ADDENDUM.md → blockchain-inventory-management-main/internal/genai/genai.go
- `Runtime Agent Dynamic Control` --references--> `Manager`  [EXTRACTED]
  blockchain-inventory-management-main/README.md → blockchain-inventory-management-main/internal/genai/manager.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Asset Criticality Scoring & Audit Flow** — tech_criticality_score_formula, assetcc_assetcontract_classifypriority, assetcc_assetcontract_updateprioritytier, assetcc_assetcontract_scheduleaudit [EXTRACTED 1.00]
- **Autonomous GenAI Agent Orchestration** — arch_genai_automation_layer, genai_genaiservice, genai_predictiveagent, genai_visionagent, genai_manager [EXTRACTED 1.00]

## Communities (30 total, 10 thin omitted)

### Community 0 - "Frontend React State & UI"
Cohesion: 0.10
Nodes (29): api, RoleContext, RoleProvider(), ROLES, useRole(), App(), NAV, Field() (+21 more)

### Community 1 - "GenAI Document & Predictive Agents"
Cohesion: 0.07
Nodes (33): Autonomous GenAI Agent Layer, NewFabricAdapter(), NewDocumentAgent(), TestDocumentAgentExtractsWarrantyAndReclassifies(), PriorityScores, New(), scorePriority(), DefaultModelProvider() (+25 more)

### Community 2 - "Fabric Gateway gRPC Client"
Cohesion: 0.06
Nodes (19): Fabric Gateway SDK Integration, Connect(), Asset, FabricClient, HistoryEntry, PriorityScores, NewFabricDriver(), TestGenAIClassification() (+11 more)

### Community 3 - "Backend Blockchain Adapter & Services"
Cohesion: 0.06
Nodes (13): BlockchainClient, ClientAsset, ClientHistoryEntry, FabricAdapter, IssueResponse, stubBlockchainClient, toClientAssetFabric(), utilizationFromHistory() (+5 more)

### Community 4 - "CLI Client Workflow & Operations"
Cohesion: 0.11
Nodes (35): classifyAsset(), complianceReport(), consumeStock(), dumpBlocks(), getHistory(), getWithRole(), issueAsset(), listAssets() (+27 more)

### Community 5 - "API Request & Response Models"
Cohesion: 0.06
Nodes (24): AssistantQueryRequest, AssistantQueryResponse, AuditTxResponse, ClassifyRequest, ClassifyResponse, ComplianceSummary, ConsumeRequest, ConsumeResponse (+16 more)

### Community 6 - "Smart Contract & Ledger Core"
Cohesion: 0.12
Nodes (16): Hybrid Cloud-Blockchain Architecture, Hybrid Ledger & Off-Chain Storage Rationale, Asset, AssetContract, HistoryEntry, PriorityScores, computeCriticalityScore(), defaultScores() (+8 more)

### Community 7 - "Multi-Provider LLM Integration"
Cohesion: 0.10
Nodes (14): priorityScorePrompt(), PriorityScores, NewGemini(), PriorityScores, NewMistral(), PriorityScores, NewOpenAI(), NewModelProviderFromConfig() (+6 more)

### Community 8 - "Fabric Chaincode Mock Testing"
Cohesion: 0.11
Nodes (19): MockContext, MockStub, TestConsumeStock(), TestIssueAsset(), TestClassifyPriority(), TestIssueAssetAutoClassify(), TestRecordAuditResult(), TestRetireAsset() (+11 more)

### Community 9 - "REST API Handlers & Endpoints"
Cohesion: 0.32
Nodes (6): Handlers, Handlers, writeJSONError(), writeJSONResponse(), net/http.Request, net/http.ResponseWriter

### Community 10 - "Frontend Dependencies & Tooling"
Cohesion: 0.10
Nodes (20): dependencies, react, react-dom, react-router-dom, devDependencies, vite, @vitejs/plugin-react, name (+12 more)

### Community 11 - "Priority Scoring & Heuristics"
Cohesion: 0.14
Nodes (6): PriorityScores, PriorityScores, heuristicScores(), DummyLLM, DummyOCR, Document Intelligence OCR/LLM Agent

### Community 12 - "Routing & Role-Based Access Control"
Cohesion: 0.22
Nodes (7): RequireRoles(), SetupRouter(), Asset Lifecycle REST API, chi.Mux, Vite React Inventory Dashboard, net/http.Handler, Handlers

## Knowledge Gaps
- **60 isolated node(s):** `assetcc`, `IssueRequest`, `ConsumeRequest`, `TransferRequest`, `ClassifyRequest` (+55 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 151 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `GenAI Document & Predictive Agents` to `Fabric Gateway gRPC Client`, `Backend Blockchain Adapter & Services`, `Routing & Role-Based Access Control`, `Multi-Provider LLM Integration`?**
  _High betweenness centrality (0.115) - this node is a cross-community bridge._
- **Why does `IssueResponse` connect `CLI Client Workflow & Operations` to `Backend Blockchain Adapter & Services`?**
  _High betweenness centrality (0.102) - this node is a cross-community bridge._
- **Why does `Asset` connect `Fabric Gateway gRPC Client` to `GenAI Document & Predictive Agents`, `Backend Blockchain Adapter & Services`?**
  _High betweenness centrality (0.095) - this node is a cross-community bridge._
- **Are the 2 inferred relationships involving `writeJSONError()` (e.g. with `.AgentControl()` and `RequireRoles()`) actually correct?**
  _`writeJSONError()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `assetcc`, `IssueRequest`, `ConsumeRequest` to the rest of the system?**
  _60 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Frontend React State & UI` be split into smaller, more focused modules?**
  _Cohesion score 0.0967741935483871 - nodes in this community are weakly interconnected._
- **Should `GenAI Document & Predictive Agents` be split into smaller, more focused modules?**
  _Cohesion score 0.07071887784921099 - nodes in this community are weakly interconnected._