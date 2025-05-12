package main

import "encoding/binary"

type EndianType uint8

const (
	_ EndianType = iota
	LittleEndian
	BigEndian
)

var (
	EndianTypeFromStr = map[string]EndianType{
		"II": LittleEndian,
		"MM": BigEndian,
	}

	endianness_to_string = map[EndianType]string{
		LittleEndian: "LittleEndian",
		BigEndian:    "BigEndian",
	}

	endianness_byte_order = map[EndianType]binary.ByteOrder{
		LittleEndian: binary.LittleEndian,
		BigEndian:    binary.BigEndian,
	}
)

func (e *EndianType) ID() uint8 {
	if e == nil {
		return 0
	}
	return uint8(*e)
}

func (e *EndianType) String() string {
	if e == nil {
		return ""
	}
	return endianness_to_string[*e]
}

func (e *EndianType) ByteOrder() binary.ByteOrder {
	if e == nil {
		return nil
	}
	return endianness_byte_order[*e]
}
