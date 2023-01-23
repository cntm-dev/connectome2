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

const Hour = 3600

func calcFileModeRestAmount(timeNow uint64, fileInfo *FileInfo) uint64 {
	fTimeNow := formatUint64TimeToHour(timeNow)
	fExpired := formatUint64TimeToHour(fileInfo.TimeExpired)

	if fTimeNow >= fExpired {
		return 0
	}
	restMinute := (fExpired - fTimeNow) / Hour
	return restMinute * fileInfo.CopyNumber * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcFileModePerServerProfit(dataClosing uint64, fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToHour(fileInfo.TimeStart)
	fExpired := formatUint64TimeToHour(fileInfo.TimeExpired)
	dataClosing = formatUint64TimeToHour(dataClosing)

	if dataClosing <= fStart {
		return 0
	}
	if dataClosing >= fExpired {
		dataClosing = fExpired
	}
	intervalMinute := (dataClosing - fStart) / Hour
	return intervalMinute * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcSpaceModePerServerProfit(dataClosing uint64, spaceExpired uint64, fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToHour(fileInfo.TimeStart)
	sExpired := formatUint64TimeToHour(spaceExpired)
	dataClosing = formatUint64TimeToHour(dataClosing)

	if dataClosing <= fStart {
		return 0
	}
	if dataClosing < sExpired {
		dataClosing = sExpired
	}
	intervalMinute := (dataClosing - fStart) / Hour
	return intervalMinute * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcTotalPayAmountWithFile(fileInfo *FileInfo) uint64 {
	fStart := formatUint64TimeToHour(fileInfo.TimeStart)
	fExpired := formatUint64TimeToHour(fileInfo.TimeExpired)
	if fExpired <= fStart {
		return 0
	}
	intervalMinute := (fExpired - fStart) / Hour
	return intervalMinute * fileInfo.CopyNumber * fileInfo.FileBlockCount * fileInfo.CurrFeeRate
}

func calcTotalPayAmountWithSpace(spaceInfo *SpaceInfo) uint64 {
	sStart := formatUint64TimeToHour(spaceInfo.TimeStart)
	sExpired := formatUint64TimeToHour(spaceInfo.TimeExpired)
	if sExpired <= sStart {
		return 0
	}
	intervalMinute := (sExpired - sStart) / Hour
	return intervalMinute * spaceInfo.CopyNumber * (spaceInfo.Volume / 256) * spaceInfo.CurrFeeRate
}
