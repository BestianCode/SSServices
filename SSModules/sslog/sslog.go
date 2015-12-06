package sslog

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	LLError = iota
	LLWarning
	LLInfo
	LLTrace
)

type LogFile struct {
	LL       int
	flog     *os.File
	lineLog  *log.Logger
	loglevel int
}

func (_s *LogFile) ON(xlog_file string, xlog_level int) {
	var err error

	_s.flog, err = os.OpenFile(xlog_file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error open log file: %s (%v)\n", xlog_file, err)
	}

	if xlog_level > 0 && xlog_level <= 3 {
		_s.LL = xlog_level
	} else {
		_s.LL = 0
	}

	_s.lineLog = log.New(_s.flog, "", log.Ldate|log.Ltime)

	//log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(_s.flog)
	_s.loglevel = xlog_level
}

func (_s *LogFile) OFF() {
	var err error
	log.SetOutput(os.Stdout)
	_s.lineLog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	err = _s.flog.Close()
	if err != nil {
		log.Fatalf("Error close log file: (%v)\n", err)
	}
}

func (_s *LogFile) Log(msg ...interface{}) {
	var message = string("")
	for _, x := range msg {
		message = fmt.Sprintf("%s%v", message, x)
	}
	message = strings.Trim(message, " ")

	_s.lineLog.Printf("%v", message)
}

func (_s *LogFile) LogDbg(ll int, msg ...interface{}) {
	var message = string("")
	for _, x := range msg {
		message = fmt.Sprintf("%s%v", message, x)
	}
	message = strings.Trim(message, " ")

	pc, _, line, _ := runtime.Caller(1)
	func_and_line := runtime.FuncForPC(pc).Name() + ":" + strconv.Itoa(line)
	tmestamp := time.Now().Format(time.StampMilli)

	if ll <= _s.loglevel {
		switch ll {
		case LLError:
			log.Println(tmestamp, "[Error]", func_and_line, message)
		case LLWarning:
			//log.Println(tmestamp, "[Warning]", func_and_line, message)
			log.Println(tmestamp, "[Warning]", message)
		case LLInfo:
			//log.Println(tmestamp, "[Info]", func_and_line, message)
			log.Println(tmestamp, "[Info]", message)
		case LLTrace:
			log.Println(tmestamp, "[Trace]", func_and_line, message)
		}
	}
}

func (_s *LogFile) Hello(pName, pVer string) {
	_s.lineLog.Printf(".")
	_s.lineLog.Printf(">")
	_s.lineLog.Printf("-> %s V%s", pName, pVer)
	_s.lineLog.Printf("--> Go!")
}

func (_s *LogFile) Bye() {
	_s.lineLog.Printf("--> To Sleep...")
	_s.lineLog.Printf("->")
	_s.lineLog.Printf(">")
	_s.lineLog.Printf(".")
}
