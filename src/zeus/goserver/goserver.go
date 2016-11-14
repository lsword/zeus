package goserver

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	gosignal "os/signal"
	"path/filepath"
	"strconv"
	"strings"
	//"syscall"
)

var signalHandlerMap map[os.Signal]func()

func PidFileName() string {
	file, _ := exec.LookPath(os.Args[0])
	pidFileName, _ := filepath.Abs(file)
	pidFileName = pidFileName + ".pid"
	return pidFileName
}

func Pid() int {
	if pidByte, err := ioutil.ReadFile(PidFileName()); err == nil {
		pidString := strings.TrimSpace(string(pidByte))
		if pid, err := strconv.Atoi(pidString); err == nil {
			return pid
		}
	}
	return -1
}

func CreatePidFile() {
	pidFileName := PidFileName()
	if pidByte, err := ioutil.ReadFile(pidFileName); err == nil {
		pidString := strings.TrimSpace(string(pidByte))
		if pid, err := strconv.Atoi(pidString); err == nil {
			if _, err := os.Stat(filepath.Join("/proc", strconv.Itoa(pid))); err == nil {
				fmt.Printf("Pid file found, ensure server is not running or delete %s\n", pidFileName)
				os.Exit(-1)
			}
		}
	}
	if err := ioutil.WriteFile(pidFileName, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		fmt.Printf("Create Pid file error: %s\n", err.Error())
		os.Exit(-1)
	}
}

func RemovePidFile() {
	os.Remove(PidFileName())
}

func SetSignalHandler(handler func(), sig os.Signal) {
	if signalHandlerMap == nil {
		signalHandlerMap = make(map[os.Signal]func())
	}
	signalHandlerMap[sig] = handler
}

func Run() {
	CreatePidFile()

	signalChan := make(chan os.Signal, 1)
	var signals []os.Signal
	for k := range signalHandlerMap {
		signals = append(signals, k)
	}
	gosignal.Notify(signalChan, signals...)
	go func() {
		for sig := range signalChan {
			if signalHandlerMap[sig] != nil {
				signalHandlerMap[sig]()
			}
		}
	}()
}
