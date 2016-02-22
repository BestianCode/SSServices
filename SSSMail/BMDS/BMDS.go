package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/BestianRU/SSServices/SSModules/sscfg"
	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sssql"
)

type queryParam struct {
	StateCode      int
	StateName      string
	StateNameShort string
	Country        string
	From           string
}

func mailGetSubject(prm queryParam, conf sscfg.ReadJSONConfig, rLog sslog.LogFile) (string, bool) {
	f, err := os.Open(conf.Conf.BMDS_TitleDir + "/" + strconv.Itoa(prm.StateCode) + "_" + prm.StateName + ".txt")
	defer f.Close()
	reader := bufio.NewReader(f)
	tl, err := reader.ReadString('\n')
	if err != nil {
		rLog.Log("Error open title file!")
		return "", false
	}
	return tl, true
}

func mailGetBody(prm queryParam, conf sscfg.ReadJSONConfig, rLog sslog.LogFile) ([]byte, []byte, bool) {
	f1, err := os.Open(conf.Conf.BMDS_BodyDir + "/mail_" + prm.StateName + ".txt")
	if err != nil {
		rLog.Log("Error open TXT Body file!")
		return nil, nil, false
	}
	f2, err := os.Open(conf.Conf.BMDS_BodyDir + "/mail_" + prm.StateName + ".html")
	if err != nil {
		rLog.Log("Error open TXT Body file!")
		return nil, nil, false
	}
	defer f1.Close()
	defer f2.Close()
	reader1 := bufio.NewReader(f1)
	reader2 := bufio.NewReader(f2)
	contents1, _ := ioutil.ReadAll(reader1)
	contents2, _ := ioutil.ReadAll(reader2)
	return contents1, contents2, true
}

func mailGetMX(name string) ([]*net.MX, bool) {
	parts := strings.Split(name, "@")
	if len(parts) < 2 {
		return nil, false
	}
	mxs, err := net.LookupMX(parts[len(parts)-1])
	if err != nil {
		return nil, false
	}
	return mxs, true
}

func mailPrepare(prm queryParam, conf sscfg.ReadJSONConfig, rLog sslog.LogFile, dbase sssql.USQL) bool {
	var (
		//err   error
		query string
		mail  string
	)

	tl, statusTL := mailGetSubject(prm, conf, rLog)
	if !statusTL {
		return false
	}
	bodyTXT, bodyHTML, statusBD := mailGetBody(prm, conf, rLog)
	if !statusBD {
		return false
	}

	if prm.StateCode == 0 {
		query = conf.Conf.SQL_QUE1
	} else {
		query = conf.Conf.SQL_QUE2
	}
	query = strings.Replace(query, "'$3'", "'"+prm.StateName+"'", -1)
	query = strings.Replace(query, "'$4'", "'"+prm.Country+"'", -1)

	rLog.Log(query)
	rows, err := dbase.D.Query(query)
	if err != nil {
		rLog.Log("SQL::Query() error: ", err)
		rLog.Log(query)
		return false
	}
	for rows.Next() {
		rows.Scan(&mail)
		mx, statusMX := mailGetMX(mail)
		if statusMX {
			fmt.Printf("From: %s\n", prm.From)
			fmt.Printf("  To: %s\n", mail)
			fmt.Printf("Subj: %s\n", tl)
			fmt.Printf("------------\n%s\n------------\n", bodyTXT)
			fmt.Printf("------------\n%s\n------------\n", bodyHTML)
			for _, smx := range mx {
				fmt.Printf("\t\tMX: %s\n", smx.Host)
			}
		}
	}

	return true
}

//func updateMX(conf sscfg.ReadJSONConfig, rLog sslog.LogFile) {

/*
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
*/
//}

func main() {

	var (
		jsonConfig sscfg.ReadJSONConfig
		rLog       sslog.LogFile
		prm        queryParam
		DBase      sssql.USQL
		//err           error
		//sleep_counter = int(0)
	)

	const (
		pName = string("SSServices / BulkMailDirectSender")
		pVer  = string("1 2016.02.23.00.00")
	)

	fmt.Printf("\n\t%s V%s\n\n", pName, pVer)

	jsonConfig.Init("./BMDS.log", "./BMDS.json")

	rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
	rLog.Hello(pName, pVer)
	defer rLog.OFF()

	x := strings.Split(jsonConfig.Keys, " ")

	if len(x) > 4 {
		prm.StateCode, _ = strconv.Atoi(x[0])
		prm.StateName = x[1]
		prm.StateNameShort = x[2]
		prm.Country = x[3]
		prm.From = x[4]
		DBase.Init("MY", jsonConfig.Conf.MY_DSN, "")
		defer DBase.Close()
		if !mailPrepare(prm, jsonConfig, rLog, DBase) {
			fmt.Println("mailPrepare Error!")
			rLog.Log("mailPrepare Error!")
		} else {
			fmt.Println("Complete!")
			rLog.Log("Complete!")
		}
	} else {
		fmt.Println("Command Line Error!")
		rLog.Log("Command Line Error!")
	}
}
