package dcerpc

import (
	"fmt"
	"net"

	"github.com/hirochachacha/go-smb2"
)

type SMBTransport struct {
	User          string
	Password      string
	Domain        string
	Target        string
	SMBConnection net.Conn
	SMBSession    *smb2.Session
	SMBShare      *smb2.Share
	SMBFile       *smb2.File
}

func NewSMBTransport(target string, user string, password string, domain string) *SMBTransport {
	transport := new(SMBTransport)
	transport.Target = target
	transport.User = user
	transport.Password = password
	transport.Domain = domain
	return transport
}

func (transport *SMBTransport) Connect() error {
	var err error

	transport.SMBConnection, err = net.Dial("tcp", fmt.Sprintf("%v:445", transport.Target))
	if err != nil {
		return err
	}

	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     transport.User,
			Password: transport.Password,
			Domain:   transport.Domain,
		},
	}

	transport.SMBSession, err = dialer.Dial(transport.SMBConnection)
	if err != nil {
		return err
	}

	transport.SMBShare, err = transport.SMBSession.Mount("IPC$")
	if err != nil {
		return err
	}

	transport.SMBShare.Open("aze")

	//Here we should create our own function to create the file
	filename := "epmapper"

	// req := &smb2library.CreateRequest{
	// 	SecurityFlags:        0,
	// 	RequestedOplockLevel: smb2library.SMB2_OPLOCK_LEVEL_NONE,
	// 	ImpersonationLevel:   smb2library.Impersonation,
	// 	SmbCreateFlags:       0,
	// 	DesiredAccess:        smb2library.FILE_READ_DATA | smb2library.FILE_WRITE_DATA,
	// 	FileAttributes:       smb2library.FILE_ATTRIBUTE_NORMAL,
	// 	ShareAccess:          smb2library.FILE_SHARE_READ,
	// 	CreateDisposition:    smb2library.FILE_OPEN,
	// 	CreateOptions:        smb2library.FILE_NON_DIRECTORY_FILE,
	// }

	transport.SMBFile, err = transport.SMBShare.Create(filename)
	if err != nil {
		panic(err)
	}

	return nil
}

func (transport *SMBTransport) Disconnect() {
	transport.SMBFile.Close()
	transport.SMBShare.Umount()
	transport.SMBSession.Logoff()
	transport.SMBConnection.Close()
}

func (transport *SMBTransport) Send(data []byte) error {
	_, err := transport.SMBFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (transport *SMBTransport) Receive() ([]byte, error) {
	buffer := make([]byte, 8192)
	n, err := transport.SMBFile.Read(buffer)
	if err != nil {
		return buffer, err
	}
	buffer = buffer[:n]
	return buffer, nil
}
