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

	dkim "github.com/toorop/go-dkim"

	"github.com/BestianRU/SSServices/SSModules/sscfg"
	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sssql"
)

var (
	instOfSenders                int
	cntSucc                      int
	cntFail                      int
	cntAll                       int
	options                      dkim.SigOptions
	rLog, rLogSc, rLogFl, rLogDb sslog.LogFile
	//slowSend                     map[string]int64
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
	Mode           string
	Limit          string
}

func printAll(msg ...interface{}) {
	fmt.Println(msg)
	rLog.Log(msg)
	rLogSc.Log(msg)
	rLogFl.Log(msg)
}

func mailCreate(prm queryParam, to, subject string, bodyTXT, bodyHTML []byte, conf sscfg.ReadJSONConfig) ([]byte, bool) {
	t := time.Now()
	rnd := t.Format("20060102.150405")
	senderDomain := strings.Split(prm.From, "@")

	msg := []byte("From: " + conf.Conf.BMDS_SenderName + " <" + prm.From + ">\r\n" +
		"Return-Path: <" + prm.From + ">\r\n" +
		"Message-ID: <" + rnd + "@" + senderDomain[len(senderDomain)-1] + ">\r\n" +
		"X-Mailjet-Campaign: " + subject + "\r\n" +
		"Date: " + t.Format(time.RFC1123Z) + "\r\n" +
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

func mailGetSubject(prm queryParam, conf sscfg.ReadJSONConfig) (string, bool) {
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

func mailGetBody(prm queryParam, conf sscfg.ReadJSONConfig) ([]byte, []byte, bool) {
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

func mailGetDKIM(conf sscfg.ReadJSONConfig) []byte {
	f1, err := os.Open(conf.Conf.BDMS_DKIMKey)
	if err != nil {
		rLog.Log("Error open DKIM file!")
		return nil
	}
	defer f1.Close()
	reader1 := bufio.NewReader(f1)
	contents1, _ := ioutil.ReadAll(reader1)
	return contents1
}

func mailGetMX(name string, dbase sssql.USQL) ([]*net.MX, bool) {
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

	query = "select distinct inet_ntoa(x.ip) from bmds_mx as x left join bmds_domain as y on (x.pid=y.id) where y.domain='" + parts[len(parts)-1] + "' and inet_ntoa(x.ip) like '%.%.%.%' order by x.weight;"
	//rLogDb.LogDbg(3, "DNS Search: ", query)
	rows, err := dbase.D.Query(query)
	if err != nil {
		rLog.Log("SQL::Query() error: ", err)
		rLog.Log(query)
		return nil, false
	}
	for rows.Next() {
		rows.Scan(&tmx.Host)
		mx = append(mx, &tmx)
	}

	rows.Close()

	if len(mx) > 0 {
		//rLogDb.LogDbg(3, "NAME GET: SQL")
		return mx, true
	}

	mxs, err := net.LookupMX(parts[len(parts)-1])
	if err != nil {
		return nil, false
	}
	//rLogDb.LogDbg(3, "NAME GET: DNS")

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
			//rLogDb.LogDbg(3, "Insert IP: ", query)
			_ = dbase.QSimple(query)
		}

	}
	dbase.Silent = 0
	return mxs, true
}

func mailPrepare(prm queryParam, conf sscfg.ReadJSONConfig, dbase sssql.USQL) bool {
	var (
		err           error
		query         string
		mail          string
		countBusy     int64
		countBusyPast int64
	)

	tl, statusTL := mailGetSubject(prm, conf)
	if !statusTL {
		rLog.Log("Error get subject!")
		return false
	}
	bodyTXT, bodyHTML, statusBD := mailGetBody(prm, conf)
	if !statusBD {
		rLog.Log("Error get body!")
		return false
	}

	if prm.StateCode == 0 {
		query = conf.Conf.SQL_QUE1
	} else {
		query = conf.Conf.SQL_QUE2
	}
	query = strings.Replace(query, "'$3'", "'"+prm.StateNameShort+"'", -1)
	query = strings.Replace(query, "'$4'", "'"+prm.Country+"'", -1)

	rLogDb.LogDbg(3, "224:", prm.Limit)
	xlimit, err := strconv.Atoi(prm.Limit)
	rLogDb.LogDbg(3, "226:", xlimit)
	if err == nil {
		if xlimit > 0 {
			query = query + " limit " + prm.Limit
		}
	} else {
		rLogDb.LogDbg(3, "232:", err)
	}

	options = dkim.NewSigOptions()
	options.PrivateKey = mailGetDKIM(conf)
	options.Domain = conf.Conf.BDMS_DKIMDomain
	options.Selector = conf.Conf.BDMS_DKIMSelector
	options.SignatureExpireIn = 3600
	options.BodyLength = 50
	options.Headers = []string{"To", "Subject", "From", "Date"}
	options.AddSignatureTimestamp = true
	options.Canonicalization = "relaxed/relaxed"
	//options.Canonicalization = "relaxed/simple"

	rLog.Log(query)
	rows, err := dbase.D.Query(query)
	if err != nil {
		rLog.Log("SQL::Query() error: ", err)
		rLog.Log(query)
		return false
	}
	for rows.Next() {
		rows.Scan(&mail)
		rLogDb.LogDbg(3, "238:", mail)
		cntAll++
		if int(cntAll/10)*10 == cntAll {
			fmt.Printf("sent: %d, wait: %d, count: %d, instances: %d\n", cntSucc, int(cntAll-cntSucc), cntAll, instOfSenders)
			rLog.Log("sent: ", cntSucc, ", wait: ", int(cntAll-cntSucc), ", count: ", cntAll, ", instances: ", instOfSenders)
		}

		fullMail, statusFM := mailCreate(prm, mail, tl, bodyTXT, bodyHTML, conf)
		if statusFM {
			if conf.Conf.BMDS_MaxInstances < 700 {
				if instOfSenders > int(conf.Conf.BMDS_MaxInstances/100)*97 {
					if countBusyPast < 1 {
						busyTimeNow := time.Now()
						countBusyPast = busyTimeNow.Unix()
					}
					busyTimeNow := time.Now()
					countBusy += (busyTimeNow.Unix() - countBusyPast)
					if countBusy > 300 {
						conf.Conf.BMDS_MaxInstances += 100
						countBusy = 0
						countBusyPast = 0
					}
				}
				if countBusy > 0 {
					rLog.Log("Busy: ", countBusy)
				}
				if instOfSenders < int(conf.Conf.BMDS_MaxInstances/100)*95 {
					countBusy = 0
					countBusyPast = 0
				}
			}
			if instOfSenders >= conf.Conf.BMDS_MaxInstances {
				for {
					if instOfSenders < conf.Conf.BMDS_MaxInstances {
						break
					}
					rLog.Log("Wait for Goroutines: ", instOfSenders, " (Allowed:", conf.Conf.BMDS_MaxInstances, ")")
					time.Sleep(time.Duration(2) * time.Second)
				}
			}
			rLogDb.LogDbg(3, "262:", prm.From, " -> ", mail)
			go mailSendMXRotate(fullMail, prm.From, mail, conf, dbase)
		}

	}
	rows.Close()

	return true
}

func mailSendMXRotate(body []byte, headFrom, headTo string, conf sscfg.ReadJSONConfig, dbase sssql.USQL) {
	var count = int(3)

	instOfSenders++
	rLogDb.LogDbg(3, "269:", headTo)
	servers, statusMX := mailGetMX(headTo, dbase)
	rLogDb.LogDbg(3, "272:", fmt.Sprintf("%v", servers))
	if statusMX {
		for _, mx := range servers {
			rLogDb.LogDbg(3, "280:", mx.Host)
			if count < 2 {
				break
			}
			count--
			if mailSend(body, headFrom, headTo, mx.Host, conf, dbase) {
				cntSucc++
				break
			}
			cntFail++
		}
	}
	instOfSenders--
}

func mailSend(body []byte, headFrom, headTo, server string, conf sscfg.ReadJSONConfig, dbase sssql.USQL) bool {
	var (
		x     int
		query string
	)

	rLogDb.LogDbg(3, "301:", headFrom, " -> ", headTo, " <><><> ", server)

	if len(conf.Conf.BMDS_IPList) > 1 {
		x = rand.Intn(len(conf.Conf.BMDS_IPList) - 1)
	} else {
		x = 0
	}

	ief, err := net.InterfaceByName(conf.Conf.BMDS_IPList[x])

	if err != nil {
		rLog.Log("net.InterfaceByName /// ", err)
		return false
	}
	addrs, err := ief.Addrs()
	if err != nil {
		rLog.Log("ief.Addrs /// ", err)
		return false
	}

	if len(addrs) > 1 {
		x = rand.Intn(len(addrs) - 1)
	} else {
		x = 0
	}

	tcpAddr := &net.TCPAddr{IP: addrs[x].(*net.IPNet).IP}

	d := net.Dialer{Timeout: time.Duration(6) * time.Second, LocalAddr: tcpAddr}

	/*
		timeNow := time.Now()
		senderDomain := strings.Split(headTo, "@")
		for _, slowx := range conf.Conf.BMDS_SlowMail {
			if senderDomain[len(senderDomain)-1] == slowx {
				//rLogDb.LogDbg(3, "337:", fmt.Sprintf("%v", tcpAddr))
				//rLogDb.LogDbg(3, "338:", slowSend[fmt.Sprintf("%v", tcpAddr)])
				//rLogDb.LogDbg(3, "339:", int64(slowSend[fmt.Sprintf("%v", tcpAddr)]))
				if int64(slowSend[fmt.Sprintf("%v", tcpAddr)]) > int64(0) {
					if (timeNow.Unix() - slowSend[fmt.Sprintf("%v", tcpAddr)]) < 5 {
						rLog.Log("slow for: ", headTo, " on ", fmt.Sprintf("%v", tcpAddr))
					}
				}
			}
		}
	*/
	if conf.Conf.BMDS_SlowMailDelay > 0 {
		time.Sleep(time.Duration(conf.Conf.BMDS_SlowMailDelay) * time.Second)
	}
	conn, err := d.Dial("tcp4", server+":25")
	if err != nil {
		rLog.Log("d.Dial /// ", err)
		query = "update members set status=-34 where email='" + headTo + "';"
		_ = dbase.QSimple(query)
		query = "delete from bmds_mx where ip=inet_aton('" + server + "');"
		_ = dbase.QSimple(query)
		return false
	}
	host, _, _ := net.SplitHostPort(server + ":25")
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		rLogFl.Log("SMTP: ", server, " connect error for ", headFrom, "->", headTo, " /// ", err)
		return false
	}
	err = c.Hello(conf.Conf.BDMS_DKIMDomain)
	if err != nil {
		rLogFl.Log("SMTP: ", server, " hello error for ", headFrom, "->", headTo, " /// ", err)
		return false
	}

	//c, err := smtp.Dial(server + ":25")
	if err != nil {
		rLogFl.Log("SMTP: ", server, " connect error for ", headFrom, "->", headTo, " /// ", err)
		query = "update members set status=-33 where email='" + headTo + "';"
		_ = dbase.QSimple(query)
		//timeNow = time.Now()
		//slowSend[fmt.Sprintf("%v", tcpAddr)] = timeNow.Unix()
		return false
	}
	defer c.Close()
	c.Mail(headFrom)
	c.Rcpt(headTo)

	wc, err := c.Data()
	if err != nil {
		if fmt.Sprintf("%v", err) == "EOF" {
			rLogFl.Log("IP: ", fmt.Sprintf("%v", tcpAddr), " mail ", headFrom, "->", headTo, " - isprejected")
		} else {
			query = "update members set status=-32 where email='" + headTo + "';"
			_ = dbase.QSimple(query)
		}
		return false
	}
	defer wc.Close()

	//rLogDb.LogDbg(3, "----------\n", fmt.Sprintf("%s", body), "----------\n")
	//rLogDb.LogDbg(3, "----------\n", fmt.Sprintf("%s", options.PrivateKey), "----------\n")
	err = dkim.Sign(&body, options)
	//rLogDb.LogDbg(3, "----------\n", fmt.Sprintf("%s", body), "----------\n")

	buf := bytes.NewBufferString(fmt.Sprintf("%s", body))
	if _, err = buf.WriteTo(wc); err != nil {
		rLogFl.Log("Send ", headFrom, "->", headTo, " error /// ", err)
		query = "update members set status=-31 where email='" + headTo + "';"
		_ = dbase.QSimple(query)
		//timeNow = time.Now()
		//slowSend[fmt.Sprintf("%v", tcpAddr)] = timeNow.Unix()
		return false
	}
	rLogSc.Log("(IP:", fmt.Sprintf("%v", tcpAddr), ") ", headFrom, "->", headTo, " via ", server, " - Sent")
	//timeNow = time.Now()
	//slowSend[fmt.Sprintf("%v", tcpAddr)] = timeNow.Unix()
	return true
}

func main() {

	var (
		jsonConfig          sscfg.ReadJSONConfig
		prm                 queryParam
		DBase               sssql.USQL
		exitCounter         = int(10)
		timeStart, timeExec int64
	)

	const (
		pName = string("SSServices / BulkMailDirectSender")
		pVer  = string("1 2016.03.04.21.00")
	)

	fmt.Printf("\n\t%s V%s\n\n", pName, pVer)

	jsonConfig.Init("./BMDS.log", "./BMDS.json")

	rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
	rLogSc.ON(jsonConfig.Conf.LOG_File+".sent", jsonConfig.Conf.LOG_Level)
	rLogFl.ON(jsonConfig.Conf.LOG_File+".fail", jsonConfig.Conf.LOG_Level)
	rLogDb.ON(jsonConfig.Conf.LOG_File+".debug", jsonConfig.Conf.LOG_Level)
	rLog.Hello(pName, pVer)
	rLogSc.Hello(pName, pVer)
	rLogFl.Hello(pName, pVer)
	defer rLog.OFF()
	defer rLogSc.OFF()
	defer rLogFl.OFF()
	defer rLogDb.OFF()

	jsonConfig.Conf.SQL_Engine = "MY"

	x := strings.Split(jsonConfig.Keys, " ")

	//slowSend = make(map[string]int64, 1000)

	timeNow := time.Now()
	timeStart = timeNow.Unix()

	if len(x) > 4 {

		prm.StateCode, _ = strconv.Atoi(x[0])
		prm.StateName = x[1]
		prm.StateNameShort = x[2]
		prm.Country = x[3]
		prm.From = x[4]
		//prm.Mode = x[5]
		if len(x) > 6 {
			prm.Limit = x[6]
		} else {
			prm.Limit = "0"
		}

		DBase.Init("MY", jsonConfig.Conf.MY_DSN, "")
		DBase.QSimple(SQLCreateTable1)
		DBase.QSimple(SQLCreateTable2)

		defer DBase.Close()
		if !mailPrepare(prm, jsonConfig, DBase) {
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
			timeNow = time.Now()
			timeExec = int64(timeNow.Unix() - timeStart)
			rLog.Log("Wait for complete all Goroutines: ", instOfSenders)
			fmt.Printf("Wait for complete all Goroutines: %d\n", instOfSenders)
			printAll("Time: ", timeExec, ", sent: ", cntSucc, ", wait: ", int(cntAll-cntSucc), ", count: ", cntAll, ", instances: ", instOfSenders)
			exitCounter--
			if exitCounter < 1 {
				break
			}
		} else {
			break
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
	timeNow = time.Now()
	timeExec = int64(timeNow.Unix() - timeStart)
	printAll("Finish for ", prm.StateCode, "/", prm.StateName, "/", prm.Country, " > Time: ", timeExec, "sec, sent: ", cntSucc, ", failed: ", int(cntAll-cntSucc), ", count: ", cntAll)
	printAll("Bye!")
}
