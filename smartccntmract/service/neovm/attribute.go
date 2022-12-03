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
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
	vmtypes "github.com/cntmio/cntmology/vm/neovm/types"
)

// AttributeGetUsage put attribute's usage to vm stack
func AttributeGetUsage(service *NeoVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if a, ok := i.Data.(*types.TxAttribute); ok {
		return engine.EvalStack.PushInt64(int64(a.Usage))
	}
	return errors.NewErr("[AttributeGetUsage] Wrcntm type!")
}

// AttributeGetData put attribute's data to vm stack
func AttributeGetData(service *NeoVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if a, ok := i.Data.(*types.TxAttribute); ok {
		val, err := vmtypes.VmValueFromBytes(a.Data)
		if err != nil {
			return err
		}
		engine.EvalStack.Push(val)
		return nil
	}
	return errors.NewErr("[AttributeGetData] Wrcntm type!")
}
