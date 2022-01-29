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

package test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/cntmio/cntmology/account"
	clicommon "github.com/cntmio/cntmology/cli/common"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/signature"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/core/utils"
	"github.com/cntmio/cntmology/http/base/rpc"
	"github.com/cntmio/cntmology/smartccntmract/service/native/states"
	sstates "github.com/cntmio/cntmology/smartccntmract/states"
	vmtypes "github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/urfave/cli"
	"encoding/binary"
	"bufio"
)

func signTransaction(signer *account.Account, tx *types.Transaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer, hash[:])
	tx.Sigs = append(tx.Sigs, &types.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func testAction(c *cli.Ccntmext) (err error) {
	txnNum := c.Int("num")
	passwd := c.String("password")
	genFile := c.Bool("gen")

	acct := account.Open(account.WALLET_FILENAME, []byte(passwd))
	acc, err := acct.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount error:", err)
		os.Exit(1)
	}
	if genFile {
		GenTransferFile(txnNum, acc, "transfer.txt")
		return nil
	}

	transferTest(txnNum, acc)

	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Tx2Hex(tx *types.Transaction) string {
	var buffer bytes.Buffer
	tx.Serialize(&buffer)
	return hex.EncodeToString(buffer.Bytes())
}

func GenTransferFile(n int, acc *account.Account, fileName string) {
	f, err := os.Create(fileName)
	check(err)
	w := bufio.NewWriter(f)

	defer func() {
		w.Flush()
		f.Close()
	}()

	for i := 0; i < n; i ++ {
		to := acc.Address
		binary.BigEndian.PutUint64(to[:], uint64(i))
		tx := NewOntTransferTransaction(acc.Address, to, 1)
		if err := signTransaction(acc, tx); err != nil {
			fmt.Println("signTransaction error:", err)
			os.Exit(1)
		}

		txhex := Tx2Hex(tx)
		_, _ = w.WriteString(fmt.Sprintf("%x,%s\n", tx.Hash(), txhex))
	}

}

func transferTest(n int, acc *account.Account) {	
	if n <= 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		tx := NewOntTransferTransaction(acc.Address, acc.Address, int64(i))
		if err := signTransaction(acc, tx); err != nil {
			fmt.Println("signTransaction error:", err)
			os.Exit(1)
		}

		txbf := new(bytes.Buffer)
		if err := tx.Serialize(txbf); err != nil {
			fmt.Println("Serialize transaction error.")
			os.Exit(1)
		}
		resp, err := rpc.Call(clicommon.RpcAddress(), "sendrawtransaction", 0,
			[]interface{}{hex.EncodeToString(txbf.Bytes())})

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		r := make(map[string]interface{})
		err = json.Unmarshal(resp, &r)
		if err != nil {
			fmt.Println("Unmarshal JSON failed")
			os.Exit(1)
		}
		switch r["result"].(type) {
		case map[string]interface{}:
		case string:
			fmt.Println(r["result"].(string))
		}
	}
}

func NewOntTransferTransaction(from, to common.Address, value int64) *types.Transaction {
	var sts []*states.State
	sts = append(sts, &states.State{
		From:  from,
		To:    to,
		Value: big.NewInt(value),
	})
	transfers := new(states.Transfers)
	transfers.States = sts

	bf := new(bytes.Buffer)

	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		os.Exit(1)
	}

	ccntm := &sstates.Ccntmract{
		Address: genesis.OntCcntmractAddress,
		Method:  "transfer",
		Args:    bf.Bytes(),
	}

	ff := new(bytes.Buffer)
	if err := ccntm.Serialize(ff); err != nil {
		fmt.Println("Serialize ccntmract struct error.")
		os.Exit(1)
	}

	tx := utils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code:   ff.Bytes(),
	})

	tx.Nonce = uint32(time.Now().Unix())

	return tx
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "test",
		Usage:       "run test routine",
		Description: "With nodectl test, you could run simple tests.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "num, n",
				Usage: "sample transaction numbers",
				Value: 1,
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
				Value: "passwordtest",
			},
			cli.BoolFlag{
				Name:  "gen, g",
				Usage: "gen transaction to file",

		},
		},
		Action: testAction,
		OnUsageError: func(c *cli.Ccntmext, err error, isSubcommand bool) error {
			clicommon.PrintError(c, err, "test")
			return cli.NewExitError("", 1)
		},
	}
}
