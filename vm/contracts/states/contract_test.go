/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */
package states

import (
	"testing"

	"github.com/conntectome/cntm/common"
)

func TestContract_Serialize_Deserialize(t *testing.T) {
	addr := common.AddressFromVmCode([]byte{1})

	c := &ContractInvokeParam{
		Version: 0,
		Address: addr,
		Method:  "init",
		Args:    []byte{2},
	}
	sink := common.NewZeroCopySink(nil)
	c.Serialization(sink)

	v := new(ContractInvokeParam)
	source := common.NewZeroCopySource(sink.Bytes())
	if err := v.Deserialization(source); err != nil {
		t.Fatalf("ContractInvokeParam deserialize error: %v", err)
	}
}
