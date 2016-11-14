package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"

	"github.com/VividCortex/godaemon"

	//"github.com/lsword/zeus/dbserver"
	"zeus/config"
	"zeus/goserver"
	"zeus/httpserver"
	"zeus/log"
	"zeus/scheduler"
)

var cmdargConfigFile string
var cmdargSignal string
var serverconfig *config.Config

func init() {
	flag.StringVar(&cmdargConfigFile, "c", "", "Configuration File")
	flag.StringVar(&cmdargSignal, "s", "", "Send Signal To Server")
	flag.Parse()
}

func SigHupHandler() {
	/*
	   if cmdargConfigFile != "" {
	       serverconfig = config.InitConfigFromFile(cmdargConfigFile)
	   } else {
	       serverconfig = config.DefaultConfig()
	   }
	*/
	log.SetupLoggerFromConfig(serverconfig)
}

func SigIntHandler() {
	os.Exit(0)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//初始化配置
	serverconfig = config.DefaultConfig()
	if cmdargConfigFile != "" {
		serverconfig = config.InitConfigFromFile(cmdargConfigFile)
	}

	//处理命令行信号
	if cmdargSignal != "" {
		pid := goserver.Pid()
		if pid < 0 {
			fmt.Println("cann't get pid from pidfile:", goserver.PidFileName())
			os.Exit(0)
		}
		switch cmdargSignal {
		case "quit":
			cmd := exec.Command("kill", "-s", "SIGQUIT", strconv.Itoa(pid))
			cmd.Run()
			break
		case "reopen":
			println("received reopen:" + strconv.Itoa(pid))
			cmd := exec.Command("kill", "-s", "SIGHUP", strconv.Itoa(pid))
			cmd.Run()
			break
		default:
			fmt.Println("signal argument: quit")
		}
		os.Exit(0)
	}

	//初始化日志
	log.SetupLoggerFromConfig(serverconfig)

	//初始化基础服务
	goserver.SetSignalHandler(SigHupHandler, syscall.SIGHUP)
	goserver.SetSignalHandler(SigIntHandler, syscall.SIGINT)
	go goserver.Run()

	//初始化数据库
	//dbserver.Run(serverconfig)

	//启动HTTP服务
	httpserver.Run(serverconfig)

	scheduler.Run(serverconfig)

	//启动pprof，用于性能分析
	if serverconfig.Pprof.Switch == "on" {
		go func() {
			http.ListenAndServe(fmt.Sprintf("%s:%d", serverconfig.Pprof.Ip, serverconfig.Pprof.Port), nil)
		}()
	}

	//以Daemon方式运行
	if serverconfig.Daemon.Switch == "on" {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	}

	runtime.Goexit()
}
