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

package neovm

import (
	"math/big"

	scommon "github.com/cntmio/cntmology/core/store/common"
	vmtype "github.com/cntmio/cntmology/vm/neovm/types"
	"github.com/cntmio/cntmology/core/store"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/states"
)

const (
	MAX_STACK_SIZE = 2 * 1024
	MAX_ARRAY_SIZE = 1024
	MAX_SIZE_FOR_BIGINTEGER = 32
)

var (
	ServiceMap = map[string]Service{
		"Neo.Attribute.GetUsage": {Execute: AttributeGetUsage, Validator: validatorAttribute},
		"Neo.Attribute.GetData": {Execute: AttributeGetData, Validator: validatorAttribute},
		"Neo.Block.GetTransactionCount": {Execute: BlockGetTransactionCount, Validator: validatorBlock},
		"Neo.Block.GetTransactions": {Execute: BlockGetTransactions, Validator: validatorBlock},
		"Neo.Block.GetTransaction": {Execute: BlockGetTransaction, Validator: validatorBlockTransaction},
		"Neo.Blockchain.GetHeight": {Execute: BlockChainGetHeight},
		"Neo.Blockchain.GetHeader": {Execute: BlockChainGetHeader},
		"Neo.Blockchain.GetBlock": {Execute: BlockChainGetBlock},
		"Neo.Blockchain.GetTransaction": {Execute: BlockChainGetTransaction},
		"Neo.Blockchain.GetCcntmract": {Execute: BlockChainGetCcntmract},
		"Neo.Header.GetIndex": {Execute: HeaderGetIndex},
		"Neo.Header.GetHash": {Execute: HeaderGetHash},
		"Neo.Header.GetVersion": {Execute: HeaderGetVersion},
		"Neo.Header.GetPrevHash": {Execute: HeaderGetVersion},
		"Neo.Header.GetTimestamp": {Execute: HeaderGetTimestamp},
		"Neo.Header.GetConsensusData": {Execute: HeaderGetConsensusData},
		"Neo.Header.GetNextConsensus": {Execute: HeaderGetNextConsensus},
		"Neo.Transaction.GetHash": {Execute: TransactionGetHash},
		"Neo.Transaction.GetType": {Execute: TransactionGetType},
		"Neo.Transaction.GetAttributes": {Execute: TransactionGetAttributes},
		"Neo.Ccntmract.Create": {Execute: CcntmractCreate},
		"Neo.Ccntmract.Migrate": {Execute: CcntmractMigrate},
		"Neo.Ccntmract.GetStorageCcntmext": {Execute: CcntmractGetStorageCcntmext},
		"Neo.Ccntmract.Destroy": {Execute: CcntmractDestory},
		"Neo.Ccntmract.GetScript": {Execute: CcntmractGetCode},
		"Neo.Runtime.GetTime": {Execute: RuntimeGetTime},
		"Neo.Runtime.CheckWitness": {Execute: RuntimeCheckWitness},
		"Neo.Runtime.Notify": {Execute: RuntimeNotify},
		"Neo.Runtime.Log": {Execute: RuntimeLog},
		"Neo.Storage.Get": {Execute: StorageGet},
		"Neo.Storage.Put": {Execute: StoragePut},
		"Neo.Storage.Delete": {Execute: StorageDelete},
		"Neo.Storage.GetCcntmext": {Execute: StorageGetCcntmext},
	}
)

var (
	ERR_CHECK_STACK_SIZE = errors.NewErr("[NeoVmService] vm over max stack size!")
	ERR_CHECK_ARRAY_SIZE = errors.NewErr("[NeoVmService] vm over max array size!")
	ERR_CHECK_BIGINTEGER = errors.NewErr("[NeoVmService] vm over max biginteger size!")
	ERR_CURRENT_CcntmEXT_NIL = errors.NewErr("[NeoVmService] neovm service current ccntmext doesn't exist!")
	ERR_EXECUTE_CODE = errors.NewErr("[NeoVmService] vm execute code invalid!")
)

type (
	Execute func(service *NeoVmService, engine *vm.ExecutionEngine) error
	Validator func(engine *vm.ExecutionEngine) error
)

type Service struct {
	Execute   Execute
	Validator Validator
}

type NeoVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	CcntmextRef    ccntmext.CcntmextRef
	Notifications []*event.NotifyEventInfo
	Tx            *types.Transaction
	Time          uint32
}

func NewNeoVmService(store store.LedgerStore, dbCache scommon.StateStore, tx *types.Transaction, time uint32, ctxRef ccntmext.CcntmextRef) *NeoVmService {
	var service NeoVmService
	service.Store = store
	service.CloneCache = storage.NewCloneCache(dbCache)
	service.Time = time
	service.Tx = tx
	service.CcntmextRef = ctxRef
	return &service
}

func (this *NeoVmService) Invoke() error {
	engine := vm.NewExecutionEngine()
	ctx := this.CcntmextRef.CurrentCcntmext()
	if ctx == nil {
		return ERR_CURRENT_CcntmEXT_NIL
	}
	if len(ctx.Code.Code) == 0 {
		return ERR_EXECUTE_CODE
	}
	engine.PushCcntmext(vm.NewExecutionCcntmext(engine, ctx.Code.Code))
	for {
		if len(engine.Ccntmexts) == 0 || engine.Ccntmext == nil {
			break
		}
		if engine.Ccntmext.GetInstructionPointer() >= len(engine.Ccntmext.Code) {
			break
		}
		if err := engine.ExecuteCode(); err != nil {
			return err
		}
		if engine.Ccntmext.GetInstructionPointer() < len(engine.Ccntmext.Code) {
			if ok := checkStackSize(engine); !ok {
				return ERR_CHECK_STACK_SIZE
			}
			if ok := checkArraySize(engine); !ok {
				return ERR_CHECK_ARRAY_SIZE
			}
			if ok := checkBigIntegers(engine); !ok {
				return ERR_CHECK_BIGINTEGER
			}
		}
		switch engine.OpCode {
		case vm.SYSCALL:
			if err := this.SystemCall(engine); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL:
			c := new(states.Ccntmract)
			if err := c.Deserialize(engine.Ccntmext.OpReader.Reader()); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] get ccntmract parameters error!")
			}
			if err := this.CcntmextRef.AppCall(c.Address, c.Method, []byte{}, c.Args, true); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service app call error!")
			}
		default:
			if err := engine.StepInto(); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execute error!")
			}
		}
	}
	this.CcntmextRef.PushNotifications(this.Notifications)
	this.CloneCache.Commit()
	return nil
}

func (this *NeoVmService) SystemCall(engine *vm.ExecutionEngine) error {
	serviceName := engine.Ccntmext.OpReader.ReadVarString()
	service, ok := ServiceMap[serviceName]
	if !ok {
		return errors.NewErr("[SystemCall] service not support!")
	}
	if service.Validator != nil {
		if err := service.Validator(engine); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service validator error!")
		}
	}

	if err := service.Execute(this, engine); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service execute error!")
	}
	return nil
}

func checkStackSize(engine *vm.ExecutionEngine) bool {
	size := 0
	if engine.OpCode < vm.PUSH16 {
		size = 1
	} else {
		switch engine.OpCode{
		case vm.DEPTH, vm.DUP, vm.OVER, vm.TUCK:
			size = 1
		case vm.UNPACK:
			if engine.EvaluationStack.Count() == 0 {
				return false
			}
			size = len(vm.PeekStackItem(engine).GetArray())
		}
	}
	size += engine.EvaluationStack.Count() + engine.AltStack.Count()
	if uint32(size) > MAX_STACK_SIZE {
		return false
	}
	return true
}

func checkArraySize(engine *vm.ExecutionEngine) bool {
	switch engine.OpCode {
	case vm.PACK:
	case vm.NEWARRAY:
	case vm.NEWSTRUCT:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		size := vm.PeekInt(engine)
		if size > MAX_ARRAY_SIZE {
			return false
		}
	}
	return true
}

func checkBigIntegers(engine *vm.ExecutionEngine) bool {
	switch engine.OpCode {
	case vm.INC:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		x := vm.PeekBigInteger(engine)
		if !checkBigInteger(x) || !checkBigInteger(new(big.Int).Add(x, big.NewInt(1))) {
			return false
		}
	case vm.DEC:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		x := vm.PeekBigInteger(engine)
		if !checkBigInteger(x) || (x.Sign() < 0 && !checkBigInteger(new(big.Int).Sub(x, big.NewInt(1)))) {
			return false
		}
	case vm.ADD:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x1) || !checkBigInteger(x2) || !checkBigInteger(new(big.Int).Add(x1, x2)) {
			return false
		}
	case vm.SUB:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x1) || !checkBigInteger(x2) || !checkBigInteger(new(big.Int).Sub(x1, x2)) {
			return false
		}
	case vm.MUL:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		lx2 := len(vmtype.ConvertBigIntegerToBytes(x2))
		lx1 := len(vmtype.ConvertBigIntegerToBytes(x1))
		if lx2 > MAX_SIZE_FOR_BIGINTEGER || lx1 > MAX_SIZE_FOR_BIGINTEGER || (lx1 + lx2) > MAX_SIZE_FOR_BIGINTEGER {
			return false
		}
	case vm.DIV:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x2) || !checkBigInteger(x1) {
			return false
		}
		if x2.Sign() == 0 {
			return false
		}
	case vm.MOD:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x2) || !checkBigInteger(x1) {
			return false
		}
		if x2.Sign() == 0 {
			return false
		}
	}
	return true
}

func checkBigInteger(value *big.Int) bool {
	if value == nil {
		return false
	}
	if len(vmtype.ConvertBigIntegerToBytes(value)) > MAX_SIZE_FOR_BIGINTEGER {
		return false
	}
	return true
}