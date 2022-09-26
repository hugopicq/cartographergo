package dcerpc

import (
	"encoding/binary"
	"strconv"

	"github.com/hugopicq/cartographergo/utils"
)

type EPMTower struct {
	NFloors            uint16
	Interface          *EPMRPCInterface
	DataRepresentation *EPMRPCDataRepresentation
	Floors             []*EPMFloor
}

type EPMRPCInterface struct {
	LHSByteCount   uint16
	InterfaceIdent uint8
	InterfaceUUID  []byte
	MajorVersion   uint16
	RHSByteCount   uint16
	MinorVersion   uint16
}

func (rpcInterface EPMRPCInterface) ToString(version bool) string {
	toRet := utils.BinToString(rpcInterface.InterfaceUUID)
	if version {
		toRet += (" v" + strconv.Itoa(int(rpcInterface.MajorVersion)) + "." + strconv.Itoa(int(rpcInterface.MinorVersion)))
	}
	return toRet
}

func EPMRPCInterfaceFromBytes(data []byte) *EPMRPCInterface {
	rpcInterface := new(EPMRPCInterface)
	rpcInterface.LHSByteCount = binary.LittleEndian.Uint16(data[0:2])
	rpcInterface.InterfaceIdent = data[2]
	rpcInterface.InterfaceUUID = data[3:19]
	rpcInterface.MajorVersion = binary.LittleEndian.Uint16(data[19:21])
	rpcInterface.RHSByteCount = binary.LittleEndian.Uint16(data[21:23])
	rpcInterface.MinorVersion = binary.LittleEndian.Uint16(data[23:25])

	return rpcInterface
}

type EPMRPCDataRepresentation struct {
	LHSByteCount   uint16
	DrepIdentifier uint8
	DataRepUuid    []byte
	MajorVersion   uint16
	RHSByteCount   uint16
	MinorVersion   uint16
}

func EPMRPCDataRepresentationFromBytes(data []byte) *EPMRPCDataRepresentation {
	rpcRepresentation := new(EPMRPCDataRepresentation)
	rpcRepresentation.LHSByteCount = binary.LittleEndian.Uint16(data[0:2])
	rpcRepresentation.DrepIdentifier = data[2]
	rpcRepresentation.DataRepUuid = data[3:19]
	rpcRepresentation.MajorVersion = binary.LittleEndian.Uint16(data[19:21])
	rpcRepresentation.RHSByteCount = binary.LittleEndian.Uint16(data[21:23])
	rpcRepresentation.MinorVersion = binary.LittleEndian.Uint16(data[23:25])

	return rpcRepresentation
}

type EPMFloor struct {
	LHSByteCount uint16
	ProtocolData []byte
	RHSByteCount uint16
	RelatedData  []byte
}

func EPMFloorFromBytes(data []byte) (*EPMFloor, uint) {
	floor := new(EPMFloor)
	floor.LHSByteCount = binary.LittleEndian.Uint16(data[0:2])
	floor.ProtocolData = data[2 : 2+floor.LHSByteCount]
	floor.RHSByteCount = binary.LittleEndian.Uint16(data[2+floor.LHSByteCount : 4+floor.LHSByteCount])
	floor.RelatedData = data[4+floor.LHSByteCount : 4+floor.LHSByteCount+floor.RHSByteCount]
	return floor, uint(floor.LHSByteCount) + uint(floor.RHSByteCount) + 4
}

func EPMTowerFromBytes(data []byte) *EPMTower {
	tower := new(EPMTower)
	offset := uint(0)
	tower.NFloors = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	if tower.NFloors == 0 || len(data[offset:]) == 0 {
		return tower
	}
	//We have more
	tower.Interface = EPMRPCInterfaceFromBytes(data[offset:])
	offset += 25 //Length
	if tower.NFloors == 1 || len(data[offset:]) == 0 {
		return tower
	}

	tower.DataRepresentation = EPMRPCDataRepresentationFromBytes(data[offset:])
	offset += 25
	for i := 2; i < int(tower.NFloors); i++ {
		if len(data[offset:]) == 0 {
			break
		}
		floor, size := EPMFloorFromBytes(data[offset:])
		tower.Floors = append(tower.Floors, floor)
		offset += size
	}

	return tower
}
