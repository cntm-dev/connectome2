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
	"fmt"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology-crypto/signature"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntmfs/pdp"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func FsFileProve(native *native.NativeService) ([]byte, error) {
	var pdpData PdpData
	source := common.NewZeroCopySource(native.Input)
	if err := pdpData.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve Deserialization error!")
	}
	if !native.CcntmextRef.CheckWitness(pdpData.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve CheckWitness failed!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve getGlobalParam error!")
	}

	fileInfo := getFileInfoByHash(native, pdpData.FileHash)
	if fileInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve getFileInfoByHash error!")
	}

	nodeInfo := getNodeInfo(native, pdpData.NodeAddr)
	if nodeInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve getNodeInfo error!")
	}

	pdpRecord := getPdpRecord(native, fileInfo.FileHash, fileInfo.FileOwner, pdpData.NodeAddr)
	if pdpRecord == nil {
		if fileInfo.FirstPdp {
			log.Info("[Node Business] FsFileProve FirstPdp is true, checkPdpData.")
			if pdpData.ChallengeHeight != fileInfo.BeginHeight {
				return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve pdpData ChallengeHeight error!")
			}
			if err = checkPdpData(native, &pdpData, fileInfo); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsFileProve checkPdpData(file) error: %s",
					err.Error())
			}
		} else {
			log.Info("[Node Business] FsFileProve FirstPdp is false, checkPdpData skip.")
		}

		pdpRecord = &PdpRecord{NodeAddr: pdpData.NodeAddr, FileHash: pdpData.FileHash,
			FileOwner: fileInfo.FileOwner, LastPdpTime: uint64(native.Time), SettleFlag: false}

		if nodeInfo.RestVol < fileInfo.FileBlockCount*DefaultPerBlockSize {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve space RestVol not enough error!")
		}
		nodeInfo.RestVol -= fileInfo.FileBlockCount * DefaultPerBlockSize

		if err = checkUint64OverflowWithSum(nodeInfo.Profit, globalParam.CcntmractInvokeGasFee); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsFileProve error: %s", err.Error())
		}
		nodeInfo.Profit += globalParam.CcntmractInvokeGasFee
		addNodeInfo(native, nodeInfo)
		addPdpRecord(native, pdpRecord)
		return utils.BYTE_TRUE, nil
	}
	if uint64(native.Time) < fileInfo.TimeExpired {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve file is not expired!")
	}
	if pdpRecord.SettleFlag {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve pdp finished!")
	}

	if pdpData.ChallengeHeight != fileInfo.ExpiredHeight {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve pdpData ChallengeHeight error!")
	}
	if err = checkPdpData(native, &pdpData, fileInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsFileProve checkPdpData(space) error: %s",
			err.Error())
	}

	var fileStoreProfit uint64
	switch fileInfo.StorageType {
	case FileStorageTypeUseFile:
		fileStoreProfit = calcFileModePerServerProfit(fileInfo.TimeExpired, fileInfo)
		if fileInfo.RestAmount < fileStoreProfit {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve file RestAmount not enough error!")
		}
		fileInfo.RestAmount -= fileStoreProfit
		addFileInfo(native, fileInfo)
	case FileStorageTypeUseSpace:
		spaceInfo := getSpaceInfoFromDb(native, fileInfo.FileOwner)
		if spaceInfo == nil {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve getSpaceInfoFromDb error!")
		}
		fileStoreProfit = calcSpaceModePerServerProfit(spaceInfo.TimeExpired, spaceInfo.TimeExpired, fileInfo)
		if spaceInfo.RestAmount < fileStoreProfit {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve space RestAmount not enough error!")
		}
		spaceInfo.RestAmount -= fileStoreProfit
		addSpaceInfo(native, spaceInfo)
	default:
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve file storage type error!")
	}

	if err = checkUint64OverflowWithSum(nodeInfo.Profit, fileStoreProfit); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsFileProve error: %s", err.Error())
	}
	nodeInfo.Profit += fileStoreProfit

	if err = checkUint64OverflowWithSum(nodeInfo.Profit, globalParam.CcntmractInvokeGasFee); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsFileProve error: %s", err.Error())
	}
	nodeInfo.Profit += globalParam.CcntmractInvokeGasFee

	fileSize := fileInfo.FileBlockCount * DefaultPerBlockSize
	if err = checkUint64OverflowWithSum(nodeInfo.RestVol, fileSize); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsFileProve error: %s", err.Error())
	}
	nodeInfo.RestVol += fileSize

	pdpRecord.SettleFlag = true
	pdpRecord.LastPdpTime = uint64(native.Time)

	recordList := getPdpRecordList(native, fileInfo.FileHash, fileInfo.FileOwner)
	if recordList == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsFileProve getPdpRecordList recordList is nil!")
	}
	var cleanFlag = true
	for _, pdpRecordTmp := range recordList.PdpRecords {
		if pdpRecordTmp.NodeAddr == pdpRecord.NodeAddr {
			ccntminue
		}
		if !pdpRecordTmp.SettleFlag {
			cleanFlag = false
		}
	}

	if cleanFlag && pdpRecord.SettleFlag {
		var errInfos Errors
		deleteFile(native, fileInfo, &errInfos)
		if len(errInfos.ObjectErrors) != 0 {
			errInfos.PrintErrors()
		}
	} else {
		addPdpRecord(native, pdpRecord)
	}

	addNodeInfo(native, nodeInfo)
	return utils.BYTE_TRUE, nil
}

func FsGetNodeChallengeList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	nodeAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[Node Business] FsGetNodeChallengeList DecodeAddress error!")), nil
	}

	challengeList := getNodeChallengeList(native, nodeAddr)
	if challengeList == nil {
		return EncRet(false, []byte("[Node Business] FsGetNodeChallengeList challengeList is nil!")), nil
	}

	sink := common.NewZeroCopySink(nil)
	challengeList.Serialization(sink)

	return EncRet(true, sink.Bytes()), nil
}

func FsResponse(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var pdpData PdpData
	source := common.NewZeroCopySource(native.Input)
	if err := pdpData.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse Deserialization error!")
	}
	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse getGlobalParam error!")
	}

	if !native.CcntmextRef.CheckWitness(pdpData.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse CheckProver failed!")
	}

	nodeInfo := getNodeInfo(native, pdpData.NodeAddr)
	if nodeInfo == nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsGetNodeInfoList getNodeInfo(%v) error", pdpData.NodeAddr)
	}

	challengeInfo := getChallenge(native, pdpData.NodeAddr, pdpData.FileHash)
	if challengeInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse getChallenge failed!")
	}

	if pdpData.ChallengeHeight != challengeInfo.ChallengeHeight {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse challenge height is error!")
	}

	switch challengeInfo.State {
	case NoReplyAndExpire:
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse challenge state is NoReplyAndExpire!")
	case RepliedAndSuccess:
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse challenge state is RepliedAndSuccess!")
	case RepliedButVerifyError:
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse challenge state is RepliedButVerifyError!")
	case Judged:
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse challenge state is Judged!")
	}

	fileInfo := getFileInfoByHash(native, pdpData.FileHash)
	if fileInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse getFileInfoByHash failed!")
	}

	if err := checkPdpData(native, &pdpData, fileInfo); err != nil {
		if err = checkUint64OverflowWithSum(fileInfo.PayAmount, globalParam.CcntmractInvokeGasFee); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsResponse error: %s", err.Error())
		}
		punishAmount := fileInfo.PayAmount + globalParam.CcntmractInvokeGasFee
		if nodeInfo.Profit > punishAmount {
			nodeInfo.Profit -= punishAmount
		} else if nodeInfo.Pledge > punishAmount {
			nodeInfo.Pledge -= punishAmount
		} else {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse profit or pledge not enough!")
		}
		if err = checkUint64OverflowWithSum(punishAmount, challengeInfo.Reward); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsResponse error: %s", err.Error())
		}
		err = appCallTransfer(native, utils.OngCcntmractAddress, ccntmract, challengeInfo.FileOwner,
			punishAmount+challengeInfo.Reward)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsResponse AppCallTransfer, transfer error!")
		}

		challengeInfo.Reward = punishAmount
		challengeInfo.State = RepliedButVerifyError
	} else {
		if err = checkUint64OverflowWithSum(nodeInfo.Profit, challengeInfo.Reward); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsResponse error: %s", err.Error())
		}
		nodeInfo.Profit += challengeInfo.Reward
		challengeInfo.Reward = 0
		challengeInfo.State = RepliedAndSuccess
	}

	addNodeInfo(native, nodeInfo)
	addChallenge(native, challengeInfo)
	return utils.BYTE_TRUE, nil
}

func checkPdpData(native *native.NativeService, pdpData *PdpData, fileInfo *FileInfo) error {
	blockHash := native.Store.GetBlockHash(uint32(pdpData.ChallengeHeight))
	hexBlockHash := blockHash.ToArray()

	log.Debugf("ChallengeHeight: %d, blockCount: %d, blockHash: %v\n", pdpData.ChallengeHeight,
		fileInfo.FileBlockCount, hexBlockHash)
	return CheckPdpProve(pdpData.NodeAddr, hexBlockHash, fileInfo.FileBlockCount, fileInfo.PdpParam, pdpData.ProveData)
}

//export this function for cntm-fs server
func CheckPdpProve(nodeAddr common.Address, blockHash []byte, fileBlockCount uint64, fileUniqueId []byte,
	proofData []byte) error {
	pdpVersion := pdp.GetPdpVersionFromProof(proofData)

	var pdpService = pdp.NewPdp(pdpVersion)
	challenge, err := pdpService.GenChallenge(nodeAddr, blockHash, fileBlockCount)
	if err != nil {
		return fmt.Errorf("[Node Business] GenChallenge error: %s", err.Error())
	}
	err = pdp.VerifyProofWithUniqueId(fileUniqueId, proofData, challenge)
	if err != nil {
		return fmt.Errorf("[Node Business] checkPdpData error: %s", err.Error())
	}
	return nil
}

func FsReadFileSettle(native *native.NativeService) ([]byte, error) {
	var settleSlice FileReadSettleSlice
	source := common.NewZeroCopySource(native.Input)
	if err := settleSlice.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle Deserialization error!")
	}

	if !native.CcntmextRef.CheckWitness(settleSlice.PayTo) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle Check Slice owner failed!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle getGlobalParam error!")
	}

	readPledge, err := getReadPledge(native, settleSlice.PayFrom, settleSlice.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle getReadPledge error!")
	}

	for i := 0; i < len(readPledge.ReadPlans); i++ {
		//search FsNode
		if readPledge.ReadPlans[i].NodeAddr != settleSlice.PayTo {
			ccntminue
		}
		if readPledge.ReadPlans[i].HaveReadBlockNum >= settleSlice.SliceId ||
			readPledge.ReadPlans[i].MaxReadBlockNum < settleSlice.SliceId {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle SliceId error!")
		}
		if readPledge.Downloader != settleSlice.PayFrom {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle Downloader error!")
		}

		ret, err := checkSettleSig(settleSlice)
		if err != nil || !ret {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle checkSettleSig failed!")
		}

		readFee := (settleSlice.SliceId - readPledge.ReadPlans[i].HaveReadBlockNum) * globalParam.FeePerBlockForRead
		if readPledge.RestMoney < readFee {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle RestMoney < readFee ")
		}

		readPledge.ReadPlans[i].HaveReadBlockNum = settleSlice.SliceId
		readPledge.RestMoney -= readFee

		nodeInfo := getNodeInfo(native, settleSlice.PayTo)
		if nodeInfo == nil {
			return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle getNodeInfo error!")
		}
		if readPledge.ReadPlans[i].NumOfSettlements == 0 {
			if err = checkUint64OverflowWithSum(readFee, globalParam.CcntmractInvokeGasFee); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsReadFileSettle error: %s", err.Error())
			}
			if err = checkUint64OverflowWithSum(readFee, readFee+globalParam.CcntmractInvokeGasFee); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsReadFileSettle error: %s", err.Error())
			}
			nodeInfo.Profit += readFee + globalParam.CcntmractInvokeGasFee
			readPledge.ReadPlans[i].NumOfSettlements++
		} else {
			if err = checkUint64OverflowWithSum(nodeInfo.Profit, readFee); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("[Node Business] FsReadFileSettle error: %s", err.Error())
			}
			nodeInfo.Profit += readFee
		}

		addNodeInfo(native, nodeInfo)
		addReadPledge(native, readPledge)

		return utils.BYTE_TRUE, nil
	}
	return utils.BYTE_FALSE, errors.NewErr("[Node Business] FsReadFileSettle settleSlice PayTo error!")
}

func checkSettleSig(settleSlice FileReadSettleSlice) (bool, error) {
	settleSliceTmp := FileReadSettleSlice{
		FileHash:     settleSlice.FileHash,
		PayFrom:      settleSlice.PayFrom,
		PayTo:        settleSlice.PayTo,
		SliceId:      settleSlice.SliceId,
		PledgeHeight: settleSlice.PledgeHeight,
	}

	sink := common.NewZeroCopySink(nil)
	settleSliceTmp.Serialization(sink)

	pubKey, err := keypair.DeserializePublicKey(settleSlice.PubKey)
	if err != nil {
		return false, fmt.Errorf("checkSettleSig DeserializePublicKey error: %s", err.Error())
	}
	addr := types.AddressFromPubKey(pubKey)
	if addr != settleSlice.PayFrom {
		return false, fmt.Errorf("checkSettleSig Pubkey not match walletAddr ")
	}
	signValue, err := signature.Deserialize(settleSlice.Sig)
	if err != nil {
		return false, fmt.Errorf("checkSettleSig signature Deserialize error: %s", err.Error())
	}

	result := signature.Verify(pubKey, sink.Bytes(), signValue)
	return result, nil
}
