package modules

import (
	"fmt"
	"time"

	"github.com/hugopicq/cartographergo/cartographer"
	"github.com/jlaffaye/ftp"
)

type ModuleFTPAnon struct{ enabled bool }

func (module *ModuleFTPAnon) IsEnabled() bool {
	return module.enabled
}

func (module *ModuleFTPAnon) Enable() {
	module.enabled = true
}

func (module *ModuleFTPAnon) GetName() string {
	return "FTPAnon"
}

func (module *ModuleFTPAnon) GetColumn() string {
	return "FTPAnon"
}

func (module *ModuleFTPAnon) GetPortFilter() []uint16 {
	return []uint16{21}
}

func (module *ModuleFTPAnon) Prepare(creds *cartographer.Credentials) error {
	return nil
}

func (module *ModuleFTPAnon) Run(ip string, hostname string, creds *cartographer.Credentials) (string, error) {

	c, err := ftp.Dial(fmt.Sprintf("%v:21", ip), ftp.DialWithTimeout(2*time.Second))
	if err != nil {
		//Connection not accepted
		return "", nil
	}

	err = c.Login("anonymous", "anonymous")
	if err != nil {
		//Anonymous not authorized
		return "0", nil
	}

	return "1", nil
}
