package apps

import (
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var SubnetAppRegistryABI string = `[ 
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "start",
                "type": "uint256"
            },
            {
                "internalType": "uint256",
                "name": "end",
                "type": "uint256"
            }
        ],
        "name": "listApps",
        "outputs": [
            {
                "components": [
                    {
                        "internalType": "string",
                        "name": "peerId",
                        "type": "string"
                    },
                    {
                        "internalType": "address",
                        "name": "owner",
                        "type": "address"
                    },
                    {
                        "internalType": "string",
                        "name": "name",
                        "type": "string"
                    },
                    {
                        "internalType": "string",
                        "name": "symbol",
                        "type": "string"
                    },
                    {
                        "internalType": "uint256",
                        "name": "budget",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "spentBudget",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "maxNodes",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "minCpu",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "minGpu",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "minMemory",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "minUploadBandwidth",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "minDownloadBandwidth",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "nodeCount",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "pricePerCpu",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "pricePerGpu",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "pricePerMemoryGB",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "pricePerStorageGB",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint256",
                        "name": "pricePerBandwidthGB",
                        "type": "uint256"
                    },
                    {
                        "internalType": "uint8",
                        "name": "paymentMethod",
                        "type": "uint8"
                    }
                ],
                "internalType": "struct SubnetAppRegistry.App[]",
                "name": "",
                "type": "tuple[]"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    }
]`

var subnetABI abi.ABI

func init() {
	var err error
	// Phân tích ABI từ chuỗi JSON
	subnetABI, err = abi.JSON(strings.NewReader(SubnetAppRegistryABI))
	if err != nil {
		log.Panicf("Failed to parse ABI: %v", err)
	}
}
