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

type Session struct {
	*xorm.Session
}

func init() {
	var err error
	var dsn, driver string

	if conf.IsEasyDeployMode() {
		if conf.IsMasterNode() {
			dsn = fmt.Sprintf(filepath.Join(conf.SqliteDir, conf.SqliteFileName))
		} else {
			dsn = fmt.Sprintf(filepath.Join(conf.SqliteDir, conf.SqliteFileName+fmt.Sprintf(".%s", conf.MasterAddr)))
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

	if dbEngineDefault, err = xorm.NewEngine(driver, dsn); err != nil {
		log.Fatal("Failed to init db engine: " + err.Error())
	}
	dbEngineDefault.SetMaxOpenConns(100)
	dbEngineDefault.SetMaxIdleConns(50)
	dbEngineDefault.ShowErr = true
	dbEngineDefault.ShowSQL = conf.DebugMode

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

func InsertMultiRows(_s *Session, m []interface{}) (err error) {
	var s *Session

	if _s == nil {
		s = NewSession()
		defer s.Close()

		if err = s.Begin(); err != nil {
			return err
		}
	} else {
		s = _s
	}

	length := len(m)
	if length > 100 {
		for i := 0; 100*i < length; i++ {
			endIndex := 100 * (i + 1)
			if endIndex > length {
				endIndex = length
			}
			gg := m[100*i : endIndex]
			_, err = s.AllCols().Insert(gg...)
			if err != nil {
				break
			}
		}
	} else {
		_, err = s.AllCols().Insert(m...)
	}

	if _s == nil {
		if err != nil {
			s.Rollback()
		} else {
			err = s.Commit()
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
