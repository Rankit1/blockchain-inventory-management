package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"inventory-chain/internal/api"
	"inventory-chain/internal/fabricclient"
	"inventory-chain/internal/genai"
)

func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

func agentEnabled(envVar string) bool {
	v := strings.ToLower(os.Getenv(envVar))
	return v == "" || v == "true"
}

func main() {
	loadDotEnv(".env")
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

	log.Printf("Connecting to peer at %s using crypto material from %s...\n", peerAddr, cryptoDir)
	fc, err := fabricclient.Connect(cryptoDir, peerAddr, peerName)
	if err != nil {
		log.Fatalf("Failed to connect to Fabric Gateway: %v", err)
	}
	defer fc.Close()

	client := api.NewFabricAdapter(fc)
	fabricDriver := genai.NewFabricDriver(fc)

	models := genai.NewModelProviderFromEnv()
	genai.SetGlobalModelProvider(models)

	genaiSvc := genai.New(fabricDriver, 1*time.Second)
	genaiSvc.Start()
	defer genaiSvc.Stop()

	predAgent := genai.NewPredictive(fabricDriver, 1*time.Second)
	predAgent.Start()
	defer predAgent.Stop()

	var visionAgent *genai.VisionAgent
	if agentEnabled("AGENT_VISION_ENABLED") {
		visionAgent = genai.NewVision(fabricDriver, 1*time.Second, models)
		visionAgent.Start()
		defer visionAgent.Stop()
	}

	var documentAgent *genai.DocumentAgent
	if agentEnabled("AGENT_DOCUMENT_ENABLED") {
		documentAgent = genai.NewDocumentAgent(fabricDriver, 1*time.Second, models)
		documentAgent.Start()
		defer documentAgent.Stop()
	}

	genai.GlobalManager.Register(genaiSvc, predAgent, visionAgent, documentAgent)

	handlers := api.NewHandlers(client)
	router := api.SetupRouter(handlers)

	log.Println("HTTP Server listening on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
