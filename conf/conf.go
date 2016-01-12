package conf

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gpmgo/gopm/modules/goconfig"
	"strconv"
)

type DatabaseConfig struct {
	Driver string
	Host   string
	Port   int
	User   string
	DBName string
	PassWd string
}


var (
	Mode string
	HttpAddr string
	SqliteDir string
	DatabaseConfig = &DatabaseConfig{}
	NodeType string
	NodeAddr string
	MasterAuth string
	MasterAddr string

	DebugMode bool
	LogLevel int

	configFile = flag.String("config", "__unset__", "service config file")
	maxThreadNum      = flag.Int("max-thread", 0, "max threads of service")
	debugMode         = flag.Bool("debug", false, "debug mode")
	logLevel          = flag.String("log-level", "INFO", "DEBUG | INFO | WARN | ERROR | FATAL | PANIC")
)

func init() {
	flag.Parse()

	DebugMode = *debugMode
	LogLevel = *logLevel
	if DebugMode {
		LogLevel = "DEBUG"
	}

	if *maxThreadNum == 0 {
		*maxThreadNum = runtime.NumCPU() / 2
	}
	runtime.GOMAXPROCS(*maxThreadNum)

	if *configFile == "__unset__" {
		p, _ := os.Getwd()
		*configFile = filepath.Join(p, "config.ini")
	}

	confFile, err := filepath.Abs(*configFile)
	if err != nil {
		log.Panicf("No correct config file: %s - %s", *configFile, err.Error())
	}

	config, err := goconfig.LoadConfigFile(confFile)
	if err != nil {
		log.Panicf("No correct config file: %s - %s", *configFile, err.Error())
	}

	Mode = config.GetValue("", "mode")
	HttpAddr = config.GetValue("", "addr")
	if IsEasyMode() {
		SqliteDir = config.GetValue("sqlite", "dir")
	} else {
		DatabaseConfig.Driver, _ = config.GetValue("db", "driver")
		DatabaseConfig.DBName, _ = config.GetValue("db", "db_name")
		DatabaseConfig.Host, _ = config.GetValue("db", "host")
		port, _ := config.GetValue("db", "port")
		DatabaseConfig.Port, err = strconv.Atoi(port)
		if err != nil {
			log.Panicf("DB port is not correct: %s - %s", *configFile, err.Error())
		}
		DatabaseConfig.PassWd = config.GetValue("db", "passwd")
		DatabaseConfig.User = config.GetValue("db", "user")
	}

	NodeType, _ = config.GetValue("node", "type")
	NodeAddr, _ = config.GetValue("node", "addr")
	MasterAuth, _ = config.GetValue("node", "master_auth")
	if !IsMasterNode() {
		MasterAddr, _ = config.GetValue("node", "master_addr")
	}

}

func IsEasyMode() bool {
	return Mode == "easy"
}

func IsMasterNode() bool {
	return NodeType == "master"
}

