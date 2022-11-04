package dcerpc

import (
	"errors"

	"github.com/hugopicq/cartographergo/utils"
)

const PFC_LAST_FRAG uint8 = 2

type DCE struct {
	CallId    uint32
	Transport *SMBTransport
}

func NewDCE(transport *SMBTransport) *DCE {
	return &DCE{Transport: transport, CallId: 1}
}

func (dce *DCE) Connect() error {
	return dce.Transport.Connect()
}

func (dce *DCE) Disconnect() {
	dce.Transport.Disconnect()
}

func (dce *DCE) Bind(ifaceUUID []byte) error {
	//Bind is authenticating
	bind := NewMSRPCBind()

	item := CtxItem{
		AbstractSyntax: ifaceUUID,
		TransferSyntax: []byte{0x04, 0x5d, 0x88, 0x8a, 0xeb, 0x1c, 0xc9, 0x11, 0x9f, 0xe8, 0x08, 0x00, 0x2b, 0x10, 0x48, 0x60, 0x02, 0x00, 0x00, 0x00},
		ContextID:      0,
		TransItems:     1,
	}

	bind.AddItem(item)

	packet := MSRPCHeader{Type: 11, CallId: dce.CallId, PduData: bind.GetData(), Flags: 1 | 2}
	err := dce.Transport.Send(packet.GetData())
	if err != nil {
		return err
	}

	_, err = dce.Transport.Receive()
	if err != nil {
		return err
	}

	return nil
}

func (dce *DCE) Request(request *EptLookup) (*EptLookupResponse, error) {
	call := new(DCERPCRawCall)
	call.Opnum = uint16(request.Opnum)
	call.PduData = request.GetData()
	dce.Send(call)
	resp, err := dce.Receive()
	if err != nil {
		return nil, errors.New("Error while receiving")
	}

	response, error := NewEptLookupResponse(resp)
	if error != nil {
		return nil, error
	}
	return response, nil
}

func (dce *DCE) Send(call *DCERPCRawCall) error {
	call.CtxId = 0
	call.CallId = dce.CallId
	call.AllocHint = uint32(len(call.PduData))

	//Should check fragment size

	//Then we transport send
	err := dce.Transport.Send(call.GetPacket())
	if err != nil {
		return err
	}
	dce.CallId += 1
	return nil
}

func (dce *DCE) Receive() ([]byte, error) {
	finished := false
	retAnswer := []byte{}
	for {
		if finished {
			break
		}

		response_data, err := dce.Transport.Receive()
		if err != nil {
			return retAnswer, err
		}

		header := NewMSRPCResponseHeader(response_data)
		if header.Flags&PFC_LAST_FRAG == PFC_LAST_FRAG {
			finished = true
		}

		answer := response_data[24:]
		retAnswer = append(retAnswer, answer...)
	}
	return retAnswer, nil
}

type DCERPCRawCall struct {
	MSRPCRequestHeader
}

func (call *DCERPCRawCall) GetPacket() []byte {
	packet := utils.NewByteBuffer()
	//Redundant code, should be reformatted
	packet.WriteU8(5)                               //Ver Major
	packet.WriteU8(0)                               //Ver Minor
	packet.WriteU8(call.Type)                       //Type
	packet.WriteU8(1 | 2)                           //Flags
	packet.WriteU32(16)                             //Representation
	packet.WriteU16(uint16(24 + len(call.PduData))) //Total length TODO Change
	packet.WriteU16(0)                              //Auth length
	packet.WriteU32(uint32(call.CallId))            //CallId
	packet.WriteU32(call.AllocHint)
	packet.WriteU16(call.CtxId)
	packet.WriteU16(call.Opnum)
	packet.WriteBytes(call.PduData)
	return packet.GetData()
}

// func NewDCERPCRawCall(opnum uint16, data []byte) DCERPCRawCall {
// 	return DCERPCRawCall{
// 		MSRPCRequestHeader: {},
// 	}
// }
