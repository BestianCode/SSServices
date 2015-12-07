package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/BestianRU/SSServices/SSModules/sscfg"
	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sspid"
	"github.com/BestianRU/SSServices/SSModules/sssql"
	"github.com/BestianRU/SSServices/SSModules/sssys"
)

var (
	multiInsert int
	db_init     = string(`
				CREATE TABLE IF NOT EXISTS users (
					username varchar(255),
					password varchar(255)
				);
				CREATE TABLE IF NOT EXISTS rosterusers (
					username varchar(255),
					jid varchar(255),
					nick varchar(255),
					subscription varchar(255),
					ask varchar(255),
					askmessage varchar(255),
					server varchar(255),
					subscribe varchar(255),
					type varchar(255)
				);
				CREATE TABLE IF NOT EXISTS rostergroups (
	            	username varchar(255),
	            	jid varchar(255),
	            	grp varchar(255)
	            );
				CREATE TABLE IF NOT EXISTS vcard (
					username varchar(255),
					vcard varchar(102400)
				);
				CREATE TABLE IF NOT EXISTS nick (
					jid varchar(255),
					nick varchar(255),
					fullname varchar(255)
				);
				delete from users;
				delete from rosterusers;
				delete from rostergroups;
				delete from vcard;
				delete from nick;
			`)
)

func selectInserts(dbase *sql.DB, rLog sslog.LogFile, file string) int {
	var (
		y         = string("")
		xrune     rune
		xwidth    int
		query     = string("")
		count     = int(0)
		err       error
		xlen_full int
		xlen      int
	)
	insRegExp := regexp.MustCompile(`^insert`)
	latinRegExp := regexp.MustCompile(`[A-Za-z\,\.\(\)\;0-9\_\ \@\'\"'\=\-\!\\\<\>\/\+\&\*\^]`)
	buff, err := ioutil.ReadFile(file)
	if err != nil {
		rLog.LogDbg(0, "Error reading ", file, " file: ", file, err)
		return -1
	}
	rLog.Log("Parse ", file)
	xlen_full = len(buff)
	for len(buff) > 0 {
		xlen = len(buff)
		if latinRegExp.FindStringIndex(string(buff[0])) == nil {
			xrune, xwidth = utf8.DecodeRuneInString(string(buff))
			if string(buff[0]) == string(xrune) {
				rLog.LogDbg(3, "RegExp need to update:", string(xrune))
			}
		} else {
			xrune = rune(buff[0])
			xwidth = 1
		}
		y = y + string(xrune)
		buff = buff[xwidth:]
		if (xrune == ';' && insRegExp.FindStringIndex(y) == nil) || (xrune == ';' && insRegExp.FindStringIndex(y) != nil && string(y[len(y)-2]) == ")") {
			if insRegExp.FindStringIndex(y) != nil {
				query = query + y + " "
				count++
				if count > multiInsert {
					_, err = dbase.Exec(query)
					if err != nil {
						rLog.LogDbg(0, "SQL: Exec() insert error: ", err)
						rLog.LogDbg(0, query)
						//return -1
					}
					query = ""
					count = 0
				}
			}
			y = ""
		}
		if int(xlen/10000)*10000 == xlen {
			rLog.Log(file, " - Parsed ", xlen_full-xlen, " bytes of ", xlen_full-1)
		}
	}
	_, err = dbase.Exec(query)
	if err != nil {
		rLog.LogDbg(0, "SQL: Exec() insert error: ", err)
		rLog.LogDbg(0, query)
		//return -1
	}
	rLog.Log(file, " - Parsed ", xlen_full-xlen, " bytes of ", xlen_full-1)
	rLog.Log(file, " - Complete!")
	return 0
}

// ------------------------
// NICK Update module START
// ------------------------
func nickUpdate(dbase *sql.DB, dbName *sql.DB, rLog sslog.LogFile) int {
	var (
		err   error
		jid   string
		query string
		name  string
		namec int
		ckl   int
	)
	rowsa, err := dbase.Query("SELECT distinct jid from rosterusers group by jid, nick order by jid")
	if err != nil {
		rLog.LogDbg(0, "SQL: Query() select error: ", err)
		return -1
	}
	for rowsa.Next() {
		rowsa.Scan(&jid)
		x := strings.Split(jid, "@")
		if len(x) == 2 {
			//if strings.ToLower(jid) == strings.ToLower(nick) || strings.ToLower(x[0]) == strings.ToLower(nick) || len(nick) < 2 {
			query = "select y.cid_name from ldapx_persons as x, ldapx_persons as y where x.lang=1 and lower(replace(replace(x.cid_name, '.', ''), ' ', '_'))=lower('" + x[0] + "') and x.uid=y.uid and y.lang=0 limit 1;"
			rowsb, err := dbName.Query(query)
			if err != nil {
				rLog.LogDbg(0, "SQL: Query() select error: ", err)
				rLog.LogDbg(0, query)
				return -1
			}
			rowsb.Next()
			name = ""
			rowsb.Scan(&name)
			rowsb.Close()
			if len(name) > 2 {
				//fmt.Printf("|%s@%s|%s|\n", x[0], x[1], name)
				query = "insert into nick (jid,nick) select '" + jid + "','" + name + "' where '" + jid + "' not in (select jid from nick where jid='" + jid + "');"
				rowsb, err := dbase.Query(query)
				if err != nil {
					rLog.LogDbg(0, "SQL: Query() insert nick error: ", err)
					rLog.LogDbg(0, query)
					return -1
				}
				rowsb.Close()
				query = "select count(cid_name) from ldapx_persons where lang=1 and lower(replace(replace(cid_name, '.', ''), ' ', '_'))=lower('" + x[0] + "') group by cid_name;"
				rowsb, err = dbName.Query(query)
				if err != nil {
					rLog.LogDbg(0, "SQL: Query() select error: ", err)
					rLog.LogDbg(0, query)
					return -1
				}
				rowsb.Next()
				namec = 0
				rowsb.Scan(&namec)
				rowsb.Close()
				if namec == 1 {
					query = "select y.fullname from ldapx_persons as x, ldapx_persons as y where x.lang=1 and lower(replace(replace(x.cid_name, '.', ''), ' ', '_'))=lower('" + x[0] + "') and x.uid=y.uid and y.lang=0 limit 1;"
					rowsb, err = dbName.Query(query)
					if err != nil {
						rLog.LogDbg(0, "SQL: Query() select error: ", err)
						rLog.LogDbg(0, query)
						return -1
					}
					rowsb.Next()
					name = ""
					rowsb.Scan(&name)
					rowsb.Close()
					if len(name) > 2 {
						query = "update nick set fullname='" + name + "' where jid='" + jid + "';"
						rowsb, err := dbase.Query(query)
						if err != nil {
							rLog.LogDbg(0, "SQL: Query() insert nick error: ", err)
							rLog.LogDbg(0, query)
							return -1
						}
						rowsb.Close()
					}
				}
				ckl++
				if int(ckl/100)*100 == ckl {
					rLog.Log("Updated nicks: ", ckl)
				}
			}
		}
	}
	rowsa.Close()
	rLog.Log("Updated nicks: ", ckl)
	return 0
}

// ------------------------
// NICK Update module   END
// ------------------------

func makeXML(dbase *sql.DB, rLog sslog.LogFile, conf sscfg.ReadJSONConfig) int {
	var (
		uname        string
		upass        string
		jid          string
		nick         string
		ask          string
		subscription string
		ufn          string
		grp          string
		ckl          int
		sql_nick     string
	)

	fnsRegExp := regexp.MustCompile(`.*\<FN\>`)
	fneRegExp := regexp.MustCompile(`\<\/FN\>.*`)
	fnnRegExp := regexp.MustCompile(`\<FN\/\>`)

	fout, err := os.OpenFile(conf.Conf.E2O_OutXML, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		rLog.LogDbg(0, "Error create OutPut XML file: ", conf.Conf.E2O_OutXML, " error: ", err)
		return -1
	}

	fout.WriteString(fmt.Sprintln("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"))
	fout.WriteString(fmt.Sprintln("<Openfire>"))

	_, err = dbase.Query("insert into vcard SELECT x.username, '' FROM users as x WHERE x.username not in (select username from vcard where username=x.username)")
	if err != nil {
		rLog.LogDbg(0, "SQL: Query() select error: ", err)
		return -1
	}
	rowsa, err := dbase.Query("SELECT y.username,y.password,z.vcard FROM users as y, vcard as z WHERE y.username=z.username")
	if err != nil {
		rLog.LogDbg(0, "SQL: Query() select error: ", err)
		return -1
	}
	for rowsa.Next() {
		rowsa.Scan(&uname, &upass, &ufn)
		if len(ufn) > 10 {
			if fnnRegExp.FindStringIndex(ufn) == nil && fnsRegExp.FindStringIndex(ufn) != nil && fneRegExp.FindStringIndex(ufn) != nil {
				ufn = fnsRegExp.ReplaceAllString(ufn, "")
				ufn = fneRegExp.ReplaceAllString(ufn, "")
			} else {
				ufn = uname
			}
		} else {
			ufn = uname
		}
		if strings.ToLower(conf.Conf.E2O_Name_Update) == "yes" {
			// ------------------------
			// NICK Update module START
			// ------------------------
			rowsfn, err := dbase.Query("SELECT nick FROM nick WHERE lower(jid) like lower('" + uname + "@%') limit 1;")
			if err != nil {
				rLog.LogDbg(0, "SQL: Query() select nick error: ", err)
				return -1
			}
			rowsfn.Next()
			sql_nick = ""
			rowsfn.Scan(&sql_nick)
			rowsfn.Close()
			if len(ufn) <= len(sql_nick) {
				ufn = sql_nick
			}
			// ------------------------
			// NICK Update module   END
			// ------------------------
		}
		//if strings.ToLower(jid) == strings.ToLower(nick) || strings.ToLower(x[0]) == strings.ToLower(nick) || len(nick) < 2 {
		fout.WriteString(fmt.Sprintf("\t<User>\n"))
		fout.WriteString(fmt.Sprintf("\t\t<Username>%s</Username>\n", uname))
		fout.WriteString(fmt.Sprintf("\t\t<Password>%s</Password>\n", upass))
		fout.WriteString(fmt.Sprintf("\t\t<Email></Email>\n"))
		fout.WriteString(fmt.Sprintf("\t\t<Name>%s</Name>\n", ufn))
		fout.WriteString(fmt.Sprintf("\t\t<CreationDate></CreationDate>\n"))
		fout.WriteString(fmt.Sprintf("\t\t<ModifiedDate></ModifiedDate>\n"))
		fout.WriteString(fmt.Sprintf("\t\t<Roster>\n"))
		rowsr, err := dbase.Query("SELECT jid,nick,ask,subscription FROM rosterusers WHERE username='" + uname + "'")
		if err != nil {
			rLog.LogDbg(0, "SQL: Query() select error: ", err)
			return -1
		}
		for rowsr.Next() {
			rowsr.Scan(&jid, &nick, &ask, &subscription)
			if ask == "N" && subscription == "B" {
				xjid := strings.Split(jid, "@")
				if len(xjid) == 2 {
					rowsjc, err := dbase.Query("SELECT username FROM users WHERE username='" + xjid[0] + "' limit 1;")
					if err != nil {
						rLog.LogDbg(0, "SQL: Query() select nick error: ", err)
						return -1
					}
					rowsjc.Next()
					xjid[1] = ""
					rowsjc.Scan(&xjid[1])
					rowsjc.Close()
					if xjid[1] == xjid[0] || !strings.Contains(jid, conf.Conf.E2O_JDomain) {
						if strings.ToLower(conf.Conf.E2O_Name_Update) == "yes" {
							// ------------------------
							// NICK Update module START
							// ------------------------
							rowsfn, err := dbase.Query("SELECT nick FROM nick WHERE lower(jid)=lower('" + jid + "') limit 1;")
							if err != nil {
								rLog.LogDbg(0, "SQL: Query() select nick error: ", err)
								return -1
							}
							rowsfn.Next()
							sql_nick = ""
							rowsfn.Scan(&sql_nick)
							rowsfn.Close()
							if len(sql_nick) > 2 {
								nick = sql_nick
							}
							// ------------------------
							// NICK Update module   END
							// ------------------------
						}
						fout.WriteString(fmt.Sprintf("\t\t\t<Item jid=\"%s\" askstatus=\"-1\" recvstatus=\"-1\" substatus=\"3\" name=\"%s\">\n", jid, nick))
						rowsg, err := dbase.Query("SELECT grp FROM rostergroups WHERE username='" + uname + "' AND jid='" + jid + "'")
						rowsg.Next()
						grp = ""
						rowsg.Scan(&grp)
						rowsg.Close()
						if err != nil {
							rLog.LogDbg(0, "SQL: Query() select error: ", err)
							return -1
						}
						if len(grp) > 1 {
							fout.WriteString(fmt.Sprintf("\t\t\t\t<Group>%s</Group>\n", grp))
						}
						fout.WriteString(fmt.Sprintf("\t\t\t</Item>\n"))
					}
				}
			}
		}
		rowsr.Close()
		fout.WriteString(fmt.Sprintf("\t\t</Roster>\n"))
		fout.WriteString(fmt.Sprintf("\t</User>\n"))
		ckl++
		if int(ckl/100)*100 == ckl {
			rLog.Log(conf.Conf.E2O_OutXML, " - recorded ", ckl, " Users with rosters...")
		}
	}
	rowsa.Close()
	fout.WriteString(fmt.Sprintln("</Openfire>"))

	fout.Close()
	rLog.Log(conf.Conf.E2O_OutXML, " - recorded ", ckl, " Users with rosters...")
	rLog.Log(conf.Conf.E2O_OutXML, " - Complete!")

	return 0
}

func vCardUpdate(dbf, dbt *sql.DB, rLog sslog.LogFile, conf sscfg.ReadJSONConfig) int {
	var (
		uname  string
		vcard  string
		ckl    int
		nick   string
		fname  string
		wrname string
	)

	fnRegExp := regexp.MustCompile(`\<FN\>.*\<\/FN\>|\<\/FN\>`)
	nickRegExp := regexp.MustCompile(`\<NICKNAME\>.*\<\/NICKNAME\>`)
	endRegExp := regexp.MustCompile(`\<\/vCard\>`)
	//rusRegExp := regexp.MustCompile(`[А-Яа-я]`)

	rowsa, err := dbf.Query("SELECT username,vcard FROM vcard")
	if err != nil {
		rLog.LogDbg(0, "SQL: Query() select error: ", err)
		return -1
	}
	for rowsa.Next() {
		rowsa.Scan(&uname, &vcard)
		vcard = strings.Replace(vcard, "\\n", "\n", -1)
		vcard = strings.Replace(vcard, "'", "\"", -1)

		if strings.ToLower(conf.Conf.E2O_Name_Update) == "yes" {
			// ------------------------
			// NICK Update module START
			// ------------------------
			query := "SELECT nick,fullname FROM nick where lower(jid) like lower('" + uname + "@%');"
			//rLog.LogDbg(3, query)
			rowsfn, err := dbf.Query(query)
			if err != nil {
				rLog.LogDbg(0, "SQL: Query() select error: ", err)
				rLog.LogDbg(0, query)
				return -1
			}
			rowsfn.Next()
			nick = ""
			fname = ""
			rowsfn.Scan(&nick, &fname)
			rowsfn.Close()

			if len(nick) > 2 {
				if len(fname) > len(nick) {
					wrname = fname
				} else {
					wrname = nick
				}
				//rLog.LogDbg(3, vcard)
				if endRegExp.FindStringIndex(vcard) != nil {
					if fnRegExp.FindStringIndex(vcard) != nil {
						vcard = fnRegExp.ReplaceAllString(vcard, "<FN>"+wrname+"</FN>")
					} else {
						vcard = endRegExp.ReplaceAllString(vcard, "<FN>"+wrname+"</FN>\n</vCard>\n")
					}
					if nickRegExp.FindStringIndex(vcard) != nil {
						vcard = nickRegExp.ReplaceAllString(vcard, "<NICKNAME>"+nick+"</NICKNAME>")
					} else {
						vcard = endRegExp.ReplaceAllString(vcard, "<NICKNAME>"+nick+"</NICKNAME>\n</vCard>\n")
					}
				} else {
					vcard = "<vCard xmlns=\"vcard-temp\" prodid=\"-//HandGen//NONSGML vGen v1.0//EN\" version=\"2.0\" xdbns=\"vcard-temp\">\n<FN>" + wrname + "</FN>\n<NICKNAME>" + nick + "</NICKNAME>\n</vCard>\n"
				}
				//rLog.LogDbg(3, vcard)
			}
			// ------------------------
			// NICK Update module   END
			// ------------------------
		}

		if len(vcard) > 10 {
			query := "delete from ofvcard where username='" + uname + "'; insert into ofvcard (username,vcard) values ('" + uname + "','" + vcard + "');"
			rowsb, err := dbt.Query(query)
			if err != nil {
				rLog.LogDbg(0, "SQL: Query() VCard insert error: ", err)
				rLog.LogDbg(0, query)
				return -1
			}
			rowsb.Close()
		}
		ckl++
		if int(ckl/100)*100 == ckl {
			rLog.Log("VCard update: ", ckl)
		}
	}
	rLog.Log("VCard update: ", ckl)
	rLog.Log("VCard update - Complete!")

	rowsa.Close()
	return 0
}

func queryWrapper(dbase *sql.DB, rLog sslog.LogFile, conf sscfg.ReadJSONConfig, phase int) {
	var (
		result = []string{"Error", "OK"}
		dDP    sssql.USQL
	)
	if phase == 1 {
		rLog.Log("Parse passwd: ", result[selectInserts(dbase, rLog, conf.Conf.E2O_passwd)+1])
		rLog.Log("Parse roster: ", result[selectInserts(dbase, rLog, conf.Conf.E2O_roster)+1])
		rLog.Log("Parse vcard:  ", result[selectInserts(dbase, rLog, conf.Conf.E2O_vcard)+1])
	}
	if phase == 2 && strings.ToLower(conf.Conf.E2O_Name_Update) == "yes" {
		// ------------------------
		// NICK Update module START
		// ------------------------
		var nDP sssql.USQL
		multiInsert = 50
		nDP.Init("PG", conf.Conf.E2O_Name_PG_DSN, "")
		rLog.Log("RosterNameUpd:", result[nickUpdate(dbase, nDP.D, rLog)+1])
		nDP.Close()
		// ------------------------
		// NICK Update module   END
		// ------------------------
	}
	if phase == 3 {
		rLog.Log("Make XML:     ", result[makeXML(dbase, rLog, conf)+1])
	}
	if phase == 4 {
		switch strings.ToLower(conf.Conf.E2O_VCD_Engine) {
		case "pg":
			multiInsert = 50
			dDP.Init("PG", conf.Conf.E2O_VCD_PG_DSN, "")
		case "my":
			multiInsert = 50
			dDP.Init("MY", conf.Conf.E2O_VCD_MY_DSN, "")
		default:
			rLog.LogDbg(0, "SQL Engine for VCard update select error! (PG|MY|none)")
			return
		}
		rLog.Log("VCard Load:   ", result[vCardUpdate(dbase, dDP.D, rLog, conf)+1])
		dDP.Close()
	}
}

func main() {
	var (
		jsonConfig sscfg.ReadJSONConfig
		rLog       sslog.LogFile
		pid        sspid.PidFile
		dbase      sssql.USQL
		phase      int
	)

	const (
		pName = string("SSServices / EJabberd2OpenFire")
		pVer  = string("1 2015.12.08.01.00")
	)

	fmt.Printf("\n\t%s V%s\n\n", pName, pVer)

	jsonConfig.Init("./EJabberd2OpenFire.log", "./EJabberd2OpenFire.json")

	rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
	defer rLog.OFF()
	pid.ON(jsonConfig.Conf.PID_File)
	defer pid.OFF()
	rLog.Hello(pName, pVer)

	sssys.Signal(pid, jsonConfig.Conf.LOG_File)

	phase, _ = strconv.Atoi(jsonConfig.Phase)
	if phase > 4 || phase < 1 {
		fmt.Println("Phase select ERROR! (1,2,3)\n")
		fmt.Println("Select phase:\n")
		fmt.Println("1 - make SQL DB from EJabberd files")
		fmt.Println("2 - update nicks for rosters in SQL DB")
		fmt.Println("3 - make XML file for OpenFire import throught import/export plugin")
		fmt.Println("4 - update vCard profiles in OpenFire SQL DB\n")
		return
	}

	if phase != 1 {
		db_init = ""
	}

	switch strings.ToLower(jsonConfig.Conf.SQL_Engine) {
	case "pg":
		multiInsert = 50
		dbase.Init("PG", jsonConfig.Conf.PG_DSN, db_init)
	case "my":
		multiInsert = 50
		dbase.Init("MY", jsonConfig.Conf.MY_DSN, db_init)
	case "sqlite":
		multiInsert = 10
		if phase != 1 {
			dbase.Init("SQLite", jsonConfig.Conf.SQLite_DB, db_init)
		} else {
			dbase.Init("SQLite", jsonConfig.Conf.SQLite_DB, "PRAGMA journal_mode=WAL;"+db_init)
		}
	default:
		rLog.LogDbg(0, "SQL Engine select error! (PG|MY|SQLite|none)")
		return
	}
	queryWrapper(dbase.D, rLog, jsonConfig, phase)
	dbase.Close()
}
