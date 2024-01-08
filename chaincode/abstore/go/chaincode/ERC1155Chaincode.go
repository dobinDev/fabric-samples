package chaincode

/*
	2021 Baran Kılıç <baran.kilic@boun.edu.tr>

	SPDX-License-Identifier: Apache-2.0
*/

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

const uriKey = "uri"

const balancePrefix = "account~tokenId~sender"
const approvalPrefix = "account~operator"

const minterMSPID = "Org1MSP"

// Define key names for options
const nameKey = "name"
const symbolKey = "symbol"

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	contractapi.Contract
}

// TransferSingle MUST emit when a single token is transferred, including zero
// value transfers as well as minting or burning.
// The operator argument MUST be msg.sender.
// The from argument MUST be the address of the holder whose balance is decreased.
// The to argument MUST be the address of the recipient whose balance is increased.
// The id argument MUST be the token type being transferred.
// The value argument MUST be the number of tokens the holder balance is decreased
// by and match what the recipient balance is increased by.
// When minting/creating tokens, the from argument MUST be set to `0x0` (i.e. zero address).
// When burning/destroying tokens, the to argument MUST be set to `0x0` (i.e. zero address).
type TransferSingle struct {
	Operator string `json:"operator"`
	From     string `json:"from"`
	To       string `json:"to"`
	ID       uint64 `json:"id"`
	Value    uint64 `json:"value"`
}

// TransferBatch MUST emit when tokens are transferred, including zero value
// transfers as well as minting or burning.
// The operator argument MUST be msg.sender.
// The from argument MUST be the address of the holder whose balance is decreased.
// The to argument MUST be the address of the recipient whose balance is increased.
// The ids argument MUST be the list of tokens being transferred.
// The values argument MUST be the list of number of tokens (matching the list
// and order of tokens specified in _ids) the holder balance is decreased by
// and match what the recipient balance is increased by.
// When minting/creating tokens, the from argument MUST be set to `0x0` (i.e. zero address).
// When burning/destroying tokens, the to argument MUST be set to `0x0` (i.e. zero address).
type TransferBatch struct {
	Operator string   `json:"operator"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	IDs      []uint64 `json:"ids"`
	Values   []uint64 `json:"values"`
}

// TransferBatchMultiRecipient MUST emit when tokens are transferred, including zero value
// transfers as well as minting or burning.
// The operator argument MUST be msg.sender.
// The from argument MUST be the address of the holder whose balance is decreased.
// The to argument MUST be the list of the addresses of the recipients whose balance is increased.
// The ids argument MUST be the list of tokens being transferred.
// The values argument MUST be the list of number of tokens (matching the list
// and order of tokens specified in _ids) the holder balance is decreased by
// and match what the recipient balance is increased by.
// When minting/creating tokens, the from argument MUST be set to `0x0` (i.e. zero address).
// When burning/destroying tokens, the to argument MUST be set to `0x0` (i.e. zero address).
type TransferBatchMultiRecipient struct {
	Operator string   `json:"operator"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	IDs      []uint64 `json:"ids"`
	Values   []uint64 `json:"values"`
}

// ApprovalForAll MUST emit when approval for a second party/operator address
// to manage all tokens for an owner address is enabled or disabled
// (absence of an event assumes disabled).
type ApprovalForAll struct {
	Owner    string `json:"owner"`
	Operator string `json:"operator"`
	Approved bool   `json:"approved"`
}

// URI MUST emit when the URI is updated for a token ID.
// Note: This event is not used in this contract implementation because in this implementation,
// only the programmatic way of setting URI is used. The URI should contain {id} as part of it
// and the clients MUST replace this with the actual token ID.
type URI struct {
	Value string `json:"value"`
	ID    uint64 `json:"id"`
}

// To represents recipient address
// ID represents token ID
type ToID struct {
	To string
	ID uint64
}

// Mint creates amount tokens of token type id and assigns them to account.
// This function emits a TransferSingle event.
// 특정 주소에 새로운 토큰을 발행하고, 해당 이벤트를 TransferSingle로 기록
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, tokenId uint64, categoryCode string,
	pollingResultId string, tokenType string, totalTicket uint64, amount uint64, owner string) error {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	// 체인코드 권한 확인
	err := authorizationHelper(ctx)
	if err != nil {
		return err
	}

	// Get ID of submitting client identity
	// 클라이언트 ID 확인
	operator, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	// Mint tokens
	// 토큰 발행
	err = mintHelper(ctx, operator, owner, tokenId, amount)
	if err != nil {
		return err
	}

	// Emit TransferSingle event
	// 이 이벤트는 트랜잭션의 성공적인 토큰 전송을 나타내며, 해당 토큰의 발행, 소유 및 양에 대한 정보를 포함
	transferSingleEvent := TransferSingle{operator, "0x0", owner, tokenId, amount}
	return emitTransferSingle(ctx, transferSingleEvent)
}

// QueryTokensByOwner retrieves the tokens owned by a specific owner.
// 특정 소유자의 토큰을 조회하는 기능
func (s *SmartContract) QueryTokensByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*TransferSingle, error) {
	queryString := fmt.Sprintf(`{"selector":{"to":"%s"}}`, owner)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute a query on the world state: %v", err)
	}
	defer resultsIterator.Close()

	var transferEvents []*TransferSingle

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get query results: %v", err)
		}

		var transferEvent TransferSingle
		err = json.Unmarshal(queryResponse.Value, &transferEvent)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal TransferSingle event: %v", err)
		}

		transferEvents = append(transferEvents, &transferEvent)
	}

	return transferEvents, nil
}

// Set information for a token and intialize contract.
// param {String} name The name of the token
// param {String} symbol The symbol of the token
// 토큰의 이름과 심볼을 설정하여 체인코드를 초기화하는 기능
func (s *SmartContract) Initialize(ctx contractapi.TransactionContextInterface, name string, symbol string) (bool, error) {

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to intitialize contract
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return false, fmt.Errorf("failed to get MSPID: %v", err)
	}
	if clientMSPID != minterMSPID {
		return false, fmt.Errorf("client is not authorized to initialize contract")
	}

	// Check contract options are not already set, client is not authorized to change them once intitialized
	bytes, err := ctx.GetStub().GetState(nameKey)
	if err != nil {
		return false, fmt.Errorf("failed to get Name: %v", err)
	}
	if bytes != nil {
		return false, fmt.Errorf("contract options are already set, client is not authorized to change them")
	}

	err = ctx.GetStub().PutState(nameKey, []byte(name))
	if err != nil {
		return false, fmt.Errorf("failed to set token name: %v", err)
	}

	err = ctx.GetStub().PutState(symbolKey, []byte(symbol))
	if err != nil {
		return false, fmt.Errorf("failed to set symbol: %v", err)
	}
	return true, nil
}

// Helper Functions

// authorizationHelper checks minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
// 호출자가 발행 권한을 가지고 있는 확인하는 헬퍼 함수
func authorizationHelper(ctx contractapi.TransactionContextInterface) error {

	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}
	if clientMSPID != minterMSPID {
		return fmt.Errorf("client is not authorized to mint new tokens")
	}
	return nil
}

func mintHelper(ctx contractapi.TransactionContextInterface, operator string, owner string, tokenId uint64, amount uint64) error {
	if owner == "0x0" {
		return fmt.Errorf("mint to the zero address")
	}

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	err := addBalance(ctx, operator, owner, tokenId, amount)
	if err != nil {
		return err
	}
	return nil
}

func addBalance(ctx contractapi.TransactionContextInterface, sender string, recipient string, tokenId uint64, amount uint64) error {
	// Convert tokenId to string
	tokenIdString := strconv.FormatUint(uint64(tokenId), 10)

	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{recipient, tokenIdString, sender})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", balancePrefix, err)
	}

	balanceBytes, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return fmt.Errorf("failed to read account %s from world state: %v", recipient, err)
	}

	var balance uint64 = 0
	if balanceBytes != nil {
		balance, _ = strconv.ParseUint(string(balanceBytes), 10, 64)
	}

	balance += amount

	err = ctx.GetStub().PutState(balanceKey, []byte(strconv.FormatUint(uint64(balance), 10)))
	if err != nil {
		return err
	}
	return nil
}

func emitTransferSingle(ctx contractapi.TransactionContextInterface, transferSingleEvent TransferSingle) error {
	transferSingleEventJSON, err := json.Marshal(transferSingleEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}

	err = ctx.GetStub().SetEvent("TransferSingle", transferSingleEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}
	return nil
}
