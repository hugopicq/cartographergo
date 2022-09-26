package dcerpc

import (
	"encoding/binary"

	"github.com/hugopicq/cartographergo/utils"
)

type MSRPCBind struct {
	MaxTFrag   uint16
	MaxRFrag   uint16
	AssocGroup uint64
	Items      []CtxItem
}

type CtxItem struct {
	AbstractSyntax []byte
	TransferSyntax []byte
	ContextID      uint16
	TransItems     uint8
}

type MSRPCHeader struct {
	Type    uint8
	PduData []byte
	CallId  uint32
	Flags   uint8
}

type MSRPCRequestHeader struct {
	MSRPCHeader
	AllocHint uint32
	CtxId     uint16
	Opnum     uint16
}

type MSRPCResponseHeader struct {
	MSRPCHeader
	AllocHint   uint32
	CtxId       uint16
	CancelCount uint8
	padding     uint8
}

func NewMSRPCResponseHeader(data []byte) *MSRPCResponseHeader {
	header := new(MSRPCResponseHeader)
	header.Type = data[2]
	header.Flags = data[3]
	header.CallId = binary.LittleEndian.Uint32(data[12:16])
	header.AllocHint = binary.LittleEndian.Uint32(data[16:20])
	header.CtxId = binary.LittleEndian.Uint16(data[20:22])
	header.CancelCount = data[22]
	return header
}

func NewMSRPCBind() MSRPCBind {
	return MSRPCBind{
		MaxTFrag:   4280,
		MaxRFrag:   4280,
		AssocGroup: 0,
		Items:      []CtxItem{},
	}
}

func (bind *MSRPCBind) AddItem(item CtxItem) {
	bind.Items = append(bind.Items, item)
}

func (bind MSRPCBind) GetData() []byte {
	packet := utils.NewByteBuffer()
	packet.WriteU16(bind.MaxTFrag)
	packet.WriteU16(bind.MaxRFrag)
	packet.WriteU32(0)
	packet.WriteU8(uint8(len(bind.Items))) //Ctx num
	packet.WriteU8(0)
	packet.WriteU16(0)
	for _, item := range bind.Items {
		packet.WriteBytes(item.GetData())
	}
	return packet.GetData()
}

func (item CtxItem) GetData() []byte {
	packet := utils.NewByteBuffer()
	packet.WriteU16(item.ContextID)
	packet.WriteU8(item.TransItems)
	packet.WriteU8(0) //Pad
	packet.WriteBytes(item.AbstractSyntax)
	packet.WriteBytes(item.TransferSyntax)
	return packet.GetData()
}

func (header MSRPCHeader) GetData() []byte {
	//TODO : Support Auth
	packet := utils.NewByteBuffer()
	packet.WriteU8(5)                                 //Ver Major
	packet.WriteU8(0)                                 //Ver Minor
	packet.WriteU8(header.Type)                       //Type
	packet.WriteU8(header.Flags)                      //Flags
	packet.WriteU32(16)                               //Representation
	packet.WriteU16(uint16(16 + len(header.PduData))) //Total length TODO Change
	packet.WriteU16(0)                                //Auth length
	packet.WriteU32(header.CallId)                    //CallId
	packet.WriteBytes(header.PduData)
	return packet.GetData()
}
