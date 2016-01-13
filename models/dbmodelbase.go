package models

import (
	//	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/appwilldev/Instafig/conf"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbEngineDefault *xorm.Engine
)

type ModelSession struct {
	*xorm.Session
}

func init() {
	var err error
	var dsn, driver string

	if conf.IsEasyDeployMode() {
		dsn = fmt.Sprintf(filepath.Join(conf.SqliteDir, "instafig.db"))
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

	if dbEngineDefault, err = xorm.NewEngine(driver, dsn); err != nil {
		log.Fatal("Failed to init db engine: " + err.Error())
	}
	dbEngineDefault.SetMaxOpenConns(100)
	dbEngineDefault.SetMaxIdleConns(50)
	dbEngineDefault.ShowErr = true
	dbEngineDefault.ShowSQL = conf.DebugMode

	if conf.IsEasyDeployMode() {
		if err = dbEngineDefault.Sync2(&User{}, &App{}, &Config{}, &Node{}); err != nil {
			log.Panicf("Failed to sync db scheme: %s", err.Error())
		}
	}

}

func NewModelSession() *ModelSession {
	ms := new(ModelSession)
	ms.Session = dbEngineDefault.NewSession()

	return ms
}

func newAutoCloseModelsSession() *ModelSession {
	ms := new(ModelSession)
	ms.Session = dbEngineDefault.NewSession()
	ms.IsAutoClose = true

	return ms
}

type DBModel interface {
	TableName() string
	UniqueCond() (string, []interface{})
}

func InsertDBModel(s *ModelSession, m DBModel) (err error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}
	_, err = s.AllCols().InsertOne(m)

	return
}

func UpdateDBModel(s *ModelSession, m DBModel) (err error) {
	whereStr, whereArgs := m.UniqueCond()
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	_, err = s.AllCols().Where(whereStr, whereArgs...).Update(m)

	return
}

func DeleteDBModel(s *ModelSession, m DBModel) (err error) {
	whereStr, whereArgs := m.UniqueCond()

	if s == nil {
		s = newAutoCloseModelsSession()
	}

	_, err = s.Where(whereStr, whereArgs...).Delete(m)

	return
}
