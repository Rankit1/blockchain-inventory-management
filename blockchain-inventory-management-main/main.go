package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"inventory-chain/internal/api"
	"inventory-chain/internal/fabricclient"
	"inventory-chain/internal/genai"
	"inventory-chain/internal/network"

	bolt "go.etcd.io/bbolt"
)

func main() {
	log.Println("Starting Inventory Management System...")

	cryptoDir := os.Getenv("FABRIC_CRYPTO_DIR")
	if cryptoDir == "" {
		cryptoDir = "./network/crypto-config"
	}

	peerAddr := os.Getenv("FABRIC_PEER_ADDRESS")
	if peerAddr == "" {
		peerAddr = "localhost:7051"
	}

	peerName := os.Getenv("FABRIC_PEER_NAME")
	if peerName == "" {
		peerName = "peer0.lab.inventory.com"
	}

	runMode := os.Getenv("RUN_MODE")
	var db *bolt.DB
	var client api.BlockchainClient
	var genaiSvc *genai.GenAIService
	var predAgent *genai.PredictiveAgent
	var visionAgent *genai.VisionAgent
	var documentAgent *genai.DocumentAgent

	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	if runMode == "fabric" {
		log.Printf("Connecting to peer at %s using crypto material from %s...\n", peerAddr, cryptoDir)
		fc, err := fabricclient.Connect(cryptoDir, peerAddr, peerName)
		if err != nil {
			log.Fatalf("Failed to connect to Fabric Gateway: %v", err)
		}
		defer fc.Close()
		client = api.NewFabricAdapter(fc)
		// Start GenAI services backed by FabricDriver
		fabricDriver := genai.NewFabricDriver(fc)
		genaiSvc = genai.New(fabricDriver, 1*time.Second)
		genaiSvc.Start()
		predAgent = genai.NewPredictive(fabricDriver, 1*time.Second)
		predAgent.Start()
		models := genai.NewModelProviderFromEnv()
		visEnabled := strings.ToLower(os.Getenv("AGENT_VISION_ENABLED"))
		if visEnabled == "" || visEnabled == "true" {
			vis := genai.NewVision(fabricDriver, 1*time.Second, models)
			vis.Start()
			visionAgent = vis
		}
		docEnabled := strings.ToLower(os.Getenv("AGENT_DOCUMENT_ENABLED"))
		if docEnabled == "" || docEnabled == "true" {
			doc := genai.NewDocumentAgent(fabricDriver, 1*time.Second, models)
			doc.Start()
			documentAgent = doc
		}
		genai.GlobalManager.Register(genaiSvc, predAgent, visionAgent, documentAgent)
		defer genaiSvc.Stop()
		defer predAgent.Stop()
		if visionAgent != nil {
			defer visionAgent.Stop()
		}
		if documentAgent != nil {
			defer documentAgent.Stop()
		}
	} else if runMode == "simulation" {
		log.Println("Starting in Simulation Mode...")
		var err error
		var svc *genai.GenAIService
		var pa *genai.PredictiveAgent
		var va *genai.VisionAgent
		var da *genai.DocumentAgent
		db, client, svc, pa, va, da, err = startSimulation()
		if err != nil {
			log.Fatalf("Failed to start simulation: %v", err)
		}
		genaiSvc = svc
		predAgent = pa
		visionAgent = va
		documentAgent = da
		// register agents with global manager for runtime control
		genai.GlobalManager.Register(genaiSvc, predAgent, visionAgent, documentAgent)
		if genaiSvc != nil {
			defer genaiSvc.Stop()
		}
		if predAgent != nil {
			defer predAgent.Stop()
		}
		if visionAgent != nil {
			defer visionAgent.Stop()
		}
		if documentAgent != nil {
			defer documentAgent.Stop()
		}
	} else {
		// Auto-detect: try Fabric first, fallback to simulation
		log.Printf("Detecting run mode... trying Fabric peer at %s...\n", peerAddr)
		fc, err := fabricclient.Connect(cryptoDir, peerAddr, peerName)
		if err != nil {
			log.Printf("Fabric connection failed: %v. Falling back to Simulation Mode.\n", err)
			var simErr error
			var svc *genai.GenAIService
			var pa *genai.PredictiveAgent
			var va *genai.VisionAgent
			var da *genai.DocumentAgent
			db, client, svc, pa, va, da, simErr = startSimulation()
			genaiSvc = svc
			predAgent = pa
			visionAgent = va
			documentAgent = da
			if genaiSvc != nil {
				defer genaiSvc.Stop()
			}
			if predAgent != nil {
				defer predAgent.Stop()
			}
			if visionAgent != nil {
				defer visionAgent.Stop()
			}
			if documentAgent != nil {
				defer documentAgent.Stop()
			}
			if simErr != nil {
				log.Fatalf("Failed to start simulation: %v", simErr)
			}
		} else {
			log.Println("Successfully connected to Fabric. Starting in Fabric Mode.")
			defer fc.Close()
			client = api.NewFabricAdapter(fc)
			// Start GenAI services backed by FabricDriver
			fabricDriver := genai.NewFabricDriver(fc)
			genaiSvc = genai.New(fabricDriver, 1*time.Second)
			genaiSvc.Start()
			predAgent = genai.NewPredictive(fabricDriver, 1*time.Second)
			predAgent.Start()
			models := genai.NewModelProviderFromEnv()
			visEnabled := strings.ToLower(os.Getenv("AGENT_VISION_ENABLED"))
			if visEnabled == "" || visEnabled == "true" {
				vis := genai.NewVision(fabricDriver, 1*time.Second, models)
				vis.Start()
				visionAgent = vis
			}
			docEnabled := strings.ToLower(os.Getenv("AGENT_DOCUMENT_ENABLED"))
			if docEnabled == "" || docEnabled == "true" {
				doc := genai.NewDocumentAgent(fabricDriver, 1*time.Second, models)
				doc.Start()
				documentAgent = doc
			}
			genai.GlobalManager.Register(genaiSvc, predAgent, visionAgent, documentAgent)
			defer genaiSvc.Stop()
			defer predAgent.Stop()
			if visionAgent != nil {
				defer visionAgent.Stop()
			}
			if documentAgent != nil {
				defer documentAgent.Stop()
			}
		}
	}

	// Connect API router endpoints
	handlers := api.NewHandlers(client)
	router := api.SetupRouter(handlers)

	log.Println("HTTP Server listening on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func startSimulation() (*bolt.DB, api.BlockchainClient, *genai.GenAIService, *genai.PredictiveAgent, *genai.VisionAgent, *genai.DocumentAgent, error) {
	// 1. Ensure the data directory exists
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 2. Open bbolt DB for persistent WorldState and Ledger
	dbPath := filepath.Join(dataDir, "inventory.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 3. Setup core blockchain network layer
	net, err := network.SetupNetwork(db)
	if err != nil {
		db.Close()
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to setup network: %w", err)
	}

	// 4. Start block event listener as a background goroutine (simulating event listener)
	go func() {
		for event := range net.Orderer.EventChan {
			log.Printf("[BLOCK COMMITTED] Index: %d, Hash: %s, TxCount: %d\n",
				event.Block.Index, event.Block.Hash, len(event.Block.Transactions))
			for _, tx := range event.Block.Transactions {
				log.Printf("  -> TxID: %s, Type: %s, Dept: %s\n", tx.TxID, tx.Type, tx.DeptID)
			}
		}
	}()

	// 5. Start GenAI automation services (simulation-only)
	driver := genai.NewSimulationDriver(net)
	svc := genai.New(driver, 1*time.Second)
	svc.Start()

	// Start predictive maintenance agent as part of the automation layer
	pred := genai.NewPredictive(driver, 1*time.Second)
	pred.Start()

	// Start Vision and Document agents (configurable via env)
	models := genai.NewModelProviderFromEnv()
	var vis *genai.VisionAgent
	var doc *genai.DocumentAgent
	visEnabled := strings.ToLower(os.Getenv("AGENT_VISION_ENABLED"))
	if visEnabled == "" || visEnabled == "true" {
		vis = genai.NewVision(driver, 1*time.Second, models)
		vis.Start()
	}
	docEnabled := strings.ToLower(os.Getenv("AGENT_DOCUMENT_ENABLED"))
	if docEnabled == "" || docEnabled == "true" {
		doc = genai.NewDocumentAgent(driver, 1*time.Second, models)
		doc.Start()
	}

	return db, api.NewSimulationAdapter(net), svc, pred, vis, doc, nil
}
