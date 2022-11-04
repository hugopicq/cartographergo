package modules

import (
	"fmt"
	"net"
	"time"

	"github.com/hugopicq/cartographergo/cartographer"
	"github.com/hugopicq/cartographergo/utils"

	"github.com/hirochachacha/go-smb2"
)

type ModuleListShares struct{ enabled bool }

func (module *ModuleListShares) IsEnabled() bool {
	return module.enabled
}

func (module *ModuleListShares) Enable() {
	module.enabled = true
}

func (module *ModuleListShares) GetName() string {
	return "ListShares"
}

func (module *ModuleListShares) GetColumn() string {
	return "ReadShares"
}

func (module *ModuleListShares) Filter(computer *cartographer.Computer) bool {
	return true
}

func (module *ModuleListShares) GetPortFilter() []uint16 {
	return []uint16{445}
}

func (module *ModuleListShares) Prepare(creds *cartographer.Credentials) error {
	return nil
}

func (module *ModuleListShares) Run(ip string, hostname string, creds *cartographer.Credentials, timeout time.Duration) (string, error) {

	results := []string{}

	conn, err := net.Dial("tcp", fmt.Sprintf("%v:445", ip))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     (*creds).User,
			Password: (*creds).Password,
			Domain:   (*creds).Domain,
		},
	}

	c, err := d.Dial(conn)
	if err != nil {
		return "", err
	}
	defer c.Logoff()

	names, err := c.ListSharenames()
	if err != nil {
		return "", err
	}

	for _, name := range names {
		fs, err := c.Mount(name)
		if err != nil {
			//We do nothing and we try the next one
			continue
		}
		defer fs.Umount()

		f, err := fs.Open("")
		if err != nil {
			//We don't have read rights
		} else {
			defer f.Close()
			results = append(results, name)
		}
	}

	//We format directly here but with the use of generics and reflect we might be able to do better in the future
	stringresult := utils.StringsToString(results)
	return stringresult, nil
}
