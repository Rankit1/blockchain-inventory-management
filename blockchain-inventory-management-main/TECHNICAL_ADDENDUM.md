# Technical Addendum: GenAI-Augmented Asset Prioritization & Automation Workflow

**Document Version:** 1.0 (Draft for Stakeholder Review)
**Applies To:** Hybrid Cloud–Blockchain Inventory Management System
**Underlying Platform:** Hyperledger Fabric v2.5 (inventorychain) + Go REST API Gateway

---

## 1. Executive Summary

This addendum extends the existing Hybrid Cloud–Blockchain Inventory Management System with a
**priority-based asset classification framework**, **nine GenAI-driven automation services**, and a
corresponding set of **blockchain data-model, REST API, and role-based access control (RBAC)**
enhancements.

The enhancements are designed to improve **asset utilization, preventive/corrective maintenance,
auditing, procurement, and compliance** while **strictly preserving the existing Hyperledger Fabric
architecture and consensus mechanism** (single-org, 4-peer Raft ordering service, CouchDB world
state).

### 1.1 Business Value

- **Cost Optimization:** Enables dynamic reorder points and idle-asset reallocation, reducing
  over-procurement and under-utilization.
- **Risk Reduction:** Prioritized P1 assets receive real-time monitoring and monthly audits,
  lowering the probability of critical asset failure.
- **Regulatory Readiness:** Automated compliance report generation aligned with India's
  E-Waste Management Rules, 2022 and internal audit policy.
- **Operational Efficiency:** Conversational assistant and document intelligence reduce manual data
  entry and procurement lead time.
- **Auditability:** Every classification, audit, and retirement event is written to the immutable
  blockchain ledger.

### 1.2 Scope

This addendum covers the business rules, data model, API surface, RBAC matrix, and phased rollout
for the GenAI augmentation. Cost estimates are provided separately and are **out of scope for code
implementation**.

---

## 2. Asset Priority Classification (P1–P3)

Assets are automatically categorized into one of three priority tiers based on a weighted
**Criticality Score**. The score is computed from five criteria, each rated on a **1–5 scale**.

### 2.1 Scoring Criteria and Weights

| # | Criterion                        | Weight | Description |
|---|----------------------------------|--------|-------------|
| 1 | Business Criticality             | 30%    | Impact of asset failure on core operations and SLAs |
| 2 | Replacement Cost                 | 20%    | Financial value to replace the asset |
| 3 | Replacement Lead Time            | 20%    | Procurement lead time to restore capability |
| 4 | Safety / Compliance Impact       | 15%    | Regulatory, environmental, or safety obligations |
| 5 | Redundancy / Availability        | 15%    | Availability of backups or failover capacity |

### 2.2 Criticality Score Calculation

```
Criticality Score = (0.30 × BusinessCriticality)
                 + (0.20 × ReplacementCost)
                 + (0.20 × ReplacementLeadTime)
                 + (0.15 × SafetyComplianceImpact)
                 + (0.15 × RedundancyAvailability)
```

The weighted result is normalized to a **1.0–5.0** scale.

### 2.3 Tier Assignment

| Tier | Score Range  | Audit Cadence              | Monitoring                       |
|------|--------------|----------------------------|----------------------------------|
| **P1 (High)**   | ≥ 4.0     | Monthly                    | Real-time AI monitoring          |
| **P2 (Medium)** | 2.5 – 3.9  | Semi-annual                | Weekly anomaly detection         |
| **P3 (Low)**    | < 2.5     | Annual                     | Ledger record only               |

### 2.4 Tier Boundaries

A scoring matrix (recommended as a table or heat-map diagram in the final document):

| Asset Class                        | Criticality | Cost | Lead Time | Safety | Redundancy | Score | Tier |
|------------------------------------|-------------|------|-----------|--------|------------|-------|------|
| Production Server                  | 5           | 5    | 4         | 3      | 2          | 4.15  | P1   |
| Network Switch                     | 5           | 3    | 4         | 2      | 3          | 3.80  | P2   |
| Diagnostic Equipment               | 4           | 5    | 3         | 4      | 2          | 3.80  | P2   |
| Secure Storage System              | 5           | 5    | 5         | 4      | 1          | 4.50  | P1   |
| Office Furniture                   | 1           | 1    | 2         | 1      | 5          | 1.70  | P3   |

### 2.5 Service-Level Requirements per Tier

| Requirement       | P1                                        | P2                       | P3            |
|-------------------|-------------------------------------------|--------------------------|---------------|
| Tagging           | RFID + QR                                 | QR                       | Basic tag     |
| AMC               | Mandatory                                 | Recommended              | Not required  |
| Insurance         | Mandatory                                 | Recommended              | Not required  |
| Depreciation      | 5-year schedule                           | Standard                 | Standard      |
| Disposal          | Dual-signature + E-Waste rules            | E-Waste rules            | Standard      |

### 2.6 Manual Override

Managers may **manually override** a computed tier (e.g., promote a P2 asset to P1 during a
project window). The override **must** include a recorded justification and is written to the
blockchain as an immutable `PRIORITY_UPDATE` transaction for auditability.

---

## 3. Asset Management Lifecycle

The system supports the complete enterprise asset lifecycle. Blockchain transactions anchor each
stage, and GenAI services augment monitoring and decision-making.

| Stage                                   | Blockchain Interaction                 | GenAI / Automation Input |
|-----------------------------------------|----------------------------------------|--------------------------|
| 1. Planning & Demand Forecasting        | Budget asset demand                    | Reorder optimization, demand forecast |
| 2. Procurement & Acquisition            | Purchase record anchor                 | Document intelligence, vendor negotiation |
| 3. Asset Tagging & Registration         | `IssueAsset` — QR/RFID metadata on-chain | Document intelligence (invoices, warranty) |
| 4. Deployment to Departments            | `TransferAsset`                        | —                        |
| 5. Preventive / Corrective Maintenance  | `scheduleAudit` / maintenance tx       | Predictive maintenance agent |
| 6. Continuous Monitoring                | Consume / utilization transactions     | Utilization & idle detection |
| 7. Depreciation Calculation             | Off-chain analytics over ledger data   | Utilization & idle detection |
| 8. Physical & AI-Assisted Audits        | `scheduleAudit`, `recordAuditResult`   | Vision-based audit agent |
| 9. Retirement & Compliant Disposal      | `retireAsset` (dual-signature)         | Compliance report generator |

> **P1 requirement:** RFID + QR tags, mandatory AMC, mandatory insurance, and a five-year
> depreciation schedule. Disposal follows **India's E-Waste Management Rules, 2022** with
> **dual-signature approval** recorded on-chain.

---

## 4. GenAI Agent Architecture

A new **GenAI Automation Layer** sits between the AI services and the cloud database, consuming the
existing **Kafka/Pulsar event stream** without altering Hyperledger Fabric consensus. Nine agents
are introduced.

| # | Agent                              | Inputs                                                     | Outputs                                          | Blockchain Interaction |
|---|------------------------------------|------------------------------------------------------------|--------------------------------------------------|------------------------|
| 1 | **Priority Classification Agent**  | Asset attributes, scoring criteria                         | `priorityTier`, `criticalityScore`               | `classifyPriority`, `updatePriorityTier` |
| 2 | **Predictive Maintenance Agent**   | Telemetry, maintenance history (LSTM/Prophet)              | Failure probability, maintenance schedule         | `scheduleAudit` |
| 3 | **Vision-Based Audit Agent**       | Uploaded asset images                                      | Verification result (present / damaged / missing) | `recordAuditResult` |
| 4 | **Document Intelligence Agent**    | Invoices, warranty, AMC PDFs (OCR + LLM)                   | Extracted metadata, expiry dates                   | `classifyPriority` enrichment |
| 5 | **Dynamic Reorder Optimization**   | Consumption history, priority tier                         | Predicted reorder point, recommended quantity      | Event-based alerts |
| 6 | **Utilization & Idle Detection**   | Ledger transactions, telemetry                             | Utilization rate, idle-asset list                  | Read-only queries |
| 7 | **Compliance Report Generator**    | Audit results, lifecycle state, disposal records           | Regulatory report documents                        | Read-only queries |
| 8 | **Conversational Asset Assistant** | Natural-language user queries (LLM)                        | Operation commands / answers                       | `IssueAsset`, `ConsumeStock`, etc. |
| 9 | **Vendor Negotiation Enrichment**  | Vendor catalogs, procurement history                       | Negotiation drafts, price benchmarks               | Off-chain procurement data |

### 4.1 Agent Interaction Flow

```
┌──────────────┐      ┌─────────────────────────────┐
│  Client / UI │ ───▶ │  REST API Gateway (Go)      │
└──────────────┘      └──────────────┬──────────────┘
                                     │
                         ┌───────────▼────────────┐
                         │   GenAI Automation     │
                         │   Layer (Kafka/Pulsar) │
                         └───┬─────────┬──────────┘
                             │         │
                 ┌───────────▼───┐   ┌─▼───────────────────┐
                 │  Fabric       │   │  Cloud DB / Vector  │
                 │  Network      │   │  DB (off-chain)     │
                 └───────────────┘   └─────────────────────┘
```

- The GenAI layer **subscribes** to block/event streams and **invokes** the existing REST API or
  Fabric Gateway for state transitions.
- **Hyperledger Fabric consensus and channel configuration are unchanged.**

---

## 5. Technical Implementation & Architecture

### 5.1 GenAI Automation Layer Integration

New components in the automation layer:

- **Classification Service** — applies the scoring model on asset creation/update.
- **Predictive Maintenance Consumer** — consumes telemetry + maintenance history.
- **Vision/OCR Service** — image + document processing.
- **Utilization Analysis Batch Job** — scheduled aggregation of ledger data.
- **Compliance Report Generator** — templated report production.

Optional **IoT telemetry** may be connected for P1 assets to improve predictive maintenance
accuracy.

### 5.2 Blockchain & Data Model Enhancements

#### 5.2.1 New Asset Attributes

| Field              | Type   | Allowed Values / Example           | Notes                                  |
|--------------------|--------|------------------------------------|----------------------------------------|
| `priorityTier`     | string | `P1`, `P2`, `P3`                   | Computed by classification service     |
| `criticalityScore` | number | `1.0` – `5.0`                      | Weighted score                         |
| `lastAuditDate`    | string | `2026-08-01` (ISO 8601)            | Updated by `recordAuditResult`         |
| `utilizationRate`  | number | `0.0` – `1.0`                      | Aggregated by utilization analysis     |
| `warrantyExpiry`   | string | `2027-03-15` (ISO 8601)            | Extracted by document intelligence     |
| `amcExpiry`        | string | `2026-12-31` (ISO 8601)            | Annual maintenance contract expiry     |
| `lifecycleState`   | string | `ACTIVE`, `MAINTENANCE`, `IDLE`, `RETIRED` | Managed via lifecycle transactions |

> All fields are additive — the existing `assetId`, `deptId`, `name`, `category`, `qty`,
> `threshold`, and `updatedAt` fields remain unchanged, preserving backward compatibility.

#### 5.2.2 New Chaincode Functions

| Function                | Description                                                        | Transaction Type     |
|-------------------------|--------------------------------------------------------------------|----------------------|
| `classifyPriority()`    | Computes `criticalityScore` + `priorityTier` for an asset          | `PRIORITY_CLASSIFY`  |
| `updatePriorityTier()`  | Manual override of tier with recorded justification                | `PRIORITY_UPDATE`    |
| `scheduleAudit()`       | Records a planned audit (date + scope) for an asset                | `AUDIT_SCHEDULE`     |
| `recordAuditResult()`   | Records audit outcome and updates `lastAuditDate`                  | `AUDIT_RESULT`       |
| `retireAsset()`         | Marks asset `RETIRED` with reason (dual-signature enforced)        | `RETIRE`             |

### 5.3 Database Schema Enhancements

- **CouchDB Indexes (on-chain world state):**
  - `indexPriorityTier.json` — `["docType", "priorityTier"]`
  - `indexLifecycle.json` — `["docType", "lifecycleState"]`
  - `indexAudit.json` — `["docType", "lastAuditDate"]`
- **PostgreSQL (off-chain analytics):**
  - `utilization` — per-asset aggregation of consume/transfer activity
  - `maintenance_predictions` — model outputs from predictive maintenance
  - `ai_audit_logs` — vision audit results and confidence scores
  - `vendor_negotiations` — negotiation drafts, benchmarks, approvals
- **Vector DB:** embeddings for warranty/AMC documents to power conversational retrieval.

### 5.4 REST API Additions

All new endpoints are protected with **JWT authentication and RBAC**.

| Method | Endpoint                         | Description                                  | Required Role            |
|--------|----------------------------------|----------------------------------------------|--------------------------|
| POST   | `/api/assets/classify`           | Compute and persist priority classification  | AI_OPS, ASSET_AUDITOR    |
| POST   | `/api/assets/update-priority`    | Manual priority override + justification     | ASSET_AUDITOR, IT_ADMIN  |
| POST   | `/api/assets/schedule-audit`     | Schedule an audit for an asset               | ASSET_AUDITOR            |
| POST   | `/api/assets/record-audit`       | Record audit result                          | ASSET_AUDITOR            |
| POST   | `/api/assets/retire`             | Retire asset (dual-signature)                | IT_ADMIN                 |
| GET    | `/api/reports/utilization`       | Utilization & idle-asset report              | STORE_MANAGER, IT_ADMIN  |
| GET    | `/api/reports/compliance`        | Compliance report summary                    | IT_ADMIN                 |
| POST   | `/api/assistant/query`           | Conversational assistant interface           | DEPARTMENT_USER + roles  |

> Predictive-maintenance and vision-audit endpoints depend on GenAI cloud services and are listed
> here for API contract completeness; their heavy-AI implementations are deferred to Phases 3b/3c.

### 5.5 Updated RBAC Matrix

| Role               | Classify | Update Priority | Schedule Audit | Record Audit | Retire | Utilization Report | Compliance Report | Assistant |
|--------------------|:--------:|:---------------:|:--------------:|:------------:|:------:|:------------------:|:-----------------:|:---------:|
| Asset Auditor      | ✅       | ✅              | ✅             | ✅           | ❌     | ✅                 | ❌                | ❌        |
| AI Operations      | ✅       | ❌              | ❌             | ❌           | ❌     | ❌                 | ❌                | ❌        |
| Store Manager      | ❌       | ❌              | ❌             | ❌           | ❌     | ✅                 | ❌                | ❌        |
| Department User    | ❌       | ❌              | ❌             | ❌           | ❌     | ❌                 | ❌                | ✅        |
| IT Administrator   | ✅       | ✅              | ✅             | ✅           | ✅     | ✅                 | ✅                | ✅        |
| System Admin       | ✅       | ✅              | ✅             | ✅           | ✅     | ✅                 | ✅                | ✅        |

---

## 6. Phased Rollout Strategy

The rollout is split into three phases. **Phase 3a is the critical path** — every downstream AI
capability depends on an accurate priority tier for every asset.

### 6.1 Phase 3a — Foundation (Priority Classification + Data Model)

- Introduce `priorityTier`, `criticalityScore`, and `lifecycleState` fields.
- Implement `classifyPriority()` and `updatePriorityTier()`.
- Implement dynamic reorder optimization (tier-aware reorder points).
- Add CouchDB indexes and RBAC for classification roles.

### 6.2 Phase 3b — Predictive Maintenance + AI Audit

- Predictive maintenance consumer (LSTM/Prophet) over telemetry.
- Vision-based audit verification (`recordAuditResult` enrichment).
- `scheduleAudit` / `recordAuditResult` end-to-end.

### 6.3 Phase 3c — Conversational + Reporting + Procurement

- Conversational asset assistant.
- Compliance report generator.
- Vendor negotiation enrichment and off-chain procurement data.

> **Recommended first step:** implement the `priorityTier` field and `classifyPriority()` function;
> all later capabilities depend on asset prioritization.

---

## 7. Security, Compliance, and Risk Management

### 7.1 Security

- **Authentication:** JWT-issued tokens; all `/api` routes validated.
- **Authorization:** RBAC enforced via middleware; role claims embedded in JWT.
- **On-chain integrity:** classification, audit, and retirement events are immutable ledger
  transactions; manual overrides always carry a justification.
- **Data isolation:** off-chain analytics (PostgreSQL, Vector DB) hold no authoritative state; the
  blockchain remains the system of record.

### 7.2 Compliance

- **E-Waste Management Rules, 2022 (India):** enforced via `retireAsset` dual-signature flow.
- **Internal audit policy:** tier-based audit cadence (P1 monthly, P2 semi-annual, P3 annual).
- **Data protection:** PII minimisation in AI pipelines; retention policies for OCR/Vision assets.

### 7.3 Risk Management

| Risk                                        | Mitigation                                             |
|---------------------------------------------|--------------------------------------------------------|
| AI misclassification                        | Manual override with justification; human approval     |
| Vision OCR false positives                  | Confidence thresholds + auditor review                 |
| Model drift in predictive maintenance       | Scheduled model re-training; telemetry quality gates   |
| High GenAI operational cost                 | Tiered monitoring (only P1 uses real-time AI)          |
| Vendor lock-in on cloud AI providers        | Abstraction layer; provider-agnostic model interfaces  |

### 7.4 Estimated Additional Cost (Informational)

> Provided for budgeting only — not implemented in code.

| Component                 | Monthly Estimate (INR) |
|---------------------------|------------------------|
| Vision/OCR                | ₹4,500 – ₹6,500        |
| Predictive Maintenance    | ₹3,000 – ₹5,000        |
| Vector Storage            | ₹1,200 – ₹2,000        |
| Kafka Consumers           | ₹2,500 – ₹4,000        |
| Conversational Assistant  | ₹6,500 – ₹9,000        |
| **Estimated Total**       | **₹17,700 – ₹26,500**  |

*(Assumes ~200 users and ~50,000 transactions/month.)*

---

## 8. Implementation Status in Codebase

The following are implemented as of this addendum's code delivery (Phase 3a foundation):

- Asset data model extended with priority/lifecycle fields (Fabric chaincode + REST client DTOs).
- `ClassifyPriority`, `UpdatePriorityTier`, `ScheduleAudit`, `RecordAuditResult`, `RetireAsset`
  implemented in the Fabric `assetcc` contract, submitted via the Fabric Gateway SDK.
- `GenAIService` scores the five priority criteria via a real LLM call (OpenAI/Gemini/Mistral),
  falling back to a deterministic heuristic when no provider is configured.
- REST endpoints and RBAC middleware added.
- CouchDB index documents added for `priorityTier`, `lifecycleState`, and `lastAuditDate`.

Items dependent on paid/cloud services (Vision-LLM photo audits, LSTM predictive maintenance,
Vector DB, conversational tool-calling LLM, Kafka consumers) are intentionally **not**
implemented in code.
