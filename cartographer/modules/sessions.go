package modules

import (
	"time"

	"github.com/hugopicq/cartographergo/cartographer"
	"github.com/hugopicq/cartographergo/utils"
	"golang.org/x/sys/windows/registry"
)

type SessionsModule struct {
	Users   map[string]string
	LDAPs   bool
	enabled bool
}

func NewSessionsModule(ldaps bool) *SessionsModule {
	module := new(SessionsModule)
	module.LDAPs = ldaps
	return module
}

func (module *SessionsModule) IsEnabled() bool {
	return module.enabled
}

func (module *SessionsModule) Enable() {
	module.enabled = true
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

func (module *SessionsModule) Filter(computer *cartographer.Computer) bool {
	return true
}

func (m *SessionsModule) Prepare(creds *cartographer.Credentials) error {
	//Get domain admins
	filter := "(&(objectClass=user)(objectCategory=Person)(adminCount=1))"
	attributes := []string{"objectSid", "name"}
	entries, err := cartographer.ExecuteLDAPQuery(creds, filter, attributes, 256, m.LDAPs)
	if err != nil {
		return err
	}

	m.Users = make(map[string]string)
	for _, entry := range entries {
		sid := utils.DecodeSID([]byte(entry.GetAttributeValue("objectSid")))
		m.Users[sid.String()] = entry.GetAttributeValue("name")
	}

	return nil
}

func (m *SessionsModule) Run(ip string, hostname string, credentials *cartographer.Credentials, timeout time.Duration) (string, error) {
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

	return utils.StringsToString(sessions), nil
}
