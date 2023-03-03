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
package handlers

import (
	"testing"

	clisvrcom "github.com/conntectome/cntm/cmd/sigsvr/common"
)

func TestCreateAccount(t *testing.T) {
	walletStore := clisvrcom.DefWalletStore
	req := &clisvrcom.CliRpcRequest{
		Qid:    "t",
		Method: "createaccount",
		Pwd:    string(pwd),
	}
	resp := &clisvrcom.CliRpcResponse{}
	CreateAccount(req, resp)
	if resp.ErrorCode != 0 {
		t.Errorf("CreateAccount failed. ErrorCode:%d", resp.ErrorCode)
		return
	}
	createRsp, ok := resp.Result.(*CreateAccountRsp)
	if !ok {
		t.Errorf("CreateAccount resp asset to CreateAccountRsp failed")
		return
	}
	_, err := walletStore.GetAccountByAddress(createRsp.Account, pwd)
	if err != nil {
		t.Errorf("Test CreateAccount failed GetAccountByAddress error:%s", err)
		return
	}
}
