package modules

import (
	"github.com/hugopicq/cartographergo/cartographer"
	"github.com/hugopicq/cartographergo/util"
	"golang.org/x/sys/windows/registry"
)

type SessionsModule struct {
	Users map[string]string
}

func (m *SessionsModule) GetPortFilter() []uint16 {
	return []uint16{445}
}

func (m *SessionsModule) GetName() string {
	return "Sessions"
}

func (m *SessionsModule) GetColumn() string {
	return "Sessions"
}

func (m *SessionsModule) Prepare(creds *cartographer.Credentials) error {
	//Get domain admins
	filter := "(&(objectClass=user)(objectCategory=Person)(adminCount=1))"
	attributes := []string{"objectSid", "name"}
	entries, err := cartographer.ExecuteLDAPQuery(creds, filter, attributes, 256)
	if err != nil {
		return err
	}

	m.Users = make(map[string]string)
	for _, entry := range entries {
		sid := util.DecodeSID([]byte(entry.GetAttributeValue("objectSid")))
		m.Users[sid.String()] = entry.GetAttributeValue("name")
	}

	return nil
}

func (m *SessionsModule) Run(ip string, hostname string, credentials *cartographer.Credentials) (string, error) {
	k, err := registry.OpenRemoteKey(ip, registry.USERS)
	if err != nil {
		return "", err
	}
	defer k.Close()

	keys, err := k.ReadSubKeyNames(0)
	if err != nil {
		return "", err
	}

	sessions := make([]string, 0, len(keys))

	for _, sid := range keys {
		user := m.Users[sid]
		if user != "" {
			sessions = append(sessions, user)
		}
	}

	return util.StringsToString(sessions), nil
}
