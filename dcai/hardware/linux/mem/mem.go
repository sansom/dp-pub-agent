package mem

import (
	"bytes"
	"fmt"
	"github.com/influxdata/telegraf/dcai/util"
	"regexp"
)

var (
	execcmd               = util.ExecuteSudoCmdWithTimeout
	checkCmdPermission    = util.CheckCmdRootPermission
	memDeviceStartRegexp  = regexp.MustCompile("^(Memory Device)\\s*$")
	memSizeRegexp         = regexp.MustCompile("^\\s*Size\\s*:\\s*(\\w.*\\w) MB\\s*$")
	memBankRegexp         = regexp.MustCompile("^\\s*Locator\\s*:\\s*(\\w.*\\w)\\s*$")
	memTypeRegexp         = regexp.MustCompile("^\\s*Type\\s*:\\s*(\\w.*\\w)\\s*$")
	memTypeDetailRegexp   = regexp.MustCompile("^\\s*Type Detail\\s*:\\s*(\\w.*\\w)\\s*$")
	memSpeedRegexp        = regexp.MustCompile("^\\s*Speed\\s*:\\s*(\\w.*\\w)\\s*$")
	memManufacturerRegexp = regexp.MustCompile("^\\s*Manufacturer\\s*:\\s*(\\w.*\\w)\\s*$")
	memSerialNumberRegexp = regexp.MustCompile("^\\s*Serial Number\\s*:\\s*(\\w.*\\w)\\s*$")
	memAssetTagRegexp     = regexp.MustCompile("^\\s*Asset Tag\\s*:\\s*(\\w.*\\w)\\s*$")
	memPartNumberRegexp   = regexp.MustCompile("^\\s*Part Number\\s*:\\s*(\\w.*\\w)\\s*$")
)

type MemInfo struct {
	Manufacturer string
	SerialNumber string
	AssetTag     string
	PartNumber   string
	Type         string
	TypeDetail   string
	Speed        string
	Bank         string
	SizeMB       string
}

func NewAllMemInfo(dmidecodePath string) ([]*MemInfo, error) {
	var (
		mems         []*MemInfo
		mem          *MemInfo
		memstart     int
		nextmemstart int
		out          []byte
		memtxt       []byte
		err          error
	)

	if _, err := checkCmdPermission("dmidecode"); err != nil {
		return nil, err
	}
	if out, err = execcmd(dmidecodePath, "-t memory"); err != nil {
		return nil, err
	}

	memstart = bytes.Index(out, []byte("Memory Device"))
	if memstart == -1 {
		return nil, fmt.Errorf("Cannot find any memory device")
	}
	for nextmemstart != -1 {
		nextmemstart = bytes.Index(out[memstart+1:], []byte("Memory Device"))
		if nextmemstart != -1 {
			memtxt = out[memstart : memstart+1+nextmemstart]
		} else {
			memtxt = out[memstart:]
		}
		memstart = memstart + 1 + nextmemstart
		mem = new(MemInfo)
		m := []*util.FindRegexpMatchAndSetType{
			{&mem.Manufacturer, memManufacturerRegexp},
			{&mem.SerialNumber, memSerialNumberRegexp},
			{&mem.AssetTag, memAssetTagRegexp},
			{&mem.PartNumber, memPartNumberRegexp},
			{&mem.Type, memTypeRegexp},
			{&mem.TypeDetail, memTypeDetailRegexp},
			{&mem.Speed, memSpeedRegexp},
			{&mem.Bank, memBankRegexp},
			{&mem.SizeMB, memSizeRegexp},
		}

		util.FindRegexpMatchAndSet(string(memtxt), m)
		mems = append(mems, mem)
	}
	return mems, nil

}
