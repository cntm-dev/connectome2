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

// DataEntryPrefix
type DataEntryPrefix byte

const (
	// DATA
	DATA_BLOCK       DataEntryPrefix = 0x00
	DATA_HEADER                      = 0x01
	DATA_TRANSACTION                 = 0x02

	// Transaction
	ST_BOOKKEEPER DataEntryPrefix = 0x03
	ST_CcntmRACT   DataEntryPrefix = 0x04
	ST_STORAGE    DataEntryPrefix = 0x05
	ST_VALIDATOR  DataEntryPrefix = 0x07
	ST_VOTE       DataEntryPrefix = 0x08

	IX_HEADER_HASH_LIST DataEntryPrefix = 0x09

	//SYSTEM
	SYS_CURRENT_BLOCK      DataEntryPrefix = 0x10
	SYS_VERSION            DataEntryPrefix = 0x11
	SYS_CURRENT_STATE_ROOT DataEntryPrefix = 0x12
	SYS_BLOCK_MERKLE_TREE  DataEntryPrefix = 0x13

	EVENT_NOTIFY DataEntryPrefix = 0x14
)
