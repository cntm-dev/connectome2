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

	"encoding/json"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/genesis"
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

var GLOBAL_PARAM = map[string]string{
	"init-key1": "init-value1",
	"init-key2": "init-value2",
	"init-key3": "init-value3",
	"init-key4": "init-value4",
}

type paramType byte

const (
	CURRENT_VALUE paramType = 0x00
	PREPARE_VALUE paramType = 0x01
)

var paramCache *ParamCache
var admin *Admin

func InitGlobalParams() {
	native.Ccntmracts[genesis.ParamCcntmractAddress] = RegisterParamCcntmract
	paramCache = new(ParamCache)
	paramCache.Params = make(map[string]string)
}

func RegisterParamCcntmract(native *native.NativeService) {
	native.Register("init", ParamInit)
	native.Register("acceptAdmin", AcceptAdmin)
	native.Register("transferAdmin", TransferAdmin)
	native.Register("setGlobalParam", SetGlobalParam)
	native.Register("getGlobalParam", GetGlobalParam)
	native.Register("createSnapshot", CreateSnapshot)
}

func ParamInit(native *native.NativeService) ([]byte, error) {
	paramCache = new(ParamCache)
	paramCache.Params = make(map[string]string)
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	initParams := new(Params)
	*initParams = make(map[string]string)
	for k, v := range GLOBAL_PARAM {
		(*initParams)[k] = v
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
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(ccntmract, false), getAdminStorageItem(admin))
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
	transferAdmin, err := getStorageAdmin(native, getAdminKey(ccntmract, true))
	if err != nil || *transferAdmin != *destinationAdmin {
		return utils.BYTE_FALSE, fmt.Errorf("accept admin, destination account hasn't been approved, casused by %v", err)
	}
	// delete transfer admin item
	native.CloneCache.Delete(scommon.ST_STORAGE, getAdminKey(ccntmract, true))
	// modify admin in database
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(ccntmract, false), getAdminStorageItem(destinationAdmin))

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
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(ccntmract, true),
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
	for key, value := range *params {
		(*storageParams)[key] = value
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(ccntmract, PREPARE_VALUE),
		getParamStorageItem(storageParams))
	notifyParamSetSuccess(native, ccntmract, *params)
	return utils.BYTE_TRUE, nil
}

func GetGlobalParam(native *native.NativeService) ([]byte, error) {
	paramNameList := new(ParamNameList)
	if err := paramNameList.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("get param, deserialize failed!")
	}
	params := new(Params)
	*params = make(map[string]string, 0)
	var paramNotInCache = make([]string, 0)
	// read from cache
	for _, paramName := range *paramNameList {
		if value := getParamFromCache(paramName); value != "" {
			(*params)[paramName] = value
		} else {
			paramNotInCache = append(paramNotInCache, paramName)
		}
	}
	if len(paramNotInCache) == 0 { // all request param exist in cache
		result, err := json.Marshal(params)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, results to json error!")
		}
		return result, nil
	}
	// read from db
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	storageParams, err := getStorageParam(native, getParamKey(ccntmract, CURRENT_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, storage error!")
	}
	if len(*storageParams) == 0 {
		return []byte{}, nil
	}
	setCache(storageParams)                     // set param to cache
	for _, paramName := range paramNotInCache { // read param not in cache
		if value, ok := (*storageParams)[paramName]; ok {
			(*params)[paramName] = value
		} else {
			return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("get param, param %v doesn't exist!", paramName))
		}
	}
	result, err := json.Marshal(params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, results to json error!")
	}
	return result, nil
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
		admin, err = getStorageAdmin(native, getAdminKey(ccntmract, false))
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
	paramCache.Params = make(map[string]string)
}

func setCache(params *Params) {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params = *params
}

func getParamFromCache(key string) string {
	paramCache.lock.RLock()
	defer paramCache.lock.RUnlock()
	return paramCache.Params[key]
}
