/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	common4 "github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/types"
	common3 "github.com/cntmio/cntmology/wasmtest/common"
)

const WingABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

var testEthAddr common4.Address
var testPrivateKey *ecdsa.PrivateKey

func getBridgeCcntmract(ccntmract []Item) (bridgeWasm, bridge, wingErc20, wingOep4 common3.ConAddr) {
	for _, item := range ccntmract {
		file := item.File
		code := item.Ccntmract
		if strings.HasSuffix(file, "wing_eth.evm") {
			ethAddr := crypto.CreateAddress(testEthAddr, 0)
			addr, _ := common.AddressParseFromBytes(ethAddr.Bytes())
			wingErc20 = common3.ConAddr{
				File:    file,
				Address: addr,
			}
			log.Infof("wingErc20 token address: %s", wingErc20.Address.ToHexString())
		} else if strings.HasSuffix(file, "bridge.avm") {
			bridge = common3.ConAddr{
				File:    file,
				Address: common.AddressFromVmCode(code),
			}
			log.Infof("bridge address: %s", bridge.Address.ToHexString())
		} else if strings.HasSuffix(file, "WingToken.avm") {
			wingOep4 = common3.ConAddr{
				File:    file,
				Address: common.AddressFromVmCode(code),
			}
			log.Infof("wingOep4 address: %s", wingOep4.Address.ToHexString())
		} else if strings.HasSuffix(file, "bridge_optimized.wasm") {
			bridgeWasm = common3.ConAddr{
				File:    file,
				Address: common.AddressFromVmCode(code),
			}
			log.Infof("bridge wasm address: %s", bridgeWasm.Address.ToHexString())
		} else {
			ccntminue
		}
	}
	return
}

func deployCcntmract(ccntmract []Item, acct *account.Account, database *ledger.Ledger) {
	txes := make([]*types.Transaction, 0, len(ccntmract))
	nonce := int64(0)
	for _, item := range ccntmract {
		file := item.File
		ccntm := item.Ccntmract
		var tx *types.Transaction
		var err error
		if strings.HasSuffix(file, ".wasm") {
			if file == "bridge_optimized2.wasm" {
				ccntminue
			}
			tx, err = NewDeployWasmCcntmract(acct, ccntm)
		} else if strings.HasSuffix(file, ".avm") {
			if file == "bridge2.avm" {
				// migrate ccntmract
				ccntminue
			}
			tx, err = NewDeployNeoCcntmract(acct, ccntm)
		} else if strings.HasSuffix(file, "wing_eth.evm") {
			chainId := big.NewInt(int64(config.DefConfig.P2PNode.EVMChainId))
			testPrivateKeyStr := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
			testPrivateKey, err = crypto.HexToECDSA(testPrivateKeyStr)
			checkErr(err)
			testEthAddr = crypto.PubkeyToAddress(testPrivateKey.PublicKey)
			opts, err := bind.NewKeyedTransactorWithChainID(testPrivateKey, chainId)
			opts.GasPrice = big.NewInt(0)
			opts.Nonce = big.NewInt(nonce)
			opts.GasLimit = 8000000
			checkErr(err)
			ethtx, err := NewDeployEvmCcntmract(opts, ccntm, WingABI)
			checkErr(err)
			tx, err = types.TransactionFromEIP155(ethtx)
			checkErr(err)
			_, err = tx.GetEIP155Tx()
			checkErr(err)
			nonce++
		}
		checkErr(err)
		_, err = database.PreExecuteCcntmract(tx)
		//log.Infof("deploy %s consume gas: %d, %s", file, res.Gas, JsonString(res))
		checkErr(err)
		txes = append(txes, tx)
	}
	block, err := makeBlock(acct, txes)
	checkErr(err)
	err = database.AddBlock(block, nil, common.UINT256_EMPTY)
	checkErr(err)
}

func migrateBridge(isWasm bool, bridge common3.ConAddr, newCode []byte, admin common.Address, database *ledger.Ledger, acct *account.Account) common3.ConAddr {
	te := common3.TestEnv{Witness: []common.Address{admin, acct.Address}}
	newCodeHex := hex.EncodeToString(newCode)
	//name, version, author, email, description
	var param string
	if isWasm {
		param = fmt.Sprintf("bytearray:%s,int:3,string:%s,string:%s,string:%s,string:%s,string:%s",
			newCodeHex, "bridge", "1.0", "cntmology", "@cntm.io", "desc")
	} else {
		param = fmt.Sprintf("[bytearray:%s,string:%s,string:%s,string:%s,string:%s,string:%s]",
			newCodeHex, "bridge", "1.0", "cntmology", "@cntm.io", "desc")
	}

	tc := common3.NewTestCase(te, false, "migrate", param, "bool:true", "")
	var tx *types.Transaction
	var err error
	if isWasm {
		tx, err = common3.GenWasmTransaction(tc, bridge.Address, &common3.TestCcntmext{})
	} else {
		tx, err = common3.GenNeoVMTransaction(tc, bridge.Address, &common3.TestCcntmext{})
	}
	checkErr(err)
	execTxCheckRes(tx, tc, database, bridge.Address, acct)
	newAddr := common.AddressFromVmCode(newCode)
	return common3.ConAddr{
		Address: newAddr,
	}
}

func bridgeTest(bridgeWasm, bridge, wingOep4, wingErc20 common3.ConAddr, database *ledger.Ledger, acct *account.Account) {
	testPrivateKeyStr := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	testPrivateKey, err := crypto.HexToECDSA(testPrivateKeyStr)
	checkErr(err)
	testEthAddr = crypto.PubkeyToAddress(testPrivateKey.PublicKey)

	// 调用 bridge init 方法
	admin, _ := common.AddressFromBase58("ARGK44mXXZfU6vcdSfFKMzjaabWxyog1qb")
	ccntmractInit(admin, bridgeWasm, bridge, wingOep4, wingErc20, database, acct)
	txNonce := int64(3)
	txNonce = bridgeTestInner(admin, bridgeWasm, bridge, wingOep4, wingErc20, database, acct, txNonce)
	oep4BalanceBefore := oep4BalanceOf(database, wingOep4, bridge.Address)
	erc20BalanceBefore := erc20BalanceOf(database, wingErc20, cntmAddrToEthAddr(bridge.Address), txNonce)

	newCode := loadCcntmract(ccntmractDir2 + "/" + "bridge2.avm")
	newBridge := migrateBridge(false, bridge, newCode, admin, database, acct)

	newCodeBridge := loadCcntmract(ccntmractDir2 + "/" + "bridge_optimized2.wasm")
	newBridgeWasm := migrateBridge(true, bridgeWasm, newCodeBridge, admin, database, acct)

	log.Infof("newBridge: %s, newBridgeWasm: %s", newBridge.Address.ToHexString(), newBridgeWasm.Address.ToHexString())
	oep4BalanceAfter := oep4BalanceOf(database, wingOep4, newBridge.Address)
	erc20BalanceAfter := erc20BalanceOf(database, wingErc20, cntmAddrToEthAddr(newBridge.Address), txNonce)
	ensureTrue(oep4BalanceBefore, oep4BalanceAfter)
	ensureTrue(erc20BalanceBefore, erc20BalanceAfter)
	bridgeTestInner(admin, newBridgeWasm, newBridge, wingOep4, wingErc20, database, acct, txNonce)
}

func bridgeTestInner(admin common.Address, bridgeWasm, bridge, wingOep4, wingErc20 common3.ConAddr, database *ledger.Ledger, acct *account.Account, txNonce int64) int64 {
	rand.Seed(time.Now().UnixNano())
	ethAcct, _ := common.AddressParseFromBytes(testEthAddr.Bytes())
	for i := 0; i < 100; i++ {
		log.Infof("bridgeTestInner i: %d", i)
		amount := uint64(rand.Int63n(int64(100)))
		if amount == 0 {
			ccntminue
		}
		// oep4 to erc20
		beforeAdmin := oep4BalanceOf(database, wingOep4, admin)
		beforeEthAcct := erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		oep4ToErc20(bridge, admin, ethAcct, amount, database, acct, "oep4ToErc20")
		afterAdmin := oep4BalanceOf(database, wingOep4, admin)
		ensureTrue(afterAdmin, beforeAdmin-amount)
		afterEthAcct := erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		ensureTrue(afterEthAcct, beforeEthAcct+amount)

		// bridgewasm
		beforeAdmin = oep4BalanceOf(database, wingOep4, admin)
		beforeEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		oep4ToErc20Wasm(bridgeWasm, admin, ethAcct, amount, database, acct, "oep4ToErc20")
		afterAdmin = oep4BalanceOf(database, wingOep4, admin)
		ensureTrue(afterAdmin, beforeAdmin-amount)
		afterEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		ensureTrue(afterEthAcct, beforeEthAcct+amount)

		// erc20 to oep4
		erc20Approve(database, wingErc20, cntmAddrToEthAddr(bridge.Address), amount, acct, txNonce)
		txNonce++
		beforeEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		beforeAdmin = oep4BalanceOf(database, wingOep4, admin)
		oep4ToErc20(bridge, admin, ethAcct, amount, database, acct, "erc20ToOep4")
		afterEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		ensureTrue(afterEthAcct, beforeEthAcct-amount)
		afterAdmin = oep4BalanceOf(database, wingOep4, admin)
		ensureTrue(afterAdmin, beforeAdmin+amount)

		erc20Approve(database, wingErc20, cntmAddrToEthAddr(bridgeWasm.Address), amount, acct, txNonce)
		txNonce++
		beforeEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		beforeAdmin = oep4BalanceOf(database, wingOep4, admin)
		log.Infof("amount: %d", amount)
		oep4ToErc20Wasm(bridgeWasm, admin, ethAcct, amount, database, acct, "erc20ToOep4")
		afterEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		ensureTrue(afterEthAcct, beforeEthAcct-amount)
		afterAdmin = oep4BalanceOf(database, wingOep4, admin)
		ensureTrue(afterAdmin, beforeAdmin+amount)
	}
	return txNonce
}

func oep4ToErc20Wasm(bridgeWasm common3.ConAddr, admin common.Address, ethAcct common.Address, amount uint64, database *ledger.Ledger, acct *account.Account, method string) {
	oep4ToErc20Inner(true, bridgeWasm, admin, ethAcct, amount, database, acct, method)
}

func oep4ToErc20(bridge common3.ConAddr, admin common.Address, ethAcct common.Address, amount uint64, database *ledger.Ledger, acct *account.Account, method string) {
	oep4ToErc20Inner(false, bridge, admin, ethAcct, amount, database, acct, method)
}

func oep4ToErc20Inner(isWasm bool, bridge common3.ConAddr, admin common.Address, ethAcct common.Address, amount uint64, database *ledger.Ledger, acct *account.Account, method string) {
	var param string
	if method == "oep4ToErc20" {
		if isWasm {
			param = fmt.Sprintf("address:%s,address:%s,int:%d", admin.ToBase58(), ethAcct.ToBase58(), amount)
		} else {
			param = fmt.Sprintf("[address:%s,address:%s,int:%d]", admin.ToBase58(), ethAcct.ToBase58(), amount)
		}
	} else if method == "erc20ToOep4" {
		if isWasm {
			param = fmt.Sprintf("address:%s,address:%s,int:%d", ethAcct.ToBase58(), admin.ToBase58(), amount)
		} else {
			param = fmt.Sprintf("[address:%s,address:%s,int:%d]", ethAcct.ToBase58(), admin.ToBase58(), amount)
		}
	} else {
		panic(method)
	}

	testCcntmext := common3.TestCcntmext{
		Admin:   admin,
		AddrMap: nil,
	}
	env := common3.TestEnv{
		Witness: []common.Address{admin, ethAcct},
	}
	tc := common3.NewTestCase(env, false, method, param, "bool:true", WingABI)
	var tx *types.Transaction
	var err error
	if isWasm {
		tx, err = common3.GenWasmTransaction(tc, bridge.Address, &testCcntmext)
	} else {
		tx, err = common3.GenNeoVMTransaction(tc, bridge.Address, &testCcntmext)
	}
	checkErr(err)
	log.Infof("method: %s, isWasm: %v", method, isWasm)
	_, err = database.PreExecuteCcntmract(tx)
	checkErr(err)
	//log.Infof("method: %s, pre: %s", method, JsonString(reee))
	execTxCheckRes(tx, tc, database, bridge.Address, acct)
}

func ccntmractInit(admin common.Address, bridgeWasm, bridge, wingOep4, wingErc20 common3.ConAddr, database *ledger.Ledger, acct *account.Account) {
	//wing oep4 init
	param := "int:1"
	testCcntmext := common3.TestCcntmext{
		Admin:   admin,
		AddrMap: nil,
	}
	te := common3.TestEnv{Witness: []common.Address{admin, acct.Address}}
	tc := common3.NewTestCase(te, false, "init", param, "bool:true", "")
	tx, err := common3.GenNeoVMTransaction(tc, wingOep4.Address, &testCcntmext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, wingOep4.Address, acct)

	// oep4 balanceOf
	ba := oep4BalanceOf(database, wingOep4, admin)
	ensureTrue(1000000000000000, ba)
	amount := uint64(100000000000000)
	oep4Transfer(database, wingOep4, admin, bridge.Address, amount, testCcntmext, acct)
	ba = oep4BalanceOf(database, wingOep4, bridge.Address)
	ensureTrue(amount, ba)

	oep4Transfer(database, wingOep4, admin, bridgeWasm.Address, amount, testCcntmext, acct)
	ba = oep4BalanceOf(database, wingOep4, bridgeWasm.Address)
	ensureTrue(amount, ba)

	log.Infof("wingOep4: %s", wingOep4.Address.ToHexString())
	param = fmt.Sprintf("address:%s,address:%s", wingOep4.Address.ToBase58(), wingErc20.Address.ToBase58())
	// bridge wasm init
	tc = common3.NewTestCase(te, false, "init", param, "bool:true", "")
	tx, err = common3.GenWasmTransaction(tc, bridgeWasm.Address, &testCcntmext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, bridgeWasm.Address, acct)

	// bridge wasm get_cntm_address
	tc = common3.NewTestCase(te, false, "get_oep4_address", "int:1", "address:"+wingOep4.Address.ToBase58(), "")
	tx, err = common3.GenWasmTransaction(tc, bridgeWasm.Address, &testCcntmext)
	checkErr(err)
	resTemp, err := database.PreExecuteCcntmract(tx)
	log.Infof("res:%v", resTemp.Result)
	execTxCheckRes(tx, tc, database, bridgeWasm.Address, acct)

	// bridge init
	param = fmt.Sprintf("[address:%s,address:%s]", wingOep4.Address.ToBase58(), wingErc20.Address.ToBase58())
	tc = common3.NewTestCase(te, false, "init", param, "bool:true", "")
	tx, err = common3.GenNeoVMTransaction(tc, bridge.Address, &testCcntmext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, bridge.Address, acct)

	// bridge get_cntm_address
	tc = common3.NewTestCase(te, false, "get_oep4_address", "int:1", "address:"+wingOep4.Address.ToBase58(), "")
	tx, err = common3.GenNeoVMTransaction(tc, bridge.Address, &testCcntmext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, bridge.Address, acct)

	// wing erc20 totalSupply
	evmTx, err := GenEVMTx(1, common4.BytesToAddress(wingErc20.Address[:]), "totalSupply", "")
	checkErr(err)
	tx, err = types.TransactionFromEIP155(evmTx)
	checkErr(err)
	res, err := database.PreExecuteCcntmract(tx)
	checkErr(err)
	r := res.Result.([]byte)
	log.Infof("execute totalSupply: %v", JsonString(r))

	// wing erc20 name
	evmTx, err = GenEVMTx(1, common4.BytesToAddress(wingErc20.Address[:]), "name")
	checkErr(err)
	tx, err = types.TransactionFromEIP155(evmTx)
	checkErr(err)
	res, err = database.PreExecuteCcntmract(tx)
	checkErr(err)
	parseEthResult("name", res.Result, WingABI)

	// wingErc20 balanceOf
	ba = erc20BalanceOf(database, wingErc20, testEthAddr, 1)
	ensureTrue(500000000000000, ba)
	erc20Transfer(database, wingErc20, cntmAddrToEthAddr(bridge.Address), amount, acct, 1)
	ba = erc20BalanceOf(database, wingErc20, cntmAddrToEthAddr(bridge.Address), 2)
	ensureTrue(amount, ba)

	erc20Transfer(database, wingErc20, cntmAddrToEthAddr(bridgeWasm.Address), amount, acct, 2)
	ba = erc20BalanceOf(database, wingErc20, cntmAddrToEthAddr(bridgeWasm.Address), 3)
	ensureTrue(amount, ba)
}

func ensureTrue(expect, actual uint64) {
	if expect != actual {
		panic(fmt.Sprintf("expect: %d, actual: %d", expect, actual))
	}
}

func cntmAddrToEthAddr(cntmAddr common.Address) common4.Address {
	return common4.BytesToAddress(cntmAddr[:])
}

func erc20Transfer(database *ledger.Ledger, wingErc20 common3.ConAddr, to common4.Address, amount uint64, acct *account.Account, txNonce int64) {
	evmTx, err := GenEVMTx(txNonce, common4.BytesToAddress(wingErc20.Address[:]), "transfer", to, big.NewInt(0).SetUint64(amount))
	checkErr(err)
	tx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	tc := common3.NewTestCase(common3.TestEnv{}, false, "transfer", "", "bool:true", WingABI)
	execTxCheckRes(tx, tc, database, wingErc20.Address, acct)
}

func erc20Approve(database *ledger.Ledger, wingErc20 common3.ConAddr, to common4.Address, amount uint64, acct *account.Account, txNonce int64) {
	evmTx, err := GenEVMTx(txNonce, common4.BytesToAddress(wingErc20.Address[:]), "approve", to, big.NewInt(0).SetUint64(amount))
	checkErr(err)
	tx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	tc := common3.NewTestCase(common3.TestEnv{}, false, "approve", "", "bool:true", WingABI)
	execTxCheckRes(tx, tc, database, wingErc20.Address, acct)

	evmTx, err = GenEVMTx(txNonce, common4.BytesToAddress(wingErc20.Address[:]), "allowance", testEthAddr, to)
	checkErr(err)
	tx, err = types.TransactionFromEIP155(evmTx)
	checkErr(err)
	res, err := database.PreExecuteCcntmract(tx)
	checkErr(err)
	log.Infof("allowance: %v", res.Result)
}

func erc20BalanceOf(database *ledger.Ledger, wingErc20 common3.ConAddr, addr common4.Address, txNonce int64) uint64 {
	evmTx, err := GenEVMTx(txNonce, common4.BytesToAddress(wingErc20.Address[:]), "balanceOf", addr)
	checkErr(err)
	tx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	res, err := database.PreExecuteCcntmract(tx)
	checkErr(err)
	data := parseEthResult("balanceOf", res.Result, WingABI)
	d, ok := data.(*big.Int)
	if !ok {
		panic(data)
	}
	return d.Uint64()
}

func oep4BalanceOf(database *ledger.Ledger, wingOep4 common3.ConAddr, addr common.Address) uint64 {
	tc := common3.NewTestCase(common3.TestEnv{}, false, "balanceOf", fmt.Sprintf("[address:%s]", addr.ToBase58()), "", "")
	tx, err := common3.GenNeoVMTransaction(tc, wingOep4.Address, &common3.TestCcntmext{})
	checkErr(err)
	res, err := database.PreExecuteCcntmract(tx)
	checkErr(err)
	data := res.Result.(string)
	d, _ := hex.DecodeString(data)
	return common.BigIntFromNeoBytes(d).Uint64()
}

func oep4Transfer(database *ledger.Ledger, wingOep4 common3.ConAddr, from, to common.Address, amount uint64, testCcntmext common3.TestCcntmext, acct *account.Account) {
	tc := common3.NewTestCase(common3.TestEnv{}, false, "transfer", fmt.Sprintf("[address:%s,address:%s,int:%d]", from.ToBase58(), to.ToBase58(), amount), "bool:true", "")
	tx, err := common3.GenNeoVMTransaction(tc, wingOep4.Address, &testCcntmext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, wingOep4.Address, acct)
}

func parseEthResult(method string, data interface{}, jsonAbi string) interface{} {
	r := data.([]byte)
	parsed, _ := abi.JSON(strings.NewReader(jsonAbi))
	eee, err := parsed.Unpack(method, r)
	checkErr(err)
	log.Infof("method: %s, result: %v", method, eee)
	return eee[0]
}

func GenEVMTx(nonce int64, ccntmractAddr common4.Address, method string, params ...interface{}) (*types2.Transaction, error) {
	chainId := big.NewInt(int64(config.DefConfig.P2PNode.EVMChainId))
	opts, err := bind.NewKeyedTransactorWithChainID(testPrivateKey, chainId)
	opts.GasPrice = big.NewInt(0)
	opts.Nonce = big.NewInt(nonce)
	opts.GasLimit = 8000000

	checkErr(err)
	parsed, err := abi.JSON(strings.NewReader(WingABI))
	checkErr(err)
	input, err := parsed.Pack(method, params...)
	deployTx := types2.NewTransaction(opts.Nonce.Uint64(), ccntmractAddr, opts.Value, opts.GasLimit, opts.GasPrice, input)
	signedTx, err := opts.Signer(opts.From, deployTx)
	checkErr(err)
	return signedTx, err
}
