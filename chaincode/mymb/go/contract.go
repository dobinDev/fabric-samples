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
	Id     string `json:"id"`
	Owner  string `json:"owner"`
	Amount uint64 `json:"amount"`
}

// InitLedger initializes the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// 초기화 로직 추가
	// 예: 특정 토큰을 미리 발행하거나 초기 상태를 설정하는 등의 동작
	return nil
}

// IssueToken issues a new ERC-1155 token
/*
	- 토큰 발급 권한 확인
	- 새로운 ERC-1155 토큰 생성
	- 토큰 정보를 원장에 저장
*/
func (s *SmartContract) IssueToken(ctx contractapi.TransactionContextInterface, owner string, tokenId uint64, amount uint64) error {

	// Check that the caller is authorized to issue tokens
	caller, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get caller identity: %w", err)
	}

	if owner != caller {
		return fmt.Errorf("caller is not authorized to issue tokens")
	}

	// Create a new ERC-1155 token
	token := &ERC1155Token{
		Id:     fmt.Sprintf("%d", tokenId),
		Owner:  owner,
		Amount: amount,
	}

	// Save the token to the ledger
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	err = ctx.GetStub().PutState(token.Id, tokenBytes)
	if err != nil {
		return fmt.Errorf("failed to put state: %w", err)
	}
	return nil
}

// TransferToken transfers an ERC-1155 token
/*
	- 토큰 소유권 확인
	- 토큰 수신자 유효성 검사
	- 토큰 잔액 갱신
	- 토큰 잔액이 0이면 원장에서 삭제
*/
func (s *SmartContract) TransferToken(ctx contractapi.TransactionContextInterface, from string, to string, tokenId uint64, amount uint64) error {

	// Check that the caller owns the token
	fromToken, err := s.getToken(ctx, tokenId)
	if err != nil {
		return err
	}

	if fromToken.Owner != from {
		return fmt.Errorf("caller is not the owner of the token")
	}

	// Update the token balance
	fromToken.Amount -= amount

	// Get or create the destination token
	toToken, err := s.getToken(ctx, tokenId)
	if err != nil {
		toToken = &ERC1155Token{
			Id:     fmt.Sprintf("%d", tokenId),
			Owner:  to,
			Amount: 0,
		}
	}
	toToken.Amount += amount

	// Save the tokens to the ledger
	if fromToken.Amount == 0 {
		err = ctx.GetStub().DelState(fromToken.Id)
	} else {
		tokenBytes, err := json.Marshal(fromToken)
		if err != nil {
			return fmt.Errorf("failed to marshal token: %w", err)
		}
		err = ctx.GetStub().PutState(fromToken.Id, tokenBytes)
		if err != nil {
			return fmt.Errorf("failed to put state for fromToken: %w", err)
		}
	}

	tokenBytes, err := json.Marshal(toToken)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	err = ctx.GetStub().PutState(toToken.Id, tokenBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for toToken: %w", err)
	}

	return nil
}

// Get the token from the ledger and unmarshal it from JSON.
/*
	- 원장에서 토큰 조회
	- JSON 형식으로 토큰 정보 반환
*/
// Get the token from the ledger and unmarshal it from JSON.
func (s *SmartContract) getToken(ctx contractapi.TransactionContextInterface, tokenId uint64) (*ERC1155Token, error) {
	tokenBytes, err := ctx.GetStub().GetState(fmt.Sprintf("%d", tokenId))
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	if tokenBytes == nil {
		return nil, fmt.Errorf("token not found")
	}
	token := &ERC1155Token{}
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return token, nil
}

func main() {
	// Create a new contract instance
	contract := new(SmartContract)

	// Initialize the contract
	ctx := contractapi.TransactionContext{} // 올바른 트랜잭션 컨텍스트를 생성해야 함
	err := contract.InitLedger(&ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
}
