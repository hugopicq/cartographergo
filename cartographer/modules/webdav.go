package modules

import (
	"fmt"
	"net"

	"github.com/hirochachacha/go-smb2"
	"github.com/hugopicq/cartographergo/cartographer"
)

type ModuleWebDAV struct{}

func (module *ModuleWebDAV) GetName() string {
	return "WebDAV"
}

func (module *ModuleWebDAV) GetColumn() string {
	return "WebDAV"
}

func (module *ModuleWebDAV) GetPortFilter() []uint16 {
	return []uint16{445}
}

func (module *ModuleWebDAV) Prepare(creds *cartographer.Credentials) error {
	return nil
}

func (module *ModuleWebDAV) Run(ip string, hostname string, creds *cartographer.Credentials) (string, error) {

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

	fs, err := c.Mount("IPC$")
	if err != nil {
		//We do nothing and we try the next one
		return "", err
	}
	defer fs.Umount()

	f, err := fs.Open("DAV RPC Service")
	if err != nil {
		return "0", nil
	}
	defer f.Close()

	return "1", nil
}
