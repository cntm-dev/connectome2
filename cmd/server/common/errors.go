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

const (
	CLIERR_OK                  = 0
	CLIERR_HTTP_METHOD_INVALID = 1001
	CLIERR_INVALID_REQUEST     = 1002
	CLIERR_INVALID_PARAMS      = 1003
	CLIERR_UNSUPPORT_METHOD    = 1004
	CLIERR_ACCOUNT_UNLOCK      = 1005
	CLIERR_INVALID_TX          = 1006
	CLIERR_INTERNAL_ERR        = 900
)

var RPCErrorDesc = map[int]string{
	CLIERR_OK:                  "",
	CLIERR_HTTP_METHOD_INVALID: "invalid http method",
	CLIERR_INVALID_REQUEST:     "invalid request",
	CLIERR_INVALID_PARAMS:      "invalid params",
	CLIERR_UNSUPPORT_METHOD:    "unsupport method",
	CLIERR_INVALID_TX:          "invalid tx",
	CLIERR_INTERNAL_ERR:        "internal error",
}

func GetCLIErrorDesc(errorCode int) string {
	desc, ok := RPCErrorDesc[errorCode]
	if !ok {
		return RPCErrorDesc[CLIERR_INTERNAL_ERR]
	}
	return desc
}
