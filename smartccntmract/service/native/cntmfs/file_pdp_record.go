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

package cntmfs

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

type PdpRecordKey struct {
	RecordKey []byte
}

type PdpRecord struct {
	NodeAddr    common.Address
	FileHash    []byte
	FileOwner   common.Address
	LastPdpTime uint64
	SettleFlag  bool
}

type PdpRecordList struct {
	PdpRecords []PdpRecord
}

func (this *PdpRecord) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.NodeAddr)
	sink.WriteVarBytes(this.FileHash)
	utils.EncodeAddress(sink, this.FileOwner)
	utils.EncodeVarUint(sink, this.LastPdpTime)
	sink.WriteBool(this.SettleFlag)
}

func (this *PdpRecord) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.FileHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.FileOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.LastPdpTime, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.SettleFlag, err = DecodeBool(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *PdpRecordList) Serialization(sink *common.ZeroCopySink) {
	pdpInfoCount := uint64(len(this.PdpRecords))
	utils.EncodeVarUint(sink, pdpInfoCount)
	if pdpInfoCount != 0 {
		for _, pdpInfo := range this.PdpRecords {
			sinkTmp := common.NewZeroCopySink(nil)
			pdpInfo.Serialization(sinkTmp)
			sink.WriteVarBytes(sinkTmp.Bytes())
		}
	}
}

func (this *PdpRecordList) Deserialization(source *common.ZeroCopySource) error {
	pdpInfoCount, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	if pdpInfoCount == 0 {
		return nil
	}

	for i := uint64(0); i < pdpInfoCount; i++ {
		pdpInfoTmp, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}

		var pdpInfo PdpRecord
		src := common.NewZeroCopySource(pdpInfoTmp)
		if err = pdpInfo.Deserialization(src); err != nil {
			return err
		}
		this.PdpRecords = append(this.PdpRecords, pdpInfo)
	}
	return nil
}

func addPdpRecord(native *native.NativeService, pdpRecord *PdpRecord) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	pdpRecordKey := GenFsPdpRecordKey(ccntmract, pdpRecord.FileHash, pdpRecord.FileOwner, pdpRecord.NodeAddr)

	sink := common.NewZeroCopySink(nil)
	pdpRecord.Serialization(sink)
	utils.PutBytes(native, pdpRecordKey, sink.Bytes())
}

func delPdpRecord(native *native.NativeService, fileHash []byte, fileOwner common.Address, nodeAddr common.Address) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	pdpRecordKey := GenFsPdpRecordKey(ccntmract, fileHash, fileOwner, nodeAddr)

	native.CacheDB.Delete(pdpRecordKey)
}

func pdpRecordExist(native *native.NativeService, fileHash []byte, fileOwner common.Address, nodeAddr common.Address) bool {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	pdpRecordKey := GenFsPdpRecordKey(ccntmract, fileHash, fileOwner, nodeAddr)

	item, err := utils.GetStorageItem(native, pdpRecordKey)
	if err != nil || item == nil || item.Value == nil {
		return false
	}
	return true
}

func getPdpRecord(native *native.NativeService, fileHash []byte, fileOwner common.Address, nodeAddr common.Address) *PdpRecord {
	pdpRawInfo := getPdpRawRecord(native, fileHash, fileOwner, nodeAddr)
	if pdpRawInfo == nil {
		return nil
	}
	var pdpInfo PdpRecord
	source := common.NewZeroCopySource(pdpRawInfo)
	if err := pdpInfo.Deserialization(source); err != nil {
		return nil
	}
	return &pdpInfo
}

func getPdpRawRecord(native *native.NativeService, fileHash []byte, fileOwner common.Address, nodeAddr common.Address) []byte {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	pdpRecordKey := GenFsPdpRecordKey(ccntmract, fileHash, fileOwner, nodeAddr)

	item, err := utils.GetStorageItem(native, pdpRecordKey)
	if err != nil || item == nil || item.Value == nil {
		return nil
	}
	return item.Value
}

func getPdpRecordList(native *native.NativeService, fileHash []byte, fileOwner common.Address) *PdpRecordList {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	pdpRecordPrefix := GenFsPdpRecordPrefix(ccntmract, fileHash, fileOwner)

	var pdpRecordList PdpRecordList
	iter := native.CacheDB.NewIterator(pdpRecordPrefix[:])
	for has := iter.First(); has; has = iter.Next() {
		item, err := utils.GetStorageItem(native, iter.Key())
		if err != nil || item == nil || item.Value == nil {
			log.Error("getPdpRecordList GetStorageItem ", err)
			ccntminue
		}

		var pdpRecord PdpRecord
		source := common.NewZeroCopySource(item.Value)
		if err := pdpRecord.Deserialization(source); err != nil {
			log.Errorf("getPdpRecordList Deserialization error: ", err.Error())
			ccntminue
		}
		pdpRecordList.PdpRecords = append(pdpRecordList.PdpRecords, pdpRecord)
	}
	iter.Release()

	return &pdpRecordList
}

func delPdpRecordList(native *native.NativeService, fileHash []byte, fileOwner common.Address) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	pdpRecordPrefix := GenFsPdpRecordPrefix(ccntmract, fileHash, fileOwner)

	var pdpRecordKeyList []PdpRecordKey
	iter := native.CacheDB.NewIterator(pdpRecordPrefix[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		pdpRecordKey := PdpRecordKey{
			RecordKey: make([]byte, len(key)),
		}
		copy(pdpRecordKey.RecordKey, key)
		pdpRecordKeyList = append(pdpRecordKeyList, pdpRecordKey)
	}
	iter.Release()
	for _, pdpRecordKey := range pdpRecordKeyList {
		native.CacheDB.Delete(pdpRecordKey.RecordKey)
	}
}
