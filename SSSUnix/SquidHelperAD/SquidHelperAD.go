package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/BestianRU/SSServices/SSModules/sscfg"
	"github.com/BestianRU/SSServices/SSModules/ssldap"
	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sssql"
)

func loginAuth(conf sscfg.ReadJSONConfig, rLog sslog.LogFile, user, domain, password string) (int, string) {
	var (
		DBase          sssql.USQL
		DL             ssldap.LDAP
		get            string
		sql_sel        string
		res            *sql.Rows
		err            error
		ldapAuthResult = int(-1)
		ldapGroupCheck = int(-1)
	)

	if strings.ToLower(conf.Conf.SQL_Engine) != "none" {
		sql_sel = strings.Replace(conf.Conf.SQH_SQL_PassCheck, "{USER}", user, -1)
		sql_sel = strings.Replace(sql_sel, "{DOMAIN}", domain, -1)
		sql_sel = strings.Replace(sql_sel, "{PASS}", password, -1)
		switch strings.ToLower(conf.Conf.SQL_Engine) {
		case "pg":
			DBase.Init("PG", conf.Conf.PG_DSN, "")
			res, err = DBase.D.Query(sql_sel)
		case "my":
			DBase.Init("MY", conf.Conf.MY_DSN, "")
			res, err = DBase.D.Query(sql_sel)
		default:
			rLog.LogDbg(0, "SQL Engine select error! (PG|MY|none)")
			return -1, "Error"
		}

		defer DBase.Close()

		if err != nil {
			rLog.LogDbg(0, "SQL: Query() select user/pass error: %v\n", err)
			return -1, "Error"
		}
		res.Next()
		get = ""
		res.Scan(&get)
		if get == user {
			rLog.LogDbg(2, "SQL Auth   for user: ", user, "@", domain, " - Ok!\n")
			return 0, "SQL"
		}
		res.Close()
		rLog.LogDbg(2, "SQL Auth   for user: ", user, "@", domain, " - No!\n")

		sql_sel = strings.Replace(conf.Conf.SQH_SQL_UserCheck, "{USER}", user, -1)
		sql_sel = strings.Replace(sql_sel, "{DOMAIN}", domain, -1)
		res, err = DBase.D.Query(sql_sel)
		if err != nil {
			rLog.LogDbg(0, "SQL: Query() select user error: %v\n", err)
			return -1, "Error"
		}
		res.Next()
		get = ""
		res.Scan(&get)
		if get != user {
			rLog.LogDbg(2, "SQL Search for user: ", user, "@", domain, " - No!\n")
			return -1, "Error"
		} else {
			rLog.LogDbg(2, "SQL Search for user: ", user, "@", domain, " - Ok!\n")
		}
		res.Close()
	}

	/*if strings.Contains(password, "==") {
		xdata, _ := base64.StdEncoding.DecodeString(password)
		log.Printf("%s", string(xdata))
	}*/
	if len(conf.Conf.SQH_LDAP_URL) > 0 {
		for _, y := range conf.Conf.SQH_LDAP_URL {
			if domain != "" {
				if domain == y[0] {
					ldapAuthResult = DL.InitS(rLog, user+"@"+y[0], password, y[1])
					if ldapAuthResult == 0 && conf.Conf.SQH_AD_GroupMember != "" {
						ldapGroupCheck = DL.CheckGroupMember(rLog, "(samaccountname="+user+")", "(samaccountname="+conf.Conf.SQH_AD_GroupMember+")", y[2])
					}
					DL.Close()
				}
			} else {
				ldapAuthResult = DL.InitS(rLog, user+"@"+y[0], password, y[1])
				if ldapAuthResult == 0 && conf.Conf.SQH_AD_GroupMember != "" {
					ldapGroupCheck = DL.CheckGroupMember(rLog, "(samaccountname="+user+")", "(samaccountname="+conf.Conf.SQH_AD_GroupMember+")", y[2])
				}
				DL.Close()
			}
			if ldapAuthResult == 0 {
				if ldapGroupCheck != 0 && conf.Conf.SQH_AD_GroupMember != "" {
					return -1, "not member of group"
				}
				return 0, "Domain: " + y[0]
			}
		}
	}
	return -1, "Error"
}

func main() {
	var (
		jsonConfig       sscfg.ReadJSONConfig
		rLog             sslog.LogFile
		err              error
		loginAuthResult  int
		arMessage        string
		loginAuthResultT = []string{"ERR", "OK"}
		logPassword      string
	)

	jsonConfig.Silent = "silent"

	const (
		pName = string("SSServices / SquidHelperAD")
		pVer  = string("1 2015.12.08.21.00")
	)

	jsonConfig.Init("./SquidHelperAD.log", "./SquidHelperAD.json")

	rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
	rLog.OFF()

	rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
	rLog.Hello(pName, pVer)
	rLog.OFF()

	in := bufio.NewReader(os.Stdin)
	input := ""
	for input != "." {
		rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)

		input, err = in.ReadString('\n')
		if err != nil {
			rLog.LogDbg(1, "Error read from STDIN: ", err)
		}

		if len(input) > 5 {
			uInput := strings.Split(strings.Replace(input, "\n", "", -1), " ")
			if len(uInput) > 1 {
				if strings.ToLower(jsonConfig.Conf.SQH_LogPasswords) != "yes" {
					logPassword = "hidden"
				} else {
					logPassword = uInput[1]
				}
				uInputx := strings.Split(uInput[0], "@")
				if len(uInputx) > 1 {
					loginAuthResult, arMessage = loginAuth(jsonConfig, rLog, uInputx[0], uInputx[1], uInput[1])
					rLog.Log("--> Login: ", uInputx[0], "@", uInputx[1], " (", logPassword, ") - ", loginAuthResultT[loginAuthResult+1], " (", arMessage, ")")
				} else {
					loginAuthResult, arMessage = loginAuth(jsonConfig, rLog, uInput[0], "", uInput[1])
					rLog.Log("--> Login: ", uInputx[0], " (", logPassword, ") - ", loginAuthResultT[loginAuthResult+1], " (", arMessage, ")")
				}
				fmt.Printf("%s\n", loginAuthResultT[loginAuthResult+1])
			} else {
				rLog.LogDbg(1, "Trash: ", input)
				fmt.Printf("ERR\n")
			}
		} else {
			rLog.LogDbg(1, "Trash: ", input)
			fmt.Printf("ERR\n")
		}
	}
	rLog.OFF()
}
