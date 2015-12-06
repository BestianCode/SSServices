package sspid

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
)

type PidFile struct {
	f        *os.File
	pid_file string
}

func (_s *PidFile) ON(xpid_file string) {
	var err error

	_s.pid_file = xpid_file
	_s._check()

	_s.f, err = os.OpenFile(_s.pid_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Error create PID file: %s (%v)\n", _s.pid_file, err)
	}
	_s.f.WriteString(fmt.Sprintf("%d", os.Getpid()))
	_s.f.Close()
}

func (_s *PidFile) OFF() {
	var err error
	err = os.Remove(_s.pid_file)
	if err != nil {
		log.Fatalf("Error remove PID file: %s (%v)\n", _s.pid_file, err)
	}
}

func (_s PidFile) _check() {
	var err error

	_s.f, err = os.OpenFile(_s.pid_file, os.O_RDONLY, 0666)
	if err != nil {
		return
	}

	defer _s.f.Close()

	pid_read := make([]byte, 10)

	pid_bytes, err := _s.f.Read(pid_read)
	if err != nil {
		log.Printf("WR1/ > Remove old pid file")
		err = os.Remove(_s.pid_file)
		if err != nil {
			log.Fatalf("ER1/Error remove PID file: %s (%v)\n", _s.pid_file, err)
		}
		return
	}

	if pid_bytes > 0 {
		pid_read_int, err := strconv.Atoi(fmt.Sprintf("%s", pid_read[0:pid_bytes]))
		if err != nil {
			log.Printf("WR2/ > Remove old pid file")
			err = os.Remove(_s.pid_file)
			if err != nil {
				log.Fatalf("ER2/Error remove PID file: %s (%v)\n", _s.pid_file, err)
			}
			return
		}

		pid_proc, err := os.FindProcess(pid_read_int)
		if err != nil {
			log.Printf("WR3/ > Remove old pid file")
			err = os.Remove(_s.pid_file)
			if err != nil {
				log.Fatalf("ER3/Error remove PID file: %s (%v)\n", _s.pid_file, err)
			}
			return
		}

		err = pid_proc.Signal(syscall.Signal(0))
		if err != nil {
			log.Printf("WR4/ > Remove old pid file")
			err = os.Remove(_s.pid_file)
			if err != nil {
				log.Fatalf("ER4/Error remove PID file: %s (%v)\n", _s.pid_file, err)
			}
			return
		}

		log.Fatalf("\t<< ! Another copy of the program with PID %d is running! Exiting... ! >>\n\n\n", pid_read_int)
	} else {
		log.Printf("WR5/ > Remove old pid file")
		err = os.Remove(_s.pid_file)
		if err != nil {
			log.Fatalf("ER5/Error remove PID file: %s (%v)\n", _s.pid_file, err)
		}
		return
	}
}
