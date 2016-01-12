package models

import (
	//	"database/sql"
	"fmt"
	"log"

	"github.com/appwilldev/Instafig/conf"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbEngineDefault *xorm.Engine

//	dbEngineDefaultRaw *sql.DB
//	int64Type          = reflect.TypeOf(int64(1))
)

type ModelSession struct {
	*xorm.Session
}

func init() {
	var err error
	dsn := fmt.Sprintf(
		"user=%s dbname=%s host=%s port=%d sslmode=disable",
		conf.DBConfig["default"].User,
		conf.DBConfig["default"].DBName,
		conf.DBConfig["default"].Host,
		conf.DBConfig["default"].Port)

	if dbEngineDefault, err = xorm.NewEngine("postgres", dsn); err != nil {
		log.Fatal("Failed to init db engine: " + err.Error())
	}
	dbEngineDefault.SetMaxOpenConns(100)
	dbEngineDefault.SetMaxIdleConns(50)
	dbEngineDefault.ShowErr = true
	dbEngineDefault.ShowSQL = conf.DebugMode

	//	if dbEngineDefaultRaw, err = sql.Open("postgres", dsn); err != nil {
	//		log.Fatal("Failed to init db engine: " + err.Error())
	//	}
	//	dbEngineDefaultRaw.SetMaxIdleConns(10)
	//	dbEngineDefaultRaw.SetMaxOpenConns(20)
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

//func rawSqlQuery(sqlStr string, columnTypes []reflect.Type, queryArgs ...interface{}) ([][]interface{}, error) {
//	rows, err := dbEngineDefaultRaw.Query(sqlStr, queryArgs...)
//	if err != nil {
//		return nil, err
//	}
//
//	res := make([][]interface{}, 0)
//	scanDestValue := make([]interface{}, len(columnTypes))
//	for i, columnType := range columnTypes {
//		scanDestValue[i] = reflect.New(columnType).Interface()
//	}
//
//	for rows.Next() {
//		if err = rows.Scan(scanDestValue...); err != nil {
//			return nil, err
//		}
//		scanRow := make([]interface{}, len(columnTypes))
//		for i := range scanRow {
//			scanRow[i] = reflect.Indirect(reflect.ValueOf(scanDestValue[i])).Interface()
//		}
//		res = append(res, scanRow)
//	}
//
//	return res, nil
//}
//
//func generateSequenceValue(sequenceName string) (int64, error) {
//	var sql = fmt.Sprintf("SELECT nextval('%s')", sequenceName)
//	columnTypes := []reflect.Type{int64Type}
//
//	rows, err := rawSqlQuery(sql, columnTypes)
//	if err != nil {
//		logger.ErrorLogger.Error(map[string]interface{}{
//			"type":     "gen_model_id",
//			"seq_name": sequenceName,
//			"err_msg":  err.Error(),
//		})
//		return 0, fmt.Errorf("gen %s sequence error: %s", sequenceName, err.Error())
//	}
//	if len(rows) == 0 {
//		logger.ErrorLogger.Error(map[string]interface{}{
//			"type":     "gen_model_id",
//			"seq_name": sequenceName,
//			"err_msg":  fmt.Sprintf("gen %s sequence error: failed to increase id", sequenceName),
//		})
//		return 0, fmt.Errorf("gen %s sequence error: failed to increase id", sequenceName)
//	}
//
//	return rows[0][0].(int64), nil
//}
