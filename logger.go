package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/Instafig/Instafig/conf"
	"github.com/Instafig/Instafig/utils"
	"github.com/Sirupsen/logrus"
)

var (
	logFileLock   = make(map[string]*sync.RWMutex)
	logger        *Logger
	requestLogger *Logger
)

const (
	LOG_LEVEL_DEBUG = logrus.DebugLevel
	LOG_LEVEL_INFO  = logrus.InfoLevel
	LOG_LEVEL_WARN  = logrus.WarnLevel
	LOG_LEVEL_ERROR = logrus.ErrorLevel
	LOG_LEVEL_FATAL = logrus.FatalLevel
	LOG_LEVEL_PANIC = logrus.PanicLevel
)

func init() {
	if conf.DebugMode {
		logger = newLogger(conf.LogDir, "log", LOG_LEVEL_DEBUG)
	} else {
		logger = newLogger(conf.LogDir, "log", LOG_LEVEL_INFO)
	}

	requestLogger = newLogger(conf.LogDir, "request", LOG_LEVEL_INFO)
}

type Logger struct {
	*logrus.Logger

	path         string
	fileNameBase string
	fileNameDate string
}

func (log *Logger) init() {
	log.fileNameDate = utils.GetNowStringYMD()
	p := filepath.Join(log.path, fmt.Sprintf("%s.log", log.fileNameBase))
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		log.Fatalf("Failed to init [%s] log: %s", log.fileNameBase, err.Error())
	}

	log.Out = f
	log.Formatter = new(logrus.JSONFormatter)
}

func (log *Logger) rotate() error {
	nowDate := utils.GetNowStringYMD()
	if nowDate == log.fileNameDate {
		return nil
	}

	logFileLock[log.fileNameBase].Lock()
	defer logFileLock[log.fileNameBase].Unlock()

	if nowDate == log.fileNameDate {
		return nil
	}

	oldFile := log.Out.(*os.File)
	oldFile.Close()
	logFile := filepath.Join(log.path, fmt.Sprintf("%s.log", log.fileNameBase))
	rotateFileName := filepath.Join(log.path, fmt.Sprintf("%s.log.%s", log.fileNameBase, log.fileNameDate))
	err := exec.Command("mv", logFile, rotateFileName).Run()
	if err != nil {
		log.Printf("Failed to rotate log file[%s]: %s ", rotateFileName, err.Error())
	}

	log.fileNameDate = nowDate
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		// TODO:
		return err
	}

	log.Out = f

	return nil
}

func (log *Logger) newLog(level logrus.Level, data map[string]interface{}) {
	log.rotate()

	data["_ts_"] = utils.GetNowMillisecond()
	fields := logrus.Fields(data)

	switch level {
	case LOG_LEVEL_DEBUG:
		log.WithFields(fields).Debug("")
	case LOG_LEVEL_INFO:
		log.WithFields(fields).Info("")
	case LOG_LEVEL_WARN:
		log.WithFields(fields).Warn("")
	case LOG_LEVEL_ERROR:
		log.WithFields(fields).Error("")
	case LOG_LEVEL_FATAL:
		log.WithFields(fields).Fatal("")
	case LOG_LEVEL_PANIC:
		log.WithFields(fields).Panic("")
	default:
		log.WithFields(fields).Info("")
	}
}

func (log *Logger) Debug(data map[string]interface{}) {
	log.newLog(LOG_LEVEL_DEBUG, data)
}

func (log *Logger) Warn(data map[string]interface{}) {
	log.newLog(LOG_LEVEL_WARN, data)
}

func (log *Logger) Info(data map[string]interface{}) {
	log.newLog(LOG_LEVEL_INFO, data)
}

func (log *Logger) Error(data map[string]interface{}) {
	log.newLog(LOG_LEVEL_ERROR, data)
}

func (log *Logger) Fatal(data map[string]interface{}) {
	log.newLog(LOG_LEVEL_FATAL, data)
}

func (log *Logger) Panic(data map[string]interface{}) {
	log.newLog(LOG_LEVEL_PANIC, data)
}

func (log *Logger) SetLevel(level logrus.Level) {
	log.Level = level
}

func newLogger(path string, baseName string, level logrus.Level) *Logger {
	logFileLock[baseName] = &sync.RWMutex{}
	l := &Logger{new(logrus.Logger), path, baseName, ""}

	l.init()
	if conf.DebugMode {
		l.SetLevel(LOG_LEVEL_DEBUG)
	} else {
		l.SetLevel(level)
	}

	return l
}
