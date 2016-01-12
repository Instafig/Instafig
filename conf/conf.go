package conf

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

type DatabaseConfig struct {
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


	configFile = flag.String("config", "__unset__", "service config file")
	maxThreadNum      = flag.Int("max-thread", 0, "max threads of service")
	debugMode         = flag.Bool("debug", false, "debug mode")
	logLevel          = flag.String("log-level", "INFO", "DEBUG | INFO | WARN | ERROR | FATAL | PANIC")
)

