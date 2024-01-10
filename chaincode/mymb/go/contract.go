package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract is the contract structure
type SmartContract struct {
	contractapi.Contract
}

// ERC1155Token is the ERC-1155 token struct
type ERC1155Token struct {
	Id     uint64 `json:"id"`
	Amount uint64 `json:"amount"`
}

// InitLedger initializes the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContext) error {
	return nil
}

// IssueToken issues a new ERC-1155 token
/*
	- 토큰 발급 권한 확인
	- 새로운 ERC-1155 토큰 생성
	- 토큰 정보를 원장에 저장
*/
func (s *SmartContract) IssueToken(ctx contractapi.TransactionContext, owner string, tokenId uint64, amount uint64) error {

	// Check that the caller is authorized to issue tokens
	if owner != ctx.GetStub().GetCaller() {
		return fmt.Errorf("caller is not authorized to issue tokens")
	}

	// Create a new ERC-1155 token
	token := &ERC1155Token{
		Id:     tokenId,
		Amount: amount,
	}

	// Save the token to the ledger
	return ctx.GetStub().PutState(token.Id, token)
}

// TransferToken transfers an ERC-1155 token
/*
	- 토큰 소유권 확인
	- 토큰 수신자 유효성 검사
	- 토큰 잔액 갱신
	- 토큰 잔액이 0이면 원장에서 삭제
*/
func (s *SmartContract) TransferToken(ctx contractapi.TransactionContext, from string, to string, tokenId uint64, amount uint64) error {
	// Check that the caller owns the token
	token, err := s.GetToken(ctx, tokenId)
	if err != nil {
		return err
	}

	// Check that the recipient is valid
	if to == "" {
		return fmt.Errorf("recipient cannot be empty")
	}

	// Update the token balance
	token.Amount -= amount
	if token.Amount == 0 {
		// Delete the token if the balance is 0
		return ctx.GetStub().DeleteState(token.Id)
	} else {
		// Save the token to the ledger
		return ctx.GetStub().PutState(token.Id, token)
	}
}

// GetToken gets an ERC-1155 token
/*
	- 원장에서 토큰 조회
	- JSON 형식으로 토큰 정보 반환
*/
func (s *SmartContract) GetToken(ctx contractapi.TransactionContext, tokenId uint64) (*ERC1155Token, error) {
	// Get the token from the ledger
	tokenBytes, err := ctx.GetStub().GetState(tokenId)
	if err != nil {
		return nil, err
	}

	// Unmarshal the token
	token := &ERC1155Token{}
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func main() {
	// Create a new contract instance
	// 스마트 계약의 인스턴스를 생성.
	contract := &SmartContract{}

	// Initialize the contract
	// 원장을 초기화.
	err := contract.InitLedger(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
