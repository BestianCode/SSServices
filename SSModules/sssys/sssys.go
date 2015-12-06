package sssys

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/BestianRU/SSServices/SSModules/sslog"
	"github.com/BestianRU/SSServices/SSModules/sspid"
)

func Exit(signalType os.Signal, pid sspid.PidFile, xlog string) {
	var rLog sslog.LogFile

	rLog.ON(xlog, 0)
	defer rLog.OFF()

	rLog.Log(".")
	rLog.Log("..")
	rLog.Log("...")
	rLog.Log("Exit command received. Exiting...")
	rLog.Log("Signal type: ", signalType)
	rLog.Log("Bye...")
	rLog.Log("...")
	rLog.Log("..")
	rLog.Log(".")

	pid.OFF()

	os.Exit(0)
}

func Fork(xdaemon, xconfig string) {
	var err error

	if xdaemon != "YES" {
		return
	}

	err = exec.Command(os.Args[0], "-daemon=GO", fmt.Sprintf("-config=%s", xconfig), " &").Start()
	if err != nil {
		fmt.Println("\tFork daemon error: %v\n\n\n", err)
		os.Exit(1)
	} else {
		fmt.Println("\tForked!\n\n\n")
		os.Exit(0)
	}
}

func Signal(pid sspid.PidFile, xlog string) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		signalType := <-ch
		signal.Stop(ch)
		Exit(signalType, pid, xlog)
	}()
}
