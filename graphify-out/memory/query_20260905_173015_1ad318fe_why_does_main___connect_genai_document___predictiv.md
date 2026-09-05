---
type: "query"
date: "2026-09-05T17:30:15.890070+00:00"
question: "Why does main() connect GenAI Document & Predictive Agents to Fabric Gateway gRPC Client, Backend Blockchain Adapter & Services, Routing & Role-Based Access Control, and Multi-Provider LLM Integration?"
contributor: "graphify"
outcome: "useful"
source_nodes: ["main()", "Connect()", "NewFabricAdapter()", "NewFabricDriver()", "NewModelProviderFromEnv()", "SetupRouter()"]
---

# Q: Why does main() connect GenAI Document & Predictive Agents to Fabric Gateway gRPC Client, Backend Blockchain Adapter & Services, Routing & Role-Based Access Control, and Multi-Provider LLM Integration?

## Answer

Expanded from original query via vocab: ['main', 'fabric', 'gateway', 'client', 'adapter', 'router', 'genai', 'predictive', 'document', 'provider']. main() in main.go serves as the composition root and dependency injection orchestrator. It connects to the Fabric peer via Connect(), builds the FabricDriver and FabricAdapter, configures LLM/OCR providers via NewModelProviderFromEnv(), instantiates the autonomous agents (GenAIService, PredictiveAgent, VisionAgent, DocumentAgent), injects these into NewHandlers(), and attaches them to HTTP routes in SetupRouter().

## Outcome

- Signal: useful

## Source Nodes

- main()
- Connect()
- NewFabricAdapter()
- NewFabricDriver()
- NewModelProviderFromEnv()
- SetupRouter()