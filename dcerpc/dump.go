package dcerpc

import (
	"bytes"
	"encoding/binary"
)

func Dump(remoteHost string, user string, password string, domain string) ([]string, error) {
	transport := NewSMBTransport(remoteHost, user, password, domain)
	dce := NewDCE(transport)
	err := dce.Connect()
	if err != nil {
		return []string{}, err
	}
	resp, err := HeptLookup(dce)
	if err != nil {
		return []string{}, err
	}

	dce.Disconnect()

	protocols := []string{}
	mapProtocols := map[string]bool{}

	for _, entry := range resp {
		tower := EPMTowerFromBytes(entry.Tower.TowerOctetString)
		tmpUUID := tower.Interface.ToString(false)
		protocol, ok := KNOWN_PROTOCOLS[tmpUUID]
		if ok {
			_, exists := mapProtocols[protocol]
			if !exists {
				protocols = append(protocols, protocol)
				mapProtocols[protocol] = true
			}
		}
	}

	return protocols, nil
}

func HeptLookup(dce *DCE) ([]*EptEntry, error) {
	MSRPC_UUID_PORTMAP := []byte{0x08, 0x83, 0xaf, 0xe1, 0x1f, 0x5d, 0xc9, 0x11, 0x91, 0xa4, 0x08, 0x00, 0x2b, 0x14, 0xa0, 0xfa, 0x03, 0x00, 0x00, 0x00}
	err := dce.Bind(MSRPC_UUID_PORTMAP)
	if err != nil {
		return []*EptEntry{}, err
	}

	entries := []*EptEntry{}
	entryHandle := make([]byte, 20)
	for {
		//Here we have an infinite loop...
		request := new(EptLookup)
		request.Opnum = 2
		request.InquiryType = 0
		request.Object = 0
		request.Ifid = 0
		request.VersOption = 1
		request.EntryHandle = entryHandle
		request.MaxEnts = 500
		resp := dce.Request(request)

		entries = append(entries, resp.Entries...)

		if bytes.Compare(resp.EntryHandle.HandleUUID, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) == 0 {
			break
		}
		entryHandle = make([]byte, 0, 20)
		entryHandle = binary.LittleEndian.AppendUint32(entryHandle, resp.EntryHandle.HandleAttributes)
		entryHandle = append(entryHandle, resp.EntryHandle.HandleUUID...)
	}

	return entries, nil
}
