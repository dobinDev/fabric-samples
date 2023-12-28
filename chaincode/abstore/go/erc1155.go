/*
	2021 Baran Kılıç <baran.kilic@boun.edu.tr>

	SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/hyperledger/fabric-samples/chaincode/abstore/go/chaincode"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	smartContract := new(chaincode.SmartContract)

	cc, err := contractapi.NewChaincode(smartContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
