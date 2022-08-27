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

package ledgerstore

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/store/statestore"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	ninit "github.com/cntmio/cntmology/smartccntmract/service/native/init"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"math/big"
)

//HandleDeployTransaction deal with smart ccntmract deploy transaction
func (self *StateStore) HandleDeployTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch,
	tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	deploy := tx.Payload.(*payload.DeployCode)
	txHash := tx.Hash()
	address := types.AddressFromVmCode(deploy.Code)
	var (
		notifies []*event.NotifyEventInfo
		err      error
	)

	if tx.GasPrice != 0 {
		// init smart ccntmract configuration info
		config := &smartccntmract.Config{
			Time:   block.Header.Timestamp,
			Height: block.Header.Height,
			Tx:     tx,
		}
		cache := storage.NewCloneCache(stateBatch)
		gasLimit := neovm.GAS_TABLE[neovm.CcntmRACT_CREATE_NAME] + calcGasByCodeLen(len(deploy.Code), neovm.GAS_TABLE[neovm.UINT_DEPLOY_CODE_LEN_NAME])
		balance, err := isBalanceSufficient(tx.Payer, cache, config, store, gasLimit*tx.GasPrice)
		if err != nil {
			if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
			return err
		}
		if tx.GasLimit < gasLimit {
			log.Errorf("gasLimit insufficient, need:%d actual:%d", gasLimit, tx.GasLimit)
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
		}
		notifies, err = costGas(tx.Payer, gasLimit*tx.GasPrice, config, cache, store)
		if err != nil {
			return err
		}
		cache.Commit()
	}

	log.Infof("deploy ccntmract address:%s", address.ToHexString())
	// store ccntmract message
	err = stateBatch.TryGetOrAdd(scommon.ST_CcntmRACT, address[:], deploy)
	if err != nil {
		return err
	}

	SaveNotify(eventStore, txHash, notifies, true)
	return nil
}

//HandleInvokeTransaction deal with smart ccntmract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch,
	tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	txHash := tx.Hash()
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || block.Header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	// init smart ccntmract configuration info
	config := &smartccntmract.Config{
		Time:   block.Header.Timestamp,
		Height: block.Header.Height,
		Tx:     tx,
	}

	var (
		codeLenGas uint64
		gasLimit   uint64
		gas        uint64
		balance    uint64
		err        error
	)
	cache := storage.NewCloneCache(stateBatch)
	if isCharge {
		codeLenGas = calcGasByCodeLen(len(invoke.Code), neovm.GAS_TABLE[neovm.UINT_INVOKE_CODE_LEN_NAME])
		balance, err := isBalanceSufficient(tx.Payer, cache, config, store, gasLimit*tx.GasPrice)
		if err != nil {
			if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
			return err
		}

		if tx.GasLimit < codeLenGas {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
			return fmt.Errorf("transaction gas: %d less than code length gas: %d", tx.GasLimit, codeLenGas)
		}
	}

	//init smart ccntmract info
	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        tx.GasLimit - codeLenGas,
	}

	//start the smart ccntmract executive function
	engine, _ := sc.NewExecuteEngine(invoke.Code)

	_, err = engine.Invoke()

	if isCharge {
		gasLimit = tx.GasLimit - sc.Gas
		gas = gasLimit * tx.GasPrice
		balance, err = getBalance(config, cache, store, tx.Payer)
		if err != nil {
			return err
		}
		if balance < gas {
			if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
		}
	}

	if err != nil {
		if isCharge {
			if err := costInvalidGas(tx.Payer, gas, config, stateBatch, store, eventStore, txHash); err != nil {
				return err
			}
		}
		return err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		mixGas := neovm.MIN_TRANSACTION_GAS
		if gasLimit < mixGas {
			if balance < mixGas*tx.GasPrice {
				if err := costInvalidGas(tx.Payer, balance, config, stateBatch, store, eventStore, txHash); err != nil {
					return err
				}
			}
			gas = mixGas * tx.GasPrice
		}
		notifies, err = costGas(tx.Payer, gas, config, sc.CloneCache, store)
		if err != nil {
			return err
		}

	}

	SaveNotify(eventStore, txHash, append(sc.Notifications, notifies...), true)
	sc.CloneCache.Commit()
	return nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notifies []*event.NotifyEventInfo, execSucc bool) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	var notifyInfo *event.ExecuteNotify
	if execSucc {
		notifyInfo = &event.ExecuteNotify{TxHash: txHash,
			State: event.CcntmRACT_STATE_SUCCESS, Notify: notifies}
	} else {
		notifyInfo = &event.ExecuteNotify{TxHash: txHash,
			State: event.CcntmRACT_STATE_FAIL, Notify: notifies}
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notifyInfo); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notifyInfo)
	return nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := cntm.Transfers{States: []*cntm.State{{From: from, To: to, Value: value}}}
	tr := new(bytes.Buffer)
	transfer.Serialize(tr)
	return tr.Bytes()
}

// check whether payer cntm balance sufficient
func isBalanceSufficient(payer common.Address, cache *storage.CloneCache, config *smartccntmract.Config, store store.LedgerStore, gas uint64) (uint64, error) {
	balance, err := getBalance(config, cache, store, payer)
	if err != nil {
		return 0, err
	}
	if balance < gas {
		return 0, fmt.Errorf("payer gas insufficient, need %d , only have %d", gas, balance)
	}
	return balance, nil
}

func costGas(payer common.Address, gas uint64, config *smartccntmract.Config,
	cache *storage.CloneCache, store store.LedgerStore) ([]*event.NotifyEventInfo, error) {

	params := genNativeTransferCode(payer, utils.GovernanceCcntmractAddress, gas)

	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	_, err := service.NativeCall(utils.OngCcntmractAddress, "transfer", params)
	if err != nil {
		return nil, err
	}
	return sc.Notifications, nil
}

func refreshGlobalParam(config *smartccntmract.Config, cache *storage.CloneCache, store store.LedgerStore) error {
	bf := new(bytes.Buffer)
	if err := utils.WriteVarUint(bf, uint64(len(neovm.GAS_TABLE_KEYS))); err != nil {
		return fmt.Errorf("write gas_table_keys length error:%s", err)
	}
	for _, value := range neovm.GAS_TABLE_KEYS {
		if err := serialization.WriteString(bf, value); err != nil {
			return fmt.Errorf("serialize param name error:%s", value)
		}
	}

	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamCcntmractAddress, "getGlobalParam", bf.Bytes())
	if err != nil {
		return err
	}
	params := new(global_params.Params)
	if err := params.Deserialize(bytes.NewBuffer(result.([]byte))); err != nil {
		return fmt.Errorf("deserialize global params error:%s", err)
	}

	for k, _ := range neovm.GAS_TABLE {
		n, ps := params.GetParam(k)
		if n != -1 && ps.Value != "" {
			pu, err := strconv.ParseUint(ps.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse uint %v", err)
			}
			neovm.GAS_TABLE[k] = pu
		}
	}
	return nil
}

func getBalance(config *smartccntmract.Config, cache *storage.CloneCache, store store.LedgerStore, address common.Address) (uint64, error) {
	bf := new(bytes.Buffer)
	if err := utils.WriteAddress(bf, address); err != nil {
		return 0, err
	}
	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.OngCcntmractAddress, cntm.BALANCEOF_NAME, bf.Bytes())
	if err != nil {
		return 0, err
	}
	return new(big.Int).SetBytes(result.([]byte)).Uint64(), nil
}

func costInvalidGas(address common.Address, gas uint64, config *smartccntmract.Config, stateBatch *statestore.StateBatch,
	store store.LedgerStore, eventStore scommon.EventStore, txHash common.Uint256) error {
	cache := storage.NewCloneCache(stateBatch)
	notifies, err := costGas(address, gas, config, cache, store)
	if err != nil {
		return err
	}
	cache.Commit()
	SaveNotify(eventStore, txHash, notifies, false)
	return nil
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}
