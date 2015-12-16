package main

import (
	"bufio"
	//"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/BestianRU/SSServices/SSModules/sscfg"
	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sssql"
)

func updateIP(conf sscfg.ReadJSONConfig, rLog sslog.LogFile, user, domain, ip string) int {
	var (
		DBase   sssql.USQL
		sql_sel string
		err     error
	)

	sql_sel = strings.Replace(conf.Conf.SQH_SQL_IPUpdate, "{USER}", user, -1)
	sql_sel = strings.Replace(sql_sel, "{DOMAIN}", domain, -1)
	sql_sel = strings.Replace(sql_sel, "{IP}", ip, -1)

	switch strings.ToLower(conf.Conf.SQL_Engine) {
	case "pg":
		DBase.Init("PG", conf.Conf.PG_DSN, sql_sel)
		_, err = DBase.D.Query(sql_sel)
	case "my":
		DBase.Init("MY", conf.Conf.MY_DSN, sql_sel)
		_, err = DBase.D.Query(sql_sel)
	default:
		rLog.LogDbg(0, "SQL Engine select error! (PG|MY)")
		return -1
	}

	defer DBase.Close()

	if err != nil {
		rLog.LogDbg(0, "SQL: Query() select user/pass error: %v\n", err)
		return -1
	}
	return 0
}

func main() {
	var (
		jsonConfig   sscfg.ReadJSONConfig
		rLog         sslog.LogFile
		err          error
		ipdateIPres  int
		ipdateIPresT = []string{"ERR", "OK"}
	)

	jsonConfig.Silent = "silent"

	const (
		pName = string("SSServices / SquidHelperDBUpdateIP")
		pVer  = string("1 2015.12.16.21.00")
	)

	jsonConfig.Init("./SquidHelperDBUpdateIP.log", "./SquidHelperDBUpdateIP.json")

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
				uInputx := strings.Split(uInput[1], "@")
				if len(uInputx) > 1 {
					ipdateIPres = updateIP(jsonConfig, rLog, uInputx[0], uInputx[1], uInput[0])
					rLog.Log("--> IP AD: ", uInputx[0], "@", uInputx[1], " - ", uInput[0])
				} else {
					ipdateIPres = updateIP(jsonConfig, rLog, uInput[1], "", uInput[0])
					rLog.Log("--> IP AD: ", uInput[1], " - ", uInput[0])
				}
				fmt.Printf("%s\n", ipdateIPresT[ipdateIPres+1])
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
