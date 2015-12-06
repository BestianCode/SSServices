package ssldap

import (
	// LDAP
	"github.com/go-ldap/ldap"

	"github.com/BestianRU/SSServices/SSModules/sslog"
)

type LDAP struct {
	CS          int
	LDAPCounter int
	D           *ldap.Conn
}

func (_s *LDAP) Init(xldap_url [][]string, rLog sslog.LogFile) int {
	var (
		attemptCounter = int(0)
		err            error
	)

	_s.CS = -1

	for {
		if attemptCounter > len(xldap_url)*2 {
			rLog.LogDbg(0, "LDAP Init SRV ***** Error connect to all LDAP servers !!!")
			return -1
		}

		if _s.LDAPCounter > len(xldap_url)-1 {
			_s.LDAPCounter = 0
		}

		rLog.LogDbg(2, "LDAP Init SRV ***** Trying connect to server ", _s.LDAPCounter+1, " of ", len(xldap_url), ": ", xldap_url[_s.LDAPCounter][0])
		_s.D, err = ldap.Dial("tcp", xldap_url[_s.LDAPCounter][0])
		if err != nil {
			_s.LDAPCounter++
			attemptCounter++
			continue
		}

		rLog.LogDbg(2, "LDAP Init SRV ***** Success! Connected to server ", _s.LDAPCounter+1, " of ", len(xldap_url), ": ", xldap_url[_s.LDAPCounter][0])
		_s.LDAPCounter++
		break
	}

	//_s.D.Debug()

	err = _s.D.Bind(xldap_url[0][1], xldap_url[0][2])
	if err != nil {
		rLog.LogDbg(0, "LDAP::Bind() to server ", xldap_url[_s.LDAPCounter][0], " with login ", xldap_url[0][1], " error: ", err)
		return -1
	}

	_s.CS = 0
	return 0
}

func (_s *LDAP) InitS(rLog sslog.LogFile, user, password, server string) int {
	var err error

	_s.CS = -1

	rLog.LogDbg(2, "LDAP Init SRV ***** Trying connect to server ", server, " with login ", user)

	_s.D, err = ldap.Dial("tcp", server)
	if err != nil {
		rLog.LogDbg(0, "LDAP::Dial() to server ", server, " error: ", err)
		return -1
	}

	//L.Debug()

	err = _s.D.Bind(user, password)
	if err != nil {
		rLog.LogDbg(1, "LDAP::Bind() to server ", server, " with login ", user, " error: ", err)
		return -1
	}

	rLog.LogDbg(2, "LDAP Init SRV ***** Success! Connected to server ", server, " with login ", user)

	_s.CS = 0
	return 0
}

func (_s *LDAP) CheckGroupMember(rLog sslog.LogFile, user, group, baseDN string) int {
	const (
		recurs_count = 10
	)

	rLog.LogDbg(2, "LDAP CheckGroupMember...")

	userDN := _s._getBaseDN(rLog, user, baseDN)
	groupDN := _s._getBaseDN(rLog, group, baseDN)

	if userDN == "" || groupDN == "" {
		return -1
	}

	if _s._checkGroupMember(rLog, userDN, groupDN, baseDN, 1) == 0 {
		return 0
	} else {
		return _s._checkGroupMember(rLog, userDN, groupDN, baseDN, recurs_count)
	}

	return -1
}

func (_s *LDAP) _checkGroupMember(rLog sslog.LogFile, userDN, groupDN, baseDN string, recurse_count int) int {
	var (
		uattr  = []string{"memberOf"}
		result = int(-1)
	)

	if userDN == "" || groupDN == "" {
		return -1
	}

	if recurse_count <= 0 {
		return -1
	}

	lsearch := ldap.NewSearchRequest(userDN, 0, ldap.NeverDerefAliases, 0, 0, false, "(objectclass=*)", uattr, nil)
	sr, err := _s.D.Search(lsearch)
	if err != nil {
		rLog.LogDbg(0, "LDAP::Search() ", userDN, " error: ", err)
	}

	if len(sr.Entries) > 0 {
		for _, entry := range sr.Entries {
			for _, attr := range entry.Attributes {
				if attr.Name == "memberOf" {
					for _, x := range attr.Values {
						if groupDN == x {
							return 0
						} else {
							if x != userDN {
								result = _s._checkGroupMember(rLog, x, groupDN, baseDN, recurse_count-1)
								if result == 0 {
									return 0
								}
							}
						}
					}
				}
			}
		}
	}
	return -1
}

func (_s *LDAP) _getBaseDN(rLog sslog.LogFile, search, basedn string) string {
	var uattr = []string{"dn"}

	lsearch := ldap.NewSearchRequest(basedn, 2, ldap.NeverDerefAliases, 0, 0, false, search, uattr, nil)
	sr, err := _s.D.Search(lsearch)
	if err != nil {
		rLog.LogDbg(0, "LDAP::Search() ", basedn, " error: ", err)
	}

	if len(sr.Entries) > 0 {
		for _, entry := range sr.Entries {
			return entry.DN
		}
	}
	return ""
}

func (_s *LDAP) Close() {
	if _s.CS != -1 {
		_s.D.Close()
	}
}
