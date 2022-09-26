package dcerpc

import (
	"encoding/binary"

	"github.com/hugopicq/cartographergo/utils"
)

type EptLookup struct {
	Opnum       uint
	InquiryType uint32
	Object      uint32
	Ifid        uint32
	VersOption  uint32
	EntryHandle []byte
	MaxEnts     uint32
}

func (call EptLookup) GetData() []byte {
	buffer := utils.NewByteBuffer()
	buffer.WriteU32(call.InquiryType)
	buffer.WriteU32(call.Object)
	buffer.WriteU32(call.Ifid)
	buffer.WriteU32(call.VersOption)
	// buffer.WriteU32(call.EntryHandle.HandleAttributes)
	// buffer.WriteBytes(call.EntryHandle.HandleUUID)
	buffer.WriteBytes(call.EntryHandle)
	buffer.WriteU32(call.MaxEnts)
	return buffer.GetData()
}

type EptLookupResponse struct {
	EntryHandle *EptLookupHandle
	NumEnts     uint32
	Entries     []*EptEntry
	Status      uint32
}

type EptEntry struct {
	Object     []byte //Should be 16
	Tower      *EptTower
	Annotation *NDRUniVaryingArray
}

type NDRUniVaryingArray struct {
	Offset      uint32
	ActualCount uint32
	Data        []byte
}

type EptTower struct {
	ReferentID       uint32
	TowerLength      uint32
	TowerOctetString []byte
}

func NewEptEntryFromBuffer(buffer []byte, offset uint) (*EptEntry, uint) {
	//We build the item
	//The structure fields are object, tower and annotation
	//Object are only Data of length 16
	//Tower is only length 4
	//Anotation has 3 fields (4 + 4 + 1)
	offset0 := offset
	aligment := 4
	offset += (uint(aligment) - (offset % uint(aligment))) % uint(aligment)

	entry := new(EptEntry)
	entry.Object = buffer[offset : offset+16]
	offset += 16
	entry.Tower = new(EptTower)
	entry.Tower.ReferentID = binary.LittleEndian.Uint32(buffer[offset : offset+4])
	offset += 4
	entry.Annotation = new(NDRUniVaryingArray)
	entry.Annotation.Offset = binary.LittleEndian.Uint32(buffer[offset : offset+4])
	offset += 4
	entry.Annotation.ActualCount = binary.LittleEndian.Uint32(buffer[offset : offset+4])
	offset += 4
	entry.Annotation.Data = buffer[offset : offset+uint(entry.Annotation.ActualCount)]
	offset += uint(entry.Annotation.ActualCount)

	return entry, offset - offset0
}

type EptLookupHandle struct {
	HandleAttributes uint32
	HandleUUID       []byte
}

func NewEptLookupResponse(data []byte) *EptLookupResponse {
	response := new(EptLookupResponse)
	response.EntryHandle = new(EptLookupHandle)
	response.EntryHandle.HandleAttributes = binary.LittleEndian.Uint32(data[0:4])
	response.EntryHandle.HandleUUID = data[4:20]

	response.NumEnts = binary.LittleEndian.Uint32(data[20:24])

	// maxElements := binary.LittleEndian.Uint32(data[24:28])
	// offset := binary.LittleEndian.Uint32(data[28:32])

	actualCount := binary.LittleEndian.Uint32(data[32:36])
	offset := 36
	numItems := actualCount
	//We are about to decode the array
	soFarItems := uint(0)
	array := []*EptEntry{}
	for {
		if numItems == 0 || soFarItems >= uint(len(data)-offset) {
			break
		}

		item, size := NewEptEntryFromBuffer(data, uint(soFarItems)+uint(offset))
		array = append(array, item)
		soFarItems += size
		numItems -= 1
	}

	offset += int(soFarItems)

	for _, item := range array {
		if item.Tower.ReferentID == 0 {
			continue
		}
		//We align
		offset += (4 - (offset % 4)) % 4
		//Then we get the ArraySize by unpacking 4 bytes little endian
		// arraySize := binary.LittleEndian.Uint32(data[offset:offset+4])
		offset += 4
		offset += (4 - (offset % 4)) % 4

		//Then we unpack tower_length with 4 bytes
		item.Tower.TowerLength = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4

		//Then we unpack the array of bytes
		item.Tower.TowerOctetString = data[offset : offset+int(item.Tower.TowerLength)]
		offset += int(item.Tower.TowerLength)
	}

	response.Entries = array
	response.Status = binary.LittleEndian.Uint32(data[offset : offset+4])

	return response
}
