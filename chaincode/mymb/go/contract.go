package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pelletier/go-toml"
	"time"
)

type TokenERC1155Contract struct {
	contractapi.Contract
}

type Token1155 struct {
	TokenID          string             `json:"TokenID"`
	CategoryCode     uint               `json:"CategoryCode"`
	PollingResultID  uint               `json:"PollingResultID"`
	TokenType        string             `json:"TokenType"`
	TotalTicket      uint               `json:"TotalTicket"`
	Amount           uint               `json:"amount"`
	TokenCreatedTime toml.LocalDateTime `json:"amount"`
}

type User struct {
	nickName
	BlockCreatedTime
	totalToken
}

type QueryResult struct {
	Key    string    `json:"Key"`
	Record Token1155 `json:"Record"`
}

const (
	tokenPrefix   = "token"
	balancePrefix = "balance"
)

func (c *TokenERC1155Contract) MintToken(ctx contractapi.TransactionContextInterface,
	tokenID string, categoryCode uint, pollingResultID uint, tokenType string,
	totalTicket uint, amount uint, ownerID string) (*Token1155, error) {

	// 유니크한 데이터 생성
	uniqueData := fmt.Sprintf("%d%d%s", ownerID, totalTicket, time.Now().String())

	// SHA256 해시 생성
	hash := sha256.New()
	hash.Write([]byte(uniqueData))
	hashBytes := hash.Sum(nil)

	// TokenID 생성 (0x를 앞에 붙여서 생성)
	tokenID = fmt.Sprintf("0x%x", hashBytes)

	// Token 생성
	token := Token1155{
		TokenID:         tokenID,
		CategoryCode:    categoryCode,
		PollingResultID: pollingResultID,
		TokenType:       tokenType,
		TotalTicket:     totalTicket,
		Amount:          amount,
		Owner:           ownerID,
	}

	// TokenID, Owner, Amount 저장
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token: %v", err)
	}

	err = ctx.GetStub().PutState(tokenKey, tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to put state: %v", err)
	}

	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{ownerID, tokenID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key for balance: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKey, []byte(fmt.Sprintf("%d", amount)))
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %v", err)
	}

	return &token, nil
}

func (c *TokenERC1155Contract) GetToken(ctx contractapi.TransactionContextInterface, tokenID string) (*Token1155, error) {
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %v", err)
	}

	if tokenBytes == nil {
		return nil, fmt.Errorf("token with ID %s does not exist", tokenID)
	}

	var token Token1155
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}

	return &token, nil
}

func (c *TokenERC1155Contract) GetAllTokens(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(tokenPrefix, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get state by partial composite key: %v", err)
	}
	defer resultsIterator.Close()

	var results []QueryResult

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next query response: %v", err)
		}

		var token Token1155
		err = json.Unmarshal(queryResponse.Value, &token)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal token: %v", err)
		}

		results = append(results, QueryResult{
			Key:    queryResponse.Key,
			Record: token,
		})
	}

	return results, nil
}

/*func (c *TokenERC1155Contract) tokenExists(ctx contractapi.TransactionContextInterface, tokenID string) (bool, error) {
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenID})
	if err != nil {
		return false, fmt.Errorf("failed to create composite key: %v", err)
	}

	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		return false, fmt.Errorf("failed to get state: %v", err)
	}

	return tokenBytes != nil, nil
}*/

func main() {
	// The main function is not required for Hyperledger Fabric chaincode
	// It's here only for demonstration purposes
	cc, err := contractapi.NewChaincode(new(TokenERC1155Contract))
	if err != nil {
		panic(err.Error())
	}
	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting TokenERC1155Contract chaincode: %s", err)
	}
}
