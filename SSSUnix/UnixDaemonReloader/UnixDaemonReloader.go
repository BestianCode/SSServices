package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	//"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/BestianRU/SSServices/SSModules/sscfg"
	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sspid"
	"github.com/BestianRU/SSServices/SSModules/sssql"
	"github.com/BestianRU/SSServices/SSModules/sssys"
)

type aList struct {
	file, action, checkact, result, erroract string
}

var (
	dbase sssql.USQL
)

func readDIR(dir, exc, inc string) []string {
	var arr []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading %s dir: %v\n", dir, err)
		return nil
	}
	for _, f := range files {
		if f.IsDir() {
			for _, a := range readDIR(dir+"/"+f.Name(), exc, inc) {
				arr = append(arr, a)
			}
		} else {
			if inc != "" {
				for _, b := range strings.Split(inc, ",") {
					b = "^" + b + "$"
					b = strings.Replace(b, "*$", "*", -1)
					b = strings.Replace(b, "^*", "*", -1)
					b = strings.Replace(b, "*", ".*", -1)
					findRegExp := regexp.MustCompile(b)
					if findRegExp.FindString(f.Name()) != "" {
						arr = append(arr, fmt.Sprintf("%s/%s", dir, f.Name()))
					}
				}
			} else {
				if exc != "" {
					exc_st := 0
					for _, b := range strings.Split(exc, ",") {
						b = "^" + b + "$"
						b = strings.Replace(b, "*$", "*", -1)
						b = strings.Replace(b, "^*", "*", -1)
						b = strings.Replace(b, "*", ".*", -1)
						findRegExp := regexp.MustCompile(b)
						if findRegExp.FindString(f.Name()) != "" {
							exc_st = 1
						}
					}
					if exc_st == 0 {
						arr = append(arr, fmt.Sprintf("%s/%s", dir, f.Name()))
					}
				} else {
					arr = append(arr, fmt.Sprintf("%s/%s", dir, f.Name()))
				}
			}
		}
	}
	return arr
}

func getFullFileList(conf sscfg.ReadJSONConfig, rLog sslog.LogFile) []aList {
	var (
		arr []aList
		tmp aList
	)
	for _, y := range conf.Conf.UDR_WatchList {
		if y[1] == "*" {
			for _, a := range readDIR(y[0], "", "") {
				tmp.file = a
				tmp.action = y[2]
				if len(y) > 5 {
					tmp.checkact = y[3]
					tmp.result = y[4]
					tmp.erroract = y[5]
				} else {
					tmp.checkact = ""
					tmp.result = ""
					tmp.erroract = ""
				}
				arr = append(arr, tmp)
			}
		} else {
			if strings.Contains(y[1], "!") {
				for _, a := range readDIR(y[0], strings.Replace(y[1], "!", "", -1), "") {
					tmp.file = a
					tmp.action = y[2]
					if len(y) > 5 {
						tmp.checkact = y[3]
						tmp.result = y[4]
						tmp.erroract = y[5]
					} else {
						tmp.checkact = ""
						tmp.result = ""
						tmp.erroract = ""
					}
					arr = append(arr, tmp)
				}
			} else {
				for _, a := range readDIR(y[0], "", y[1]) {
					tmp.file = a
					tmp.action = y[2]
					if len(y) > 5 {
						tmp.checkact = y[3]
						tmp.result = y[4]
						tmp.erroract = y[5]
					} else {
						tmp.checkact = ""
						tmp.result = ""
						tmp.erroract = ""
					}
					arr = append(arr, tmp)
				}
			}
		}
	}
	return arr
}

func getMD5ofFile(file string) string {
	buff, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("Error reading %s file: %s\n", file, err)
		return "error"
	}
	return fmt.Sprintf("%x", md5.Sum(buff))
}

func checkFCS(conf sscfg.ReadJSONConfig, rLog sslog.LogFile) int {
	var (
		get    string
		result = int(0)
		a      aList
	)

	for _, a = range getFullFileList(conf, rLog) {
		summ := getMD5ofFile(a.file)
		if summ != "error" {
			res, err := dbase.D.Query(fmt.Sprintf("select md5s from files where fname='%s';", a.file))
			if err != nil {
				log.Println("SQLite: Query() select error: %v\n", err)
				return -1
			}
			res.Next()
			get = ""
			res.Scan(&get)
			if get == "" {
				_, err := dbase.D.Exec(fmt.Sprintf("insert into files (fname, md5s) values('%s','%s');", a.file, summ))
				if err != nil {
					log.Println("SQLite: Exec() insert error: %v\n", err)
					return -1
				}
				/*_, err = dbase.D.Exec(fmt.Sprintf("insert into go (action, checkact, result) select '%s','%s','%s' where '%s' not in (select action from go where action='%s');", a.action, a.checkact, a.result, a.action, a.action))
				if err != nil {
					log.Println("SQLite: Exec() ACT insert wI error: %v\n", err)
					return -1
				}*/
				rLog.Log("--> New file: ", a.file, " act: ", a.action)
				result = 1
			} else {
				if get != summ {
					_, err := dbase.D.Exec(fmt.Sprintf("update files set md5s='%s' where fname='%s';", summ, a.file))
					if err != nil {
						log.Println("SQLite: Exec() update error: %v\n", err)
						return -1
					}
					_, err = dbase.D.Exec(fmt.Sprintf("insert into go (action, checkact, result, erroract) select '%s','%s','%s','%s' where '%s' not in (select action from go where action='%s');", a.action, a.checkact, a.result, a.erroract, a.action, a.action))
					if err != nil {
						log.Println("SQLite: Exec() ACT insert wU error: %v\n", err)
						return -1
					}
					rLog.Log("--> File changed: ", a.file, " act: ", a.action)
					result = 1
				}
			}
		}
	}
	return result
}

func cleanFCSdb(conf sscfg.ReadJSONConfig, rLog sslog.LogFile) int {
	var (
		a aList
	)

	_, err := dbase.D.Exec(`CREATE temp TABLE IF NOT EXISTS files_cache (
				fname varchar(255));`)
	if err != nil {
		log.Println("SQLite: Exec() insert error: %v\n", err)
		return -1
	}

	for _, a = range getFullFileList(conf, rLog) {
		_, err = dbase.D.Exec(fmt.Sprintf("insert into files_cache (fname) values('%s');", a.file))
		if err != nil {
			log.Println("SQLite: Exec() insert error: %v\n", err)
			return -1
		}
	}
	_, err = dbase.D.Exec("delete from files where fname not in (select fname from files_cache);")
	if err != nil {
		log.Println("SQLite: Exec() insert error: %v\n", err)
		return -1
	}
	_, err = dbase.D.Exec("drop TABLE files_cache;")
	if err != nil {
		log.Println("SQLite: Exec() insert error: %v\n", err)
		return -1
	}

	rLog.Log("--> Clean DB")

	return 1
}

func startCommand(conf sscfg.ReadJSONConfig, get string) {
	err := exec.Command(conf.Conf.UDR_Shell, conf.Conf.UDR_ShellExecParam, get+" &").Start()
	if err != nil {
		log.Println("\tExec error: ", err)
	}
}

func startCommandPreCheck(conf sscfg.ReadJSONConfig, getact, getcheck, getresult, geterror string) {
	var (
		rLog sslog.LogFile
		i    int
	)
	rLog.ON(conf.Conf.LOG_File, conf.Conf.LOG_Level)
	defer rLog.OFF()
	if getcheck != "" && getresult != "" {
		if conf.Conf.UDR_PauseBefore > 0 {
			rLog.Log("/|\\ Sleep ", conf.Conf.UDR_PauseBefore, " sec. before start PRE-APP Script: ", getcheck, " before ", getact)
			time.Sleep(time.Duration(conf.Conf.UDR_PauseBefore) * time.Second)
		}
		for i = 1; i <= conf.Conf.UDR_PreAppAttempt; i++ {
			rLog.Log("--> Start PRE-APP: ", getcheck, " before ", getact, " Attempt N ", i)
			cmd := exec.Command(conf.Conf.UDR_Shell, conf.Conf.UDR_ShellExecParam, conf.Conf.UDR_ScriptsPath+"/"+getcheck)
			read_result, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("\tExec PRE-APP ", getcheck, " error: ", err)
			}
			xRes := strings.Replace(fmt.Sprintf("%s", read_result), "\n", "", -1)
			if xRes == getresult {
				rLog.Log("--> OK! PRE-APP script ", getcheck, " return ", xRes)
				rLog.Log("--> Start: ", getact)
				startCommand(conf, getact)
				break
			}
			rLog.Log("--> Error. PRE-APP script ", getcheck, " return ", xRes, " instead of ", getresult)
			if i >= conf.Conf.UDR_PreAppAttempt {
				break
			}
			rLog.Log("/|\\ Sleep ", conf.Conf.Sleep_Time, " sec. before attempt N ", i+1, " to start PRE-APP Script: ", getcheck, " before ", getact)
			time.Sleep(time.Duration(conf.Conf.Sleep_Time) * time.Second)
		}
		if i >= conf.Conf.UDR_PreAppAttempt {
			rLog.Log("**> Error starting PRE-APP Script: ", getcheck)
			if geterror != "" {
				rLog.Log("--> Start: ", geterror)
				cmd := exec.Command(conf.Conf.UDR_Shell, conf.Conf.UDR_ShellExecParam, conf.Conf.UDR_ScriptsPath+"/"+geterror)
				read_result, err := cmd.CombinedOutput()
				if err != nil {
					log.Println("\tExec ERROR Script ", geterror, " error: ", err)
				} else {
					rLog.Log("--> OK! ERROR script ", geterror, " return ", fmt.Sprintf("%s", read_result))
				}
			}
		}
	} else {
		if conf.Conf.UDR_PauseBefore > 0 {
			rLog.Log("/|\\ Sleep ", conf.Conf.UDR_PauseBefore, " sec. before start: ", getact)
			time.Sleep(time.Duration(conf.Conf.UDR_PauseBefore) * time.Second)
		}
		rLog.Log("--> Start: ", getact)
		startCommand(conf, getact)
	}
}

func restartFCS(conf sscfg.ReadJSONConfig, rLog sslog.LogFile) int {
	var getact, getcheck, getresult, geterror string
	res, err := dbase.D.Query("select action, checkact, result, erroract from go;")
	if err != nil {
		log.Println("SQLite: Query() select actions error: %v\n", err)
		return -1
	}
	for res.Next() {
		res.Scan(&getact, &getcheck, &getresult, &geterror)
		go startCommandPreCheck(conf, getact, getcheck, getresult, geterror)
	}
	return 1
}

func main() {
	var (
		jsonConfig sscfg.ReadJSONConfig
		rLog       sslog.LogFile
		pid        sspid.PidFile
		sleepWatch = int(0)
	)

	const (
		pName       = string("SSServices / UnixDaemonReloader")
		pVer        = string("2 2016.01.03.02.30")
		initDBQuery = string(`PRAGMA journal_mode=WAL;
			CREATE TABLE IF NOT EXISTS files (
				fname varchar(255) PRIMARY KEY,
				md5s varchar(255)
			);
			CREATE TABLE IF NOT EXISTS go (
				action varchar(255) PRIMARY KEY,
				checkact varchar(255),
				erroract varchar(255),
				result varchar(255)
			);
			delete from go;
		`)
	)

	fmt.Printf("\n\t%s V%s\n\n", pName, pVer)

	jsonConfig.Init("./UnixDaemonReloader.log", "./UnixDaemonReloader.json")
	/*
		for _, xx := range jsonConfig.Conf.UDR_WatchList {
			for _, yy := range xx {
				fmt.Printf("%v / ", yy)
			}
			fmt.Println("")
		}

		os.Exit(0)
	*/
	rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
	pid.ON(jsonConfig.Conf.PID_File)
	pid.OFF()
	rLog.OFF()

	sssys.Fork(jsonConfig.Daemon_mode, jsonConfig.Config_file)
	sssys.Signal(pid, jsonConfig.Conf.LOG_File)

	rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
	pid.ON(jsonConfig.Conf.PID_File)
	defer pid.OFF()
	rLog.Hello(pName, pVer)
	rLog.OFF()

	dbase.Init("SQLite", jsonConfig.Conf.SQLite_DB, "DROP TABLE IF EXISTS go;")
	dbase.Close()

	for {
		rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
		jsonConfig.Update()

		dbase.Init("SQLite", jsonConfig.Conf.SQLite_DB, initDBQuery)

		if checkFCS(jsonConfig, rLog) > 0 {
			restartFCS(jsonConfig, rLog)
			cleanFCSdb(jsonConfig, rLog)
			sleepWatch = 0
		}

		dbase.Close()

		if sleepWatch > 3600 {
			rLog.Log("<-- I'm alive ... :)")
			sleepWatch = 0
		}

		rLog.OFF()
		time.Sleep(time.Duration(jsonConfig.Conf.Sleep_Time) * time.Second)
		sleepWatch += jsonConfig.Conf.Sleep_Time
	}
}
