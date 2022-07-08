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

package global_params

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	scommon "github.com/cntmio/cntmology/core/store/common"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

type ParamCache struct {
	lock   sync.RWMutex
	Params Params
}

var GLOBAL_PARAM = map[string]string{}

type paramType byte

const (
	CURRENT_VALUE         paramType = 0x00
	PREPARE_VALUE         paramType = 0x01
	INIT_NAME                       = "init"
	ACCEPT_ADMIN_NAME               = "acceptAdmin"
	TRANSFER_ADMIN_NAME             = "transferAdmin"
	SET_GLOBAL_PARAM_NAME           = "setGlobalParam"
	GET_GLOBAL_PARAM_NAME           = "getGlobalParam"
	CREATE_SNAPSHOT_NAME            = "createSnapshot"
)

var paramCache *ParamCache
var admin *Admin

func InitGlobalParams() {
	native.Ccntmracts[utils.ParamCcntmractAddress] = RegisterParamCcntmract
	paramCache = new(ParamCache)
	paramCache.Params = make([]*Param, 0)
}

func RegisterParamCcntmract(native *native.NativeService) {
	native.Register(INIT_NAME, ParamInit)
	native.Register(ACCEPT_ADMIN_NAME, AcceptAdmin)
	native.Register(TRANSFER_ADMIN_NAME, TransferAdmin)
	native.Register(SET_GLOBAL_PARAM_NAME, SetGlobalParam)
	native.Register(GET_GLOBAL_PARAM_NAME, GetGlobalParam)
	native.Register(CREATE_SNAPSHOT_NAME, CreateSnapshot)
}

func ParamInit(native *native.NativeService) ([]byte, error) {
	paramCache = new(ParamCache)
	paramCache.Params = make([]*Param, 0)
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	initParams := new(Params)
	for k, v := range GLOBAL_PARAM {
		initParams.SetParam(&Param{k, v})
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(ccntmract, CURRENT_VALUE), getParamStorageItem(initParams))
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(ccntmract, PREPARE_VALUE), getParamStorageItem(initParams))
	admin = new(Admin)

	bookKeeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetBookkeepers error:%s", err)
	}
	initAddress := ctypes.AddressFromPubKey(bookKeeepers[0])
	copy((*admin)[:], initAddress[:])
	native.CloneCache.Add(scommon.ST_STORAGE, GetAdminKey(ccntmract, false), getAdminStorageItem(admin))
	return utils.BYTE_TRUE, nil
}

func AcceptAdmin(native *native.NativeService) ([]byte, error) {
	destinationAdmin := new(Admin)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("accept admin, deserialize admin failed!")
	}
	if !native.CcntmextRef.CheckWitness(common.Address(*destinationAdmin)) {
		return utils.BYTE_FALSE, errors.NewErr("accept admin, authentication failed!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	getAdmin(native, ccntmract)
	transferAdmin, err := GetStorageAdmin(native, GetAdminKey(ccntmract, true))
	if err != nil || transferAdmin == nil || *transferAdmin != *destinationAdmin {
		return utils.BYTE_FALSE, fmt.Errorf("accept admin, destination account hasn't been approved, casused by %v", err)
	}
	// delete transfer admin item
	native.CloneCache.Delete(scommon.ST_STORAGE, GetAdminKey(ccntmract, true))
	// modify admin in database
	native.CloneCache.Add(scommon.ST_STORAGE, GetAdminKey(ccntmract, false), getAdminStorageItem(destinationAdmin))

	admin = destinationAdmin
	return utils.BYTE_TRUE, nil
}

func TransferAdmin(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	getAdmin(native, ccntmract)
	if !native.CcntmextRef.CheckWitness(common.Address(*admin)) {
		return utils.BYTE_FALSE, errors.NewErr("transfer admin, authentication failed!")
	}
	destinationAdmin := new(Admin)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("transfer admin, deserialize admin failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, GetAdminKey(ccntmract, true),
		getAdminStorageItem(destinationAdmin))
	return utils.BYTE_TRUE, nil
}

func SetGlobalParam(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	getAdmin(native, ccntmract)
	if !native.CcntmextRef.CheckWitness(common.Address(*admin)) {
		return utils.BYTE_FALSE, errors.NewErr("set param, authentication failed!")
	}
	params := new(Params)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("set param, deserialize failed!")
	}
	// read old param from database
	storageParams, err := getStorageParam(native, getParamKey(ccntmract, PREPARE_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	// update param
	for _, param := range *params {
		storageParams.SetParam(param)
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(ccntmract, PREPARE_VALUE),
		getParamStorageItem(storageParams))
	return utils.BYTE_TRUE, nil
}

func GetGlobalParam(native *native.NativeService) ([]byte, error) {
	paramNameList := new(ParamNameList)
	if err := paramNameList.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("get param, deserialize failed!")
	}
	params := new(Params)
	var paramNotInCache = make([]string, 0)
	// read from cache
	for _, paramName := range *paramNameList {
		if index, value := getParamFromCache(paramName); index >= 0 {
			params.SetParam(value)
		} else {
			paramNotInCache = append(paramNotInCache, paramName)
		}
	}
	result := new(bytes.Buffer)
	if len(paramNotInCache) == 0 { // all request param exist in cache
		if err := params.Serialize(result); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, results seriealize error!")
		}
		return result.Bytes(), nil
	}
	// read from db
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	storageParams, err := getStorageParam(native, getParamKey(ccntmract, CURRENT_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, storage error!")
	}
	if len(*storageParams) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("get param, there are no params!")
	}
	setCache(storageParams)                     // set param to cache
	for _, paramName := range paramNotInCache { // read param not in cache
		if index, value := storageParams.GetParam(paramName); index >= 0 {
			params.SetParam(value)
		} else {
			return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("get param, param %v doesn't exist!", paramName))
		}
	}
	err = params.Serialize(result)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, results to json error!")
	}
	return result.Bytes(), nil
}

func CreateSnapshot(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	getAdmin(native, ccntmract)
	if !native.CcntmextRef.CheckWitness(common.Address(*admin)) {
		return utils.BYTE_FALSE, errors.NewErr("create snapshot, authentication failed!")
	}
	// read prepare param
	prepareParam, err := getStorageParam(native, getParamKey(ccntmract, PREPARE_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "create snapshot, storage error!")
	}
	if len(*prepareParam) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("create snapshot, prepare param doesn't exist!")
	}
	// set prepare value to current value, make it effective
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(ccntmract, CURRENT_VALUE), getParamStorageItem(prepareParam))
	// clear memory cache
	clearCache()
	return utils.BYTE_TRUE, nil
}

func getAdmin(native *native.NativeService, ccntmract common.Address) {
	if admin == nil || *admin == *new(Admin) {
		var err error
		// get admin from database
		admin, err = GetStorageAdmin(native, GetAdminKey(ccntmract, false))
		// there are no admin in database
		if err != nil {
			bookKeeepers, err := config.DefConfig.GetBookkeepers()
			if err != nil {
				log.Errorf("GetBookkeepers error: %v", err)
				return
			}
			initAddress := ctypes.AddressFromPubKey(bookKeeepers[0])
			copy((*admin)[:], initAddress[:])
		}
	}
}

func clearCache() {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params = make([]*Param, 0)
}

func setCache(params *Params) {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params = *params
}

func getParamFromCache(key string) (int, *Param) {
	paramCache.lock.RLock()
	defer paramCache.lock.RUnlock()
	return paramCache.Params.GetParam(key)
}
