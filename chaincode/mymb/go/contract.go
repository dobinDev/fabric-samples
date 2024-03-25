package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
	"time"
)

type TokenERC1155Contract struct {
	contractapi.Contract
}

type Token1155 struct {
	TokenID          string    `json:"TokenID"`
	CategoryCode     uint64    `json:"CategoryCode"`
	PollingResultID  uint64    `json:"PollingResultID"`
	TokenType        string    `json:"TokenType"`
	TotalTicket      uint64    `json:"TotalTicket"`
	Amount           uint64    `json:"Amount"`
	TokenCreatedTime time.Time `json:"TokenCreatedTime"`
}

type User struct {
	NickName         string    `json:"NickName"`
	MymPoint         uint64    `json:"MymPoint"`
	OwnedToken       string    `json:"OwnedToken"`
	BlockCreatedTime time.Time `json:"BlockCreatedTime"`
}

type QueryResultToken struct {
	Key    string    `json:"Key"`
	Record Token1155 `json:"Record"`
}

type QueryResultUser struct {
	Key    string `json:"Key"`
	Record User   `json:"Record"`
}

const (
	tokenPrefix   = "token"
	balancePrefix = "balance"
)

func (c *TokenERC1155Contract) MintToken(ctx contractapi.TransactionContextInterface,
	tokenID string, categoryCode uint64, pollingResultID uint64, tokenType string,
	totalTicket uint64, amount uint64) (*Token1155, error) {

	// UUID 생성
	uuid := uuid.New()

	// TokenID 생성
	tokenID = fmt.Sprintf("0x%x", sha256.Sum256([]byte(uuid.String())))

	/*
		// 유니크한 데이터 생성
		uniqueData := fmt.Sprintf("%d%d%d%d%s", categoryCode, pollingResultID, totalTicket, amount, time.Now().String())

		// SHA256 해시 생성
		hash := sha256.New()
		hash.Write([]byte(uniqueData))
		hashBytes := hash.Sum(nil)

		// TokenID 생성 (0x를 앞에 붙여서 생성)
		tokenID = fmt.Sprintf("0x%x", hashBytes)
	*/

	// Token 생성
	token := Token1155{
		TokenID:          tokenID,
		CategoryCode:     categoryCode,
		PollingResultID:  pollingResultID,
		TokenType:        tokenType,
		TotalTicket:      totalTicket,
		Amount:           amount,
		TokenCreatedTime: time.Now(), // 현재 시간 사용
	}

	// TokenID, Token 저장
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

	// TokenID, Amount 저장
	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{tokenID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key for balance: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKey, []byte(fmt.Sprintf("%d", amount)))
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %v", err)
	}

	return &token, nil
}

func (c *TokenERC1155Contract) CreateUserBlock(ctx contractapi.TransactionContextInterface,
	nickname string, mymPoint uint64, ownedToken string) error {

	// User 생성
	user := User{
		NickName:         nickname,
		MymPoint:         mymPoint,
		OwnedToken:       ownedToken,
		BlockCreatedTime: time.Now(),
	}

	// User 블록 저장
	userKey := nickname // 닉네임 을 키로 사용
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marsshal usser block: %v", err)
	}

	err = ctx.GetStub().PutState(userKey, userBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for user block: %v", err)
	}
	return nil
}

func (c *TokenERC1155Contract) UpdateMymPoint(ctx contractapi.TransactionContextInterface, nickName string, delta uint64) error {

	// 기존 유저 정보 가져오기
	userKey := nickName
	userBytes, err := ctx.GetStub().GetState(userKey)
	if err != nil {
		return fmt.Errorf("failed to read user block: %v", err)
	}
	if userBytes == nil {
		return fmt.Errorf("user with nickname %s does not exist", nickName)
	}

	var user User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return fmt.Errorf("failed to unmarshal user block: %v", err)
	}

	// MymPoint 업데이트
	user.MymPoint += uint64(delta)

	// 업데이트된 유저 정보 저장
	userBytes, err = json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user block: %v", err)
	}

	err = ctx.GetStub().PutState(userKey, userBytes)
	if err != nil {
		return fmt.Errorf("failed to put state for updated user block: %v", err)
	}
	return nil
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

func (c *TokenERC1155Contract) GetAllTokens(ctx contractapi.TransactionContextInterface) ([]QueryResultToken, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(tokenPrefix, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get state by partial composite key: %v", err)
	}
	defer resultsIterator.Close()

	var results []QueryResultToken

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

		results = append(results, QueryResultToken{
			Key:    queryResponse.Key,
			Record: token,
		})
	}

	return results, nil
}

func (c *TokenERC1155Contract) GetUser(ctx contractapi.TransactionContextInterface, nickName string) (*User, error) {

	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{nickName})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %v", err)
	}

	if tokenBytes == nil {
		return nil, fmt.Errorf("token with ID %s does not exist", nickName)
	}

	var user User

	err = json.Unmarshal(tokenBytes, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}

	return &user, nil
}

func (c *TokenERC1155Contract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]QueryResultUser, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(tokenPrefix, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get state by partial composite key: %v", err)
	}
	defer resultsIterator.Close()

	var results []QueryResultUser

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next query response: %v", err)
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %v", err)
		}

		results = append(results, QueryResultUser{
			Key:    queryResponse.Key,
			Record: user,
		})
	}

	return results, nil
}

/*
1. TransferToken 함수는 contractapi.TransactionContextInterface 인터페이스를 받아서 스마트 계약의 트랜잭션 컨텍스트를 제공합니다.
이 함수는 송신자의 주소(from), 수신자의 주소(to), 전송할 토큰의 ID(tokenID), 그리고 전송할 토큰의 양(amount)을 매개변수로 받습니다.
2. 송신자 잔고 확인: 송신자의 잔고를 확인하기 위해 먼저 송신자의 주소와 토큰 ID를 사용하여 컴포지트 키를 생성합니다.
그 후, 해당 키를 사용하여 송신자의 잔고를 조회합니다. 만약 송신자의 잔고가 없으면 해당하는 오류 메시지를 반환합니다.
3. 받는 사람의 잔고 업데이트: 받는 사람의 주소와 토큰 ID를 사용하여 받는 사람의 잔고를 업데이트합니다.
먼저 받는 사람의 잔고를 조회하고, 만약 잔고가 존재한다면 이를 uint64로 변환하여 증가시킨 후, 다시 상태 데이터베이스에 저장합니다.
4. 송신자의 잔고 업데이트: 마지막으로, 송신자의 잔고를 감소시킵니다. 송신자의 잔고에서 전송된 양을 뺀 후,
이를 다시 상태 데이터베이스에 저장합니다.
5. 오류 처리: 각 단계에서 발생하는 오류는 적절한 오류 메시지와 함께 반환됩니다.
*/

func (c *TokenERC1155Contract) TransferToken(ctx contractapi.TransactionContextInterface,
	from string, to string, tokenID string, amount uint64) error {

	// 송신자 잔고 확인
	fromBalanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{from, tokenID})
	if err != nil {
		return fmt.Errorf("failed to create composite key for sender balance: %v", err)
	}

	fromBalanceBytes, err := ctx.GetStub().GetState(fromBalanceKey)
	if err != nil {
		return fmt.Errorf("failed to read sender balance: %v", err)
	}

	if fromBalanceBytes == nil {
		return fmt.Errorf("sender %s does not have balance for token %s", from, tokenID)
	}

	fromBalance, err := strconv.ParseUint(string(fromBalanceBytes), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert sender balance to uint: %v", err)
	}

	if fromBalance < amount {
		return fmt.Errorf("sender %s does not have enough balance for token %s", from, tokenID)
	}

	// 받는 사람의 잔고 업데이트
	toBalanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{to, tokenID})
	if err != nil {
		return fmt.Errorf("failed to create composite key for receiver balance: %v", err)
	}

	toBalanceBytes, err := ctx.GetStub().GetState(toBalanceKey)
	if err != nil {
		return fmt.Errorf("failed to read receiver balance: %v", err)
	}

	toBalance := uint64(0)
	if toBalanceBytes != nil {
		toBalance, err = strconv.ParseUint(string(toBalanceBytes), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert receiver balance to uint: %v", err)
		}
	}
	toBalance += amount

	// 송신자와 받는 사람의 잔고 업데이트를 한 번에 처리하기 위해 트랜잭션 내부에서 수행
	err = ctx.GetStub().PutState(fromBalanceKey, []byte(fmt.Sprintf("%d", fromBalance-amount)))
	if err != nil {
		return fmt.Errorf("failed to update sender balance: %v", err)
	}

	err = ctx.GetStub().PutState(toBalanceKey, []byte(fmt.Sprintf("%d", toBalance)))
	if err != nil {
		return fmt.Errorf("failed to update receiver balance: %v", err)
	}
	return nil
}

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
