package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/blinry/goyaml"
)

type LoggingConfig struct {
	File          string "file"
	Syslog        string "syslog"
	Level         string "level"
	Filename      string "filename"
	ErrorFilename string "errorfilename"
	Maxsize       int    "maxsize"
	Maxrolls      int    "maxrolls"
}

var defaultLoggingConfig = LoggingConfig{
	Level:         "debug",
	Filename:      "/Users/liujian/code/zeus/log/server.log",
	ErrorFilename: "/Users/liujian/code/zeus/log/server_err.log",
	Maxsize:       100000,
	Maxrolls:      5,
}

type DaemonConfig struct {
	Switch string "switch"
}

var defaultDaemonConfig = DaemonConfig{
	Switch: "off",
}

type HttpServerConfig struct {
	Switch string "switch"
	Ip     string "ip"
	Port   uint16 "port"
}

var defaultHttpServerConfig = HttpServerConfig{
	Switch: "on",
	Ip:     "0.0.0.0",
	Port:   8888,
}

type PprofConfig struct {
	Switch string "switch"
	Ip     string "ip"
	Port   uint16 "port"
}

var defaultPprofConfig = PprofConfig{
	Switch: "off",
	Ip:     "0.0.0.0",
	Port:   8888,
}

type DBItemConfig struct {
	DBName         string "DBName"
	DriverName     string "DriverName"
	DataSourceName string "DataSourceName"
	MaxIdleConns   int    "MaxIdleConns"
	MaxOpenConns   int    "MaxOpenConns"
}

type DBServerConfig struct {
	Switch                  string         "switch"
	LogSQLExecuteTimeSwitch string         "log_sql_execute_time_switch"
	ConnCheckInterval       int            "conn_check_interval"
	DBItems                 []DBItemConfig "dbitems"
}

var defaultDBServerConfig = DBServerConfig{
	Switch:                  "on",
	LogSQLExecuteTimeSwitch: "on",
}

type IntervalsConfig struct {
	ContainersInfo_Refresh_Interval uint16 "containersinfo_refresh_interval"
	ImagesInfo_Refresh_Interval     uint16 "imagesinfo_refresh_interval"
	AgentInfo_Refresh_Interval      uint16 "agentinfo_refresh_interval"
}

var defaultIntervalsConfig = IntervalsConfig{
	ContainersInfo_Refresh_Interval: 10,
	ImagesInfo_Refresh_Interval:     10,
	AgentInfo_Refresh_Interval:      3,
}

type SchedulerConfig struct {
	Master         string "master"
	Address        string "address"
	AuthPrincipal  string "authprincipal"
	AuthSecretFile string "authsecretfile"
}

var defaultSchedulerConfig = SchedulerConfig{
	Master:         "127.0.0.1:5050",
	Address:        "127.0.0.1",
	AuthPrincipal:  "",
	AuthSecretFile: "",
}

type Config struct {
	Logging    LoggingConfig    "logging"
	Daemon     DaemonConfig     "daemon"
	HttpServer HttpServerConfig "httpserver"
	Pprof      PprofConfig      "pprof"
	DBServer   DBServerConfig   "dbserver"
	Intervals  IntervalsConfig  "intervals"
	Scheduler  SchedulerConfig  "scheduler"
}

var defaultConfig = Config{
	Logging:    defaultLoggingConfig,
	Daemon:     defaultDaemonConfig,
	HttpServer: defaultHttpServerConfig,
	Pprof:      defaultPprofConfig,
	DBServer:   defaultDBServerConfig,
	Intervals:  defaultIntervalsConfig,
	Scheduler:  defaultSchedulerConfig,
}

var config Config

func DefaultConfig() *Config {
	//c := defaultConfig
	//return &c
	config = Config{
		Logging:    defaultLoggingConfig,
		Daemon:     defaultDaemonConfig,
		HttpServer: defaultHttpServerConfig,
		Pprof:      defaultPprofConfig,
		DBServer:   defaultDBServerConfig,
		Intervals:  defaultIntervalsConfig,
	}
	return &config
}

func InitConfigFromFile(path string) *Config {
	var c *Config = DefaultConfig()
	var e error

	b, e := ioutil.ReadFile(path)
	if e != nil {
		panic(e.Error())
	}

	e = goyaml.Unmarshal(b, c)
	if e != nil {
		panic(e.Error())
	}

	return c
}

func CurConfig() *Config {
	return &config
}

func ConfigJson() string {
	jsonbyte, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return ""
	} else {
		return string(jsonbyte)
	}
}
