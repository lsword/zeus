package log

import (
	"fmt"
	"github.com/cihub/seelog"
	"zeus/config"
)

var Logger seelog.LoggerInterface

func init() {
}

func SetupLoggerFromConfig(c *config.Config) {
	logConfigTmp := `
        <seelog minlevel="%s">
            <outputs formatid="common">
                <rollingfile type="size" filename="%s" maxsize="%d" maxrolls="%d" />
                <filter levels="error">
                    <file path="%s" formatid="error"/>
                </filter>
            </outputs>
            <formats>
                <format id="common" format="%s" />
                <format id="error" format="%s" />
            </formats>
        </seelog>
    `
	logConfig := fmt.Sprintf(logConfigTmp,
		c.Logging.Level,
		c.Logging.Filename,
		c.Logging.Maxsize,
		c.Logging.Maxrolls,
		c.Logging.ErrorFilename,
		"%Date/%Time [%LEV] %Msg%n",
		"%Date/%Time %File %FullPath %Func %Msg%n")
	logger, err := seelog.LoggerFromConfigAsBytes([]byte(logConfig))
	if err != nil {
		panic(err)
	}
	//seelog.ReplaceLogger(logger)
	seelog.Current.Flush()
	seelog.Current.Close()
	seelog.Current = logger
	Logger = logger
}

func Flush() { seelog.Flush() }

func Critical(msg ...interface{}) { seelog.Critical(msg) }
func Fatal(msg ...interface{})    { seelog.Critical(msg) }
func Error(msg ...interface{})    { seelog.Error(msg) }
func Warn(msg ...interface{})     { seelog.Warn(msg) }
func Info(msg ...interface{})     { seelog.Info(msg) }
func Debug(msg ...interface{})    { seelog.Debug(msg) }

func Criticalf(msg string, vals ...interface{}) { seelog.Criticalf(msg, vals...) }
func Fatalf(msg string, vals ...interface{})    { seelog.Criticalf(msg, vals...) }
func Errorf(msg string, vals ...interface{})    { seelog.Errorf(msg, vals...) }
func Warnf(msg string, vals ...interface{})     { seelog.Warnf(msg, vals...) }
func Infof(msg string, vals ...interface{})     { seelog.Infof(msg, vals...) }
func Debugf(msg string, vals ...interface{})    { seelog.Debugf(msg, vals...) }
