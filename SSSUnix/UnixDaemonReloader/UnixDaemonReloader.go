package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
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
	file, action string
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
				arr = append(arr, tmp)
			}
		} else {
			if strings.Contains(y[1], "!") {
				for _, a := range readDIR(y[0], strings.Replace(y[1], "!", "", -1), "") {
					tmp.file = a
					tmp.action = y[2]
					arr = append(arr, tmp)
				}
			} else {
				for _, a := range readDIR(y[0], "", y[1]) {
					tmp.file = a
					tmp.action = y[2]
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
				_, err = dbase.D.Exec(fmt.Sprintf("insert into go (action) select '%s' where '%s' not in (select action from go where action='%s');", a.action, a.action, a.action))
				if err != nil {
					log.Println("SQLite: Exec() ACT insert wI error: %v\n", err)
					return -1
				}
				rLog.Log("--> New file: ", a.file, " act: ", a.action)
				result = 1
			} else {
				if get != summ {
					_, err := dbase.D.Exec(fmt.Sprintf("update files set md5s='%s' where fname='%s';", summ, a.file))
					if err != nil {
						log.Println("SQLite: Exec() update error: %v\n", err)
						return -1
					}
					_, err = dbase.D.Exec(fmt.Sprintf("insert into go (action) select '%s' where '%s' not in (select action from go where action='%s');", a.action, a.action, a.action))
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

func restartFCS(conf sscfg.ReadJSONConfig, rLog sslog.LogFile) int {
	var get string
	res, err := dbase.D.Query("select action from go;")
	if err != nil {
		log.Println("SQLite: Query() select actions error: %v\n", err)
		return -1
	}
	for res.Next() {
		res.Scan(&get)
		rLog.Log("--> Start: ", get)

		err := exec.Command(conf.Conf.UDR_Shell, conf.Conf.UDR_ShellExecParam, get+" &").Start()
		if err != nil {
			log.Println("\tExec error: %v\n", err)
		}
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
		pName = string("SSServices / UnixDaemonReloader")
		pVer  = string("1 2015.12.07.01.00")
	)

	fmt.Printf("\n\t%s V%s\n\n", pName, pVer)

	jsonConfig.Init("./UnixDaemonReloader.log", "./UnixDaemonReloader.json")

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

	for {
		rLog.ON(jsonConfig.Conf.LOG_File, jsonConfig.Conf.LOG_Level)
		jsonConfig.Update()

		dbase.Init("SQLite", jsonConfig.Conf.SQLite_DB, `PRAGMA journal_mode=WAL;
			CREATE TABLE IF NOT EXISTS files (
				fname varchar(255) PRIMARY KEY,
				md5s varchar(255)
			);
			CREATE TABLE IF NOT EXISTS go (
				action varchar(255) PRIMARY KEY
			);
			delete from go;
		`)

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
