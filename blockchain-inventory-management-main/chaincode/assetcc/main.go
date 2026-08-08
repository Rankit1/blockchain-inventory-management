package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	assetContract := new(AssetContract)

	cc, err := contractapi.NewChaincode(assetContract)
	if err != nil {
		log.Panicf("Error creating assetcc chaincode: %v", err)
	}

	if err := cc.Start(); err != nil {
		log.Panicf("Error starting assetcc chaincode: %v", err)
	}
}
