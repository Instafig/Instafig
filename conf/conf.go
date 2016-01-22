package conf

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/gpmgo/gopm/modules/goconfig"
)

type DBConfig struct {
	Driver string
	Host   string
	Port   int
	User   string
	DBName string
	PassWd string
}

var (
	Mode           string
	HttpAddr       string
	SqliteDir      string
	SqliteFileName string
	DatabaseConfig = &DBConfig{}
	NodeType       string
	NodeAddr       string
	ClientAddr     string
	MasterAuth     string
	MasterAddr     string
	ReplaceMaster  bool

	DebugMode bool
	LogLevel  string

	configFile   = flag.String("config", "__unset__", "service config file")
	maxThreadNum = flag.Int("max-thread", 0, "max threads of service")
	debugMode    = flag.Bool("debug", false, "debug mode")
	logLevel     = flag.String("log-level", "INFO", "DEBUG | INFO | WARN | ERROR | FATAL | PANIC")
	versionInfo  = flag.Bool("version", false, "show version info")
)

func init() {
	flag.Parse()

	DebugMode = *debugMode
	LogLevel = *logLevel

	if *versionInfo {
		fmt.Printf("%s\n", VersionString())
		os.Exit(0)
	}

	if len(os.Args) == 2 {
		if os.Args[1] == "reload" {
			wd, _ := os.Getwd()
			pidFile, err := os.Open(filepath.Join(wd, "instafig.pid"))
			if err != nil {
				log.Printf("Failed to open pid file: %s", err.Error())
				os.Exit(1)
			}
			pids := make([]byte, 10)
			n, err := pidFile.Read(pids)
			if err != nil {
				log.Printf("Failed to read pid file: %s", err.Error())
				os.Exit(1)
			}
			if n == 0 {
				log.Printf("No pid in pid file: %s", err.Error())
				os.Exit(1)
			}
			_, err = exec.Command("kill", "-USR2", string(pids[:n])).Output()
			if err != nil {
				log.Printf("Failed to restart Instafig service: %s", err.Error())
				os.Exit(1)
			}
			pidFile.Close()
			os.Exit(0)
		}
	}

	if DebugMode {
		LogLevel = "DEBUG"
	}

	if *maxThreadNum == 0 {
		*maxThreadNum = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*maxThreadNum)

	if *configFile == "__unset__" {
		p, _ := os.Getwd()
		*configFile = filepath.Join(p, "conf/config.ini")
	}

	confFile, err := filepath.Abs(*configFile)
	if err != nil {
		log.Panicf("No correct config file: %s - %s", *configFile, err.Error())
	}

	config, err := goconfig.LoadConfigFile(confFile)
	if err != nil {
		log.Panicf("No correct config file: %s - %s", *configFile, err.Error())
	}

	Mode, _ = config.GetValue("", "mode")
	HttpAddr, _ = config.GetValue("", "addr")
	if IsEasyDeployMode() {
		SqliteDir, _ = config.GetValue("sqlite", "dir")
		SqliteFileName, _ = config.GetValue("sqlite", "filename")
	} else {
		DatabaseConfig.Driver, _ = config.GetValue("db", "driver")
		DatabaseConfig.DBName, _ = config.GetValue("db", "db_name")
		DatabaseConfig.Host, _ = config.GetValue("db", "host")
		port, _ := config.GetValue("db", "port")
		DatabaseConfig.Port, err = strconv.Atoi(port)
		if err != nil {
			log.Panicf("DB port is not correct: %s - %s", *configFile, err.Error())
		}
		DatabaseConfig.PassWd, _ = config.GetValue("db", "passwd")
		DatabaseConfig.User, _ = config.GetValue("db", "user")
	}

	NodeType, _ = config.GetValue("node", "type")
	NodeAddr, _ = config.GetValue("node", "node_addr")
	ClientAddr, _ = config.GetValue("node", "client_addr")
	MasterAuth, _ = config.GetValue("node", "master_auth")
	if !IsMasterNode() {
		MasterAddr, _ = config.GetValue("node", "master_addr")
	}
}

func IsEasyDeployMode() bool {
	return Mode == "easy_deploy"
}

func IsMasterNode() bool {
	return NodeType == "master"
}
