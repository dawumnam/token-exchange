package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/dawumnam/token-trader/config"
	"github.com/dawumnam/token-trader/contracts"
	"github.com/dawumnam/token-trader/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	lineaSepoliaRPC        = "https://rpc.sepolia.linea.build"
	gasLimit        uint64 = 3000000
	gasPrice               = 1000000000 // 1 Gwei
)

var (
	platformAddress = config.Envs.PlatformAddress
	privateKey      = config.Envs.ChainPrivateKey
)

type TokenManager struct {
	client  *ethclient.Client
	auth    *bind.TransactOpts
	address common.Address
}

func NewTokenManager() (*TokenManager, error) {
	client, err := ethclient.Dial(lineaSepoliaRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorized transactor: %v", err)
	}

	auth.GasLimit = gasLimit
	auth.GasPrice = big.NewInt(gasPrice)

	return &TokenManager{
		client:  client,
		auth:    auth,
		address: fromAddress,
	}, nil
}

func (tm *TokenManager) DeployToken(payload types.IssueTokenPayload) (*types.Token, error) {
	initialSupply, ok := new(big.Int).SetString(payload.InitialSupply, 10)
	if !ok {
		return nil, fmt.Errorf("invalid initial supply")
	}

	address, tx, _, err := contracts.DeployContracts(
		tm.auth,
		tm.client,
		payload.Name,
		payload.Symbol,
		initialSupply,
		common.HexToAddress(platformAddress),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %v", err)
	}

	_, err = bind.WaitMined(context.Background(), tm.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for contract deployment: %v", err)
	}

	return &types.Token{
		ContractAddress: address.Hex(),
		Name:            payload.Name,
		Symbol:          payload.Symbol,
		OwnerID:         uint(tm.auth.From.Big().Uint64()),
	}, nil
}

func (tm *TokenManager) TransferToken(tokenAddress string, to string, amount *big.Int) error {
	token, err := contracts.NewContracts(common.HexToAddress(tokenAddress), tm.client)
	if err != nil {
		return fmt.Errorf("failed to instantiate a Token contract: %v", err)
	}

	tx, err := token.Transfer(tm.auth, common.HexToAddress(to), amount)
	if err != nil {
		return fmt.Errorf("failed to transfer tokens: %v", err)
	}

	_, err = bind.WaitMined(context.Background(), tm.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transfer transaction: %v", err)
	}

	return nil
}

func (tm *TokenManager) GetBalance(tokenAddress string, address string) (*big.Int, error) {
	token, err := contracts.NewContracts(common.HexToAddress(tokenAddress), tm.client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Token contract: %v", err)
	}

	balance, err := token.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve balance: %v", err)
	}

	return balance, nil
}

func (tm *TokenManager) GetPlatformAddress(tokenAddress string) (string, error) {
	token, err := contracts.NewContracts(common.HexToAddress(tokenAddress), tm.client)
	if err != nil {
		return "", fmt.Errorf("failed to instantiate a Token contract: %v", err)
	}

	platformAddr, err := token.PlatformAddress(&bind.CallOpts{})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve platform address: %v", err)
	}

	return platformAddr.Hex(), nil
}

// func main() {
// 	tokenManager, err := NewTokenManager()
// 	if err != nil {
// 		log.Fatalf("Error initializing tokenManager: %v", err)
// 	}

// 	err = tokenManager.TransferToken("0xCbe58bEFBEfDB02cD2cfEcCB5304E853b04864A1", "0x720cD79c896829f6142569EAdc46EBc9B497396C", big.NewInt(100000000000000000))
// 	if err != nil {
// 		log.Fatalf("Error when transferring tokens: %v", err)
// 	}

// newToken, err := tokenManager.DeployToken(types.IssueTokenPayload{
// 	Name:          "Test Token",
// 	Symbol:        "TST",
// 	InitialSupply: "1000000000000000000", // 1 token with 18 decimals
// })
// if err != nil {
// 	log.Fatalf("Error deploying token: %v", err)
// }

// fmt.Printf("New token deployed at address: %s\n", newToken.ContractAddress)

// balance, err := tokenManager.GetBalance(newToken.ContractAddress, tokenManager.address.Hex())
// if err != nil {
// 	log.Fatalf("Error getting balance: %v", err)
// }

// fmt.Printf("Balance of token owner: %s\n", balance.String())

// platformAddr, err := tokenManager.GetPlatformAddress(newToken.ContractAddress)
// if err != nil {
// 	log.Fatalf("Error getting platform address: %v", err)
// }

// fmt.Printf("Platform address: %s\n", platformAddr)

// }
