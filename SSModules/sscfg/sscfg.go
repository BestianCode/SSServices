package sscfg

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type ReadJSONConfig struct {
	Silent      string
	Config_file string
	Daemon_mode string
	Phase       string
	Keys        string
	Conf        struct {
		PG_DSN              string
		MY_DSN              string
		SQLite_DB           string
		AST_SQLite_DB       string
		AST_CID_Group       string
		AST_Num_Start       string
		AST_ARI_Host        string
		AST_ARI_Port        int
		AST_ARI_User        string
		AST_ARI_Pass        string
		Oracle_SRV          [][]string
		MSSQL_DSN           [][]string
		LDAP_URL            [][]string
		ROOT_OU             string
		ROOT_DN             [][]string
		Sleep_Time          int
		Sleep_cycles        int
		LOG_File            string
		LOG_Level           int
		PID_File            string
		TRANS_NAMES         [][]string
		BlackList_OU        []string
		WLB_SessTimeOut     int
		WLB_Listen_IP       string
		WLB_Listen_PORT     int
		WLB_LDAP_ATTR       [][]string
		WLB_SQL_PreFetch    string
		WLB_MailBT          string
		WLB_HTML_Title      string
		AD_LDAP             [][]string
		AD_ScriptDir        string
		AD_LDAP_PARENT      [][]string
		TRANS_POS           [][]string
		SABRealm            string
		WLB_DavDNTreeDepLev int
		UDR_WatchList       [][]string
		UDR_Shell           string
		UDR_ShellExecParam  string
		UDR_ScriptsPath     string
		UDR_PreAppAttempt   int
		UDR_SleepCommand    string
		UDR_PauseBefore     int
		SQH_LogPasswords    string
		SQH_SQL_UserCheck   string
		SQH_SQL_PassCheck   string
		SQH_SQL_IPUpdate    string
		SQH_LDAP_URL        [][]string
		SQH_AD_GroupMember  string
		CardDAVIPSuffix     []string
		SQL_Engine          string
		SQL_QUE1            string
		SQL_QUE2            string
		SQL_QUE3            string
		E2O_JDomain         string
		E2O_passwd          string
		E2O_roster          string
		E2O_vcard           string
		E2O_OutXML          string
		E2O_VCD_Engine      string
		E2O_VCD_PG_DSN      string
		E2O_VCD_MY_DSN      string
		E2O_Name_Update     string
		E2O_Name_PG_DSN     string
		BMDS_TitleDir       string
		BMDS_BodyDir        string
		BMDS_SenderName     string
		BMDS_MaxInstances   int
	}
}

func (_s *ReadJSONConfig) Init(defLog, defCFG string) {
	if defLog == "" {
		_s.Conf.LOG_File = "./AmnesiacDefault.log" // Default log file
	} else {
		_s.Conf.LOG_File = defLog
	}
	if defCFG == "" {
		_s.Config_file = "./AmnesiacDefault.json" // Default configuration file
	} else {
		_s.Config_file = defCFG
	}
	_s.Daemon_mode = "NO" // Default start in foreground

	_s._parseCommandLine()
	_s._readConfigFile()

	if _s.Silent != "silent" {
		fmt.Printf("Configuration file: %s\n", _s.Config_file)
		fmt.Printf("          Log file: %s\n", _s.Conf.LOG_File)
		fmt.Printf("          PID file: %s\n", _s.Conf.PID_File)
		fmt.Printf("       Daemon mode: %s\n", _s.Daemon_mode)
		fmt.Printf("\n")
		fmt.Printf("\n")
	}
}

func (_s *ReadJSONConfig) Update() {
	_s._readConfigFile()
}

func (_s *ReadJSONConfig) _parseCommandLine() {
	cp := flag.String("config", _s.Config_file, "Path to Configuration file")
	dp := flag.String("daemon", _s.Daemon_mode, "Fork as system daemon (YES or NO)")
	pp := flag.String("phase", _s.Phase, "select work phase (1,2,3)")
	ep := flag.String("keys", _s.Keys, "\"[State Code (wordpress)] [State name] [State short name] [Country] [From mail] [test/r]\"")
	flag.Parse()
	_s.Config_file = *cp
	_s.Daemon_mode = *dp
	_s.Phase = *pp
	_s.Keys = *ep
	//fmt.Println(*cp, "\n", *dp, "\n", os.Args, "\n")
}

func (_s *ReadJSONConfig) _readConfigFile() {
	f, err := os.Open(_s.Config_file)
	if err != nil {
		fmt.Printf("Error open Configuration file: %s (%v)\n", _s.Config_file, err)
		os.Exit(1)
	}

	c := json.NewDecoder(f)
	err = c.Decode(&_s.Conf)
	if err != nil {
		fmt.Printf("Error read Configuration file: %s (%v)\n", _s.Config_file, err)
		os.Exit(2)
	}
	f.Close()
}
