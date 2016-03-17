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
	"strings"

	"github.com/gpmgo/gopm/modules/goconfig"
)

var (
	Port               int
	SqliteDir          string
	SqliteFileName     string
	NodeType           string
	NodeAddr           string
	ClientAddr         string
	NodeAuth           string
	MasterAddr         string
	CheckMasterInerval int
	DataExpires        int

	UserPassCodeEncryptKey string

	WebDebugMode     bool
	DebugMode        bool
	ShowSql          bool
	RequestLogEnable bool
	LogDir           string

	StatisticEnable        bool
	InfluxURL              string
	InfluxDB               string
	InfluxUser             string
	InfluxPassword         string
	InfluxBatchPointsCount int

	configFile   = flag.String("config", "__unset__", "service config file")
	maxThreadNum = flag.Int("max-thread", 0, "max threads of service")
	debugMode    = flag.Bool("debug", false, "debug mode")
	showSql      = flag.Bool("show-sql", false, "show sql")
	webDebugMode = flag.Bool("web-debug", false, "web debug mode")
	versionInfo  = flag.Bool("version", false, "show version info")
)

func init() {
	flag.Parse()

	DebugMode = *debugMode
	WebDebugMode = *webDebugMode
	ShowSql = *showSql

	if *versionInfo {
		fmt.Printf("%s\n", VersionString())
		os.Exit(0)
	}

	if len(os.Args) == 2 && os.Args[1] == "reload" {
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
		log.Printf("No correct config file: %s - %s", *configFile, err.Error())
		os.Exit(1)
	}

	config, err := goconfig.LoadConfigFile(confFile)
	if err != nil {
		log.Printf("No correct config file: %s - %s", *configFile, err.Error())
		os.Exit(1)
	}

	ClientAddr, _ = config.GetValue("", "http_addr")
	s := strings.Split(ClientAddr, ":")
	if len(s) != 1 && len(s) != 2 {
		log.Printf("No correct http_addr(%s)", ClientAddr)
		os.Exit(1)
	}
	port := "80"
	if len(s) == 2 {
		port = s[1]
	}
	if Port, err = strconv.Atoi(port); err != nil {
		log.Printf("No correct port(%s): %s", port, err.Error())
		os.Exit(1)
	}

	if requestLogEnable, _ := config.GetValue("", "request_log_enable"); requestLogEnable == "yes" {
		RequestLogEnable = true
	}

	LogDir, _ = config.GetValue("", "log_dir")
	if LogDir, err = filepath.Abs(LogDir); err != nil {
		log.Printf("Bad log_dir value: %s - %s", LogDir, err.Error())
		os.Exit(1)
	}
	if exec.Command("mkdir", "-p", LogDir).Run() != nil {
		log.Printf("Failed to create log_dir value: %s - %s", LogDir, err.Error())
		os.Exit(1)
	}

	UserPassCodeEncryptKey, _ = config.GetValue("", "user_passcode_encrypt_key")

	SqliteDir, _ = config.GetValue("sqlite", "dir")
	if SqliteDir, err = filepath.Abs(SqliteDir); err != nil {
		log.Println("sqlite dir is not correct: " + err.Error())
		os.Exit(1)
	}

	SqliteFileName, _ = config.GetValue("sqlite", "filename")
	NodeType, _ = config.GetValue("node", "type")
	NodeAddr, _ = config.GetValue("node", "node_addr")
	NodeAuth, _ = config.GetValue("node", "node_auth")
	if !IsMasterNode() {
		MasterAddr, _ = config.GetValue("node", "master_addr")
	}

	if !IsMasterNode() {
		intervalStr, _ := config.GetValue("node", "check_master_interval")
		if CheckMasterInerval, err = strconv.Atoi(intervalStr); err != nil {
			log.Printf("No correct expires: %s - %s", intervalStr, err.Error())
			os.Exit(1)
		}

		expiresStr, _ := config.GetValue("node", "data_expires")
		if expiresStr != "" {
			if DataExpires, err = strconv.Atoi(expiresStr); err != nil {
				log.Printf("No correct expires: %s - %s", expiresStr, err.Error())
				os.Exit(1)
			}
			if DataExpires <= CheckMasterInerval {
				DataExpires = CheckMasterInerval * 2
			}
		} else {
			DataExpires = -1
		}
	}

	if statisticEnable, _ := config.GetValue("statistic", "enable"); statisticEnable == "on" {
		StatisticEnable = true
		InfluxDB, _ = config.GetValue("statistic", "influx_db")
		InfluxURL, _ = config.GetValue("statistic", "influx_url")
		InfluxUser, _ = config.GetValue("statistic", "influx_user")
		InfluxPassword, _ = config.GetValue("statistic", "influx_password")
		batchCount, _ := config.GetValue("statistic", "influx_batch_points_count")
		if InfluxBatchPointsCount, err = strconv.Atoi(batchCount); err != nil {
			log.Println("influx_batch_point_count is not number: ", batchCount)
			os.Exit(1)
		}
	}

	if !DebugMode {
		// disable all console log
		nullFile, _ := os.Open(os.DevNull)
		log.SetOutput(nullFile)
		os.Stdout = nullFile
	}
}

func IsMasterNode() bool {
	return NodeType == "master"
}
