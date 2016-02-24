package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BestianRU/SSServices/SSModules/sscfg"
	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sssql"
)

var (
	instOfSenders int
	cntSucc       int
	cntFail       int
	cntAll        int

	SQLCreateTable1 = string(`
create table if not exists bmds_domain (
	id int(10) unsigned not null auto_increment,
	domain varchar(255),
	primary key(id),
	unique key(domain)
) engine=innodb;`)
	SQLCreateTable2 = string(`
create table if not exists bmds_mx (
	id int(10) unsigned not null auto_increment,
	pid int(10) unsigned not null,
	mx varchar(255),
	ip int unsigned,
	weight int unsigned,
	primary key(id),
	unique (pid,mx,ip)
) engine=innodb;`)
)

type queryParam struct {
	StateCode      int
	StateName      string
	StateNameShort string
	Country        string
	From           string
}

func mailCreate(prm queryParam, to, subject string, bodyTXT, bodyHTML []byte, conf sscfg.ReadJSONConfig, rLog sslog.LogFile) ([]byte, bool) {
	t := time.Now()
	rnd := t.Format("20060102.150405")
	senderDomain := strings.Split(prm.From, "@")

	msg := []byte("From: " + conf.Conf.BMDS_SenderName + " <" + prm.From + ">\r\n" +
		"Return-Path: <" + prm.From + ">\r\n" +
		"Message-ID: <" + rnd + "@" + senderDomain[len(senderDomain)-1] + ">\r\n" +
		"X-Mailjet-Campaign: " + subject + "\r\n" +
		"To: " + to + "\r\n" +
		"List-Unsubscribe: <mailto:qaka@qakadeals.com?subject=unsubscribe>\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=NextPart_" + rnd + "\r\n" +
		"\r\n" +
		"--NextPart_" + rnd + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		fmt.Sprintf("%s", bodyTXT) + "\r\n" +
		"\r\n" +
		"--NextPart_" + rnd + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		fmt.Sprintf("%s", bodyHTML) + "\r\n" +
		"\r\n" +
		"--NextPart_" + rnd + "\r\n")

	return msg, true
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
	tl = strings.Replace(tl, "\r", "", -1)
	tl = strings.Replace(tl, "\n", "", -1)
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

func mailGetMX(name string, rLog sslog.LogFile, dbase sssql.USQL) ([]*net.MX, bool) {
	var (
		mx    []*net.MX
		tmx   net.MX
		query string
		xid   int
	)
	parts := strings.Split(name, "@")
	if len(parts) < 2 {
		return nil, false
	}

	query = "select inet_ntoa(x.ip),x.weight from bmds_mx as x left join bmds_domain as y on (x.pid=y.id) where y.domain='" + parts[len(parts)-1] + "';"
	//rLog.LogDbg(3, "DNS Search: ", query)
	rows, err := dbase.D.Query(query)
	if err != nil {
		rLog.Log("SQL::Query() error: ", err)
		rLog.Log(query)
		return nil, false
	}
	for rows.Next() {
		rows.Scan(&tmx.Host, &tmx.Pref)
		mx = append(mx, &tmx)
	}

	rows.Close()

	if len(mx) > 0 {
		rLog.LogDbg(3, "NAME GET: SQL")
		return mx, true
	}

	mxs, err := net.LookupMX(parts[len(parts)-1])
	if err != nil {
		return nil, false
	}
	rLog.LogDbg(3, "NAME GET: DNS")

	query = "insert into bmds_domain (domain) values ('" + parts[len(parts)-1] + "');"
	dbase.Silent = 1
	_ = dbase.QSimple(query)
	dbase.Silent = 0

	query = "select id from bmds_domain where domain='" + parts[len(parts)-1] + "';"
	rows, err = dbase.D.Query(query)
	if err != nil {
		rLog.Log("SQL::Query() error: ", err)
		rLog.Log(query)
		return nil, false
	}
	rows.Next()
	rows.Scan(&xid)
	rows.Close()

	dbase.Silent = 1
	for _, mxi := range mxs {
		lnameArray, err := net.LookupIP(mxi.Host)
		if err != nil {
			continue
		}
		for _, lname := range lnameArray {
			//fmt.Printf("%s\n", lname)
			query = "insert into bmds_mx (pid,mx,ip,weight) values (" + strconv.Itoa(xid) + ",'" + mxi.Host + "',inet_aton('" + fmt.Sprintf("%s", lname) + "')," + strconv.Itoa(int(mxi.Pref)) + ");"
			//rLog.LogDbg(3, "Insert IP: ", query)
			_ = dbase.QSimple(query)
		}

	}
	dbase.Silent = 0
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
		cntAll++
		fmt.Printf("sent: %d, fail: %d, count: %d, summ: (%d), instances: %d\n", cntSucc, cntFail, cntAll, int(cntFail+cntSucc), instOfSenders)
		mx, statusMX := mailGetMX(mail, rLog, dbase)
		if statusMX {
			fullMail, statusFM := mailCreate(prm, mail, tl, bodyTXT, bodyHTML, conf, rLog)
			if statusFM {
				for {
					if instOfSenders < conf.Conf.BMDS_MaxInstances {
						break
					}
					rLog.Log("Wait for Goroutines: ", instOfSenders, " (Allowed:", conf.Conf.BMDS_MaxInstances, ")")
					time.Sleep(time.Duration(2) * time.Second)
				}
				//rLog.LogDbg(3, "PREP ", prm.From, " -> ", mail)
				go mailSendMXRotate(fullMail, prm.From, mail, mx, conf, rLog)
			}
		}
	}
	rows.Close()

	return true
}

func mailSendMXRotate(body []byte, headFrom, headTo string, servers []*net.MX, conf sscfg.ReadJSONConfig, rLog sslog.LogFile) {
	//rLog.LogDbg(3, "MX ROTATE ", headFrom, " -> ", headTo)
	instOfSenders++
	for _, mx := range servers {
		if mailSend(body, headFrom, headTo, mx.Host, conf, rLog) {
			cntSucc++
			break
		}
		cntFail++
	}
	instOfSenders--
}

func mailSend(body []byte, headFrom, headTo, server string, conf sscfg.ReadJSONConfig, rLog sslog.LogFile) bool {
	x := rand.Intn(len(conf.Conf.BMDS_IPList))
	rLog.LogDbg(3, "MAIL SEND ", conf.Conf.BMDS_IPList[x], " ])> ", headFrom, " -> ", headTo, " ---> ", server)
	//net.Dialer.LocalAddr =
	time.Sleep(time.Duration(10) * time.Second)
	return true
	c, err := smtp.Dial(server + ":25")
	if err != nil {
		rLog.Log("SMTP: ", server, " connect error for ", headFrom, "->", headTo, " /// ", err)
		return false
	}
	defer c.Close()
	c.Mail(headFrom)
	c.Rcpt(headTo)

	wc, err := c.Data()
	if err != nil {
		rLog.Log("Body ", headFrom, "->", headTo, " error /// ", err)
		return false
	}
	defer wc.Close()

	buf := bytes.NewBufferString(fmt.Sprintf("%s", body))
	if _, err = buf.WriteTo(wc); err != nil {
		rLog.Log("Send ", headFrom, "->", headTo, " error /// ", err)
		return false
	}
	rLog.Log(headFrom, "->", headTo, " via ", server, " - Sent")
	return true
}

func main() {

	var (
		jsonConfig sscfg.ReadJSONConfig
		rLog       sslog.LogFile
		prm        queryParam
		DBase      sssql.USQL
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

	jsonConfig.Conf.SQL_Engine = "MY"

	x := strings.Split(jsonConfig.Keys, " ")

	if len(x) > 4 {

		prm.StateCode, _ = strconv.Atoi(x[0])
		prm.StateName = x[1]
		prm.StateNameShort = x[2]
		prm.Country = x[3]
		prm.From = x[4]

		DBase.Init("MY", jsonConfig.Conf.MY_DSN, "")
		DBase.QSimple(SQLCreateTable1)
		DBase.QSimple(SQLCreateTable2)

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

	for {
		if instOfSenders > 0 {
			rLog.Log("Wait for complete all Gooutines: ", instOfSenders)
		} else {
			break
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
	rLog.Log("Bye!")
}
