package modules

import (
	"github.com/hugopicq/cartographergo/cartographer"
	"github.com/hugopicq/cartographergo/dcerpc"
	"github.com/hugopicq/cartographergo/utils"
)

type ModuleRPC struct{}

func (module *ModuleRPC) GetName() string {
	return "RPC"
}

func (module *ModuleRPC) GetColumn() string {
	return "RPCInterfaces"
}

func (module *ModuleRPC) GetPortFilter() []uint16 {
	return []uint16{445}
}

func (module *ModuleRPC) Prepare(creds *cartographer.Credentials) error {
	return nil
}

func (module *ModuleRPC) Run(ip string, hostname string, creds *cartographer.Credentials) (string, error) {
	protocols, err := dcerpc.Dump(ip, creds.User, creds.Password, creds.Domain)
	if err != nil {
		return "", err
	}

	return utils.StringsToString(protocols), nil
}
