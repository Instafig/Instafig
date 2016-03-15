package models

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/appwilldev/Instafig/conf"
	xormcore "github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbEngineDefault *xorm.Engine
)

func getSerialNumFromSlaveSqliteFile(fn string) int {
	ix := strings.LastIndex(fn, ".")
	if ix == -1 {
		return -1
	}

	id, err := strconv.Atoi(fn[ix+1:])
	if err != nil {
		return -1
	}

	return id
}

func getSlaveSqliteFile() (string, error) {
	files, err := filepath.Glob(filepath.Join(conf.SqliteDir, fmt.Sprintf("%s.%s*", conf.SqliteFileName, conf.MasterAddr)))
	if err != nil {
		return "", err
	}

	maxId := 1
	for _, file := range files {
		if id := getSerialNumFromSlaveSqliteFile(file); id > maxId {
			maxId = id
		}
	}

	return filepath.Join(conf.SqliteDir, fmt.Sprintf("%s.%s.%d", conf.SqliteFileName, conf.MasterAddr, maxId)), nil
}

func getSlaveNextSqliteFile() (string, error) {
	files, err := filepath.Glob(filepath.Join(conf.SqliteDir, fmt.Sprintf("%s.%s*", conf.SqliteFileName, conf.MasterAddr)))
	if err != nil {
		return "", err
	}

	maxId := 1
	for _, file := range files {
		if id := getSerialNumFromSlaveSqliteFile(file); id >= maxId {
			maxId = id + 1
		}
	}

	return filepath.Join(conf.SqliteDir, fmt.Sprintf("%s.%s.%d", conf.SqliteFileName, conf.MasterAddr, maxId)), nil
}

func UpdateSqliteDBEngine() {
	dsn, err := getSlaveNextSqliteFile()
	if err != nil {
		log.Panicf("Failed to generate sqlie lite file: %s", err.Error())
	}
	initDBEngine("sqlite3", dsn)
}

type Session struct {
	*xorm.Session
}

func init() {
	var dsn, driver string
	var err error

	if conf.IsEasyDeployMode() {
		if conf.IsMasterNode() {
			dsn = filepath.Join(conf.SqliteDir, conf.SqliteFileName)
		} else {
			dsn, err = getSlaveSqliteFile()
			if err != nil {
				log.Panicf("Failed to generate sqlie lite file: %s", err.Error())
			}
		}
		driver = "sqlite3"
	} else {
		dsn = fmt.Sprintf(
			"user=%s dbname=%s host=%s port=%d sslmode=disable",
			conf.DatabaseConfig.User,
			conf.DatabaseConfig.DBName,
			conf.DatabaseConfig.Host,
			conf.DatabaseConfig.Port)
		if conf.DatabaseConfig.PassWd != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, conf.DatabaseConfig.PassWd)
		}
		driver = conf.DatabaseConfig.Driver
	}

	initDBEngine(driver, dsn)
}

func initDBEngine(driver, dsn string) {
	var err error

	if dbEngineDefault, err = xorm.NewEngine(driver, dsn); err != nil {
		log.Fatal("Failed to init db engine: " + err.Error())
	}
	dbEngineDefault.SetMaxOpenConns(100)
	dbEngineDefault.SetMaxIdleConns(50)
	if conf.DebugMode {
		dbEngineDefault.Logger().SetLevel(xormcore.LOG_DEBUG)
		//dbEngineDefault.ShowSQL(true)
	} else {
		dbEngineDefault.Logger().SetLevel(xormcore.LOG_ERR)
	}

	if conf.IsEasyDeployMode() {
		if err = dbEngineDefault.Sync2(
			&User{}, &App{},
			&Config{}, &ConfigUpdateHistory{},
			&Node{}, &DataVersion{}, &WebHook{},
		); err != nil {
			log.Panicf("Failed to sync db scheme: %s", err.Error())
		}

		_, err := GetDataVersion(nil)
		if err != nil {
			if err != NoDataVerError {
				log.Panicf("failed to get data version: %s", err.Error())
			} else {
				_, err = dbEngineDefault.Exec("INSERT INTO data_version(version, sign, old_sign) VALUES(0,'','')")
				if err != nil {
					log.Panicf("failed to init data version: %s", err.Error())
				}
			}
		}
	}
}

func NewSession() *Session {
	ms := new(Session)
	ms.Session = dbEngineDefault.NewSession()

	return ms
}

func newAutoCloseModelsSession() *Session {
	ms := new(Session)
	ms.Session = dbEngineDefault.NewSession()
	ms.IsAutoClose = true

	return ms
}

type DBModel interface {
	TableName() string
}

func InsertRow(s *Session, m DBModel) (err error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}
	_, err = s.AllCols().InsertOne(m)

	return
}

func InsertMultiRows(s *Session, m []interface{}) (err error) {
	var _s *Session

	if s == nil {
		_s = NewSession()
		defer _s.Close()

		if err = _s.Begin(); err != nil {
			return err
		}
	} else {
		_s = s
	}

	_, err = _s.AllCols().Insert(m...)
	if s == nil {
		if err != nil {
			_s.Rollback()
		} else {
			err = _s.Commit()
		}
	}

	return
}

type UniqueDBModel interface {
	TableName() string
	UniqueCond() (string, []interface{})
}

func UpdateDBModel(s *Session, m UniqueDBModel) (err error) {
	whereStr, whereArgs := m.UniqueCond()
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	_, err = s.AllCols().Where(whereStr, whereArgs...).Update(m)

	return
}

func DeleteDBModel(s *Session, m UniqueDBModel) (err error) {
	whereStr, whereArgs := m.UniqueCond()

	if s == nil {
		s = newAutoCloseModelsSession()
	}

	_, err = s.Where(whereStr, whereArgs...).Delete(m)

	return
}
