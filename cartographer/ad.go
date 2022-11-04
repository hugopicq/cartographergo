package cartographer

import (
	"crypto/tls"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

func ExecuteLDAPQuery(cred *Credentials, filter string, attributes []string, paging uint32, ldaps bool) ([]*ldap.Entry, error) {
	var ldapConnection *ldap.Conn
	var err error
	if ldaps {
		ldapConnection, err = ldap.DialURL("ldaps://"+cred.DomainController+":636", ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	} else {
		ldapConnection, err = ldap.DialURL("ldap://" + cred.DomainController + ":389")
	}

	if err != nil {
		return nil, err
	}
	defer ldapConnection.Close()

	err = ldapConnection.Bind(cred.User+"@"+cred.Domain, cred.Password)
	if err != nil {
		return nil, err
	}

	baseSuffix := ""
	for _, part := range strings.Split(cred.Domain, ".") {
		baseSuffix += ("DC=" + part + ",")
	}
	baseSuffix = baseSuffix[:len(baseSuffix)-1]

	searchReq := ldap.NewSearchRequest(baseSuffix, ldap.ScopeWholeSubtree, 0, 0, 0, false, filter, attributes, []ldap.Control{})

	result, err := ldapConnection.SearchWithPaging(searchReq, paging)
	if err != nil {
		return nil, err
	}

	return result.Entries, nil
}
