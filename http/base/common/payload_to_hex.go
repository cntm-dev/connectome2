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
package common

import (
	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
)

type PayloadInfo interface{}

//implement PayloadInfo define BookKeepingInfo
type BookKeepingInfo struct {
	Nonce uint64
}

type InvokeCodeInfo struct {
	Code     string
	GasLimit uint64
}
type DeployCodeInfo struct {
	Code        string
	NeedStorage bool
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
}

type RecordInfo struct {
	RecordType string
	RecordData string
}

type BookkeeperInfo struct {
	PubKey     string
	Action     string
	Issuer     string
	Ccntmroller string
}

type DataFileInfo struct {
	IPFSPath string
	Filename string
	Note     string
	Issuer   string
}

type PrivacyPayloadInfo struct {
	PayloadType uint8
	Payload     string
	EncryptType uint8
	EncryptAttr string
}

type VoteInfo struct {
	PubKeys []string
	Voter   string
}

func TransPayloadToHex(p types.Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.Bookkeeper:
		obj := new(BookkeeperInfo)
		pubKeyBytes := keypair.SerializePublicKey(object.PubKey)
		obj.PubKey = common.ToHexString(pubKeyBytes)
		if object.Action == payload.BookkeeperAction_ADD {
			obj.Action = "add"
		} else if object.Action == payload.BookkeeperAction_SUB {
			obj.Action = "sub"
		} else {
			obj.Action = "nil"
		}
		pubKeyBytes = keypair.SerializePublicKey(object.Issuer)
		obj.Issuer = common.ToHexString(pubKeyBytes)

		return obj
	case *payload.InvokeCode:
		obj := new(InvokeCodeInfo)
		obj.Code = common.ToHexString(object.Code)
		return obj
	case *payload.DeployCode:
		obj := new(DeployCodeInfo)
		obj.Code = common.ToHexString(object.Code)
		obj.NeedStorage = object.NeedStorage
		obj.Name = object.Name
		obj.CodeVersion = object.Version
		obj.Author = object.Author
		obj.Email = object.Email
		obj.Description = object.Description
		return obj
	}
	return nil
}
