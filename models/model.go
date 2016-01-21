package models

import "fmt"

const SCHEME_VERSION = "0.1"

var (
	NoDataVerError = fmt.Errorf("no data version")
)

type User struct {
	Key      string `xorm:"key TEXT PK NOT NULL" json:"key"`
	PassCode string `xorm:"pass_code TEXT NOT NULL"`
	Name     string `xorm:"name TEXT NOT NULL UNIQUE" json:"name"`
	Creator  string `xorm:"creator TEXT NOT NULL" json:"creator"`
}

func (*User) TableName() string {
	return "user"
}

func (m *User) UniqueCond() (string, []interface{}) {
	return "key=?", []interface{}{m.Key}
}

func GetAllUser(s *Session) ([]*User, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*User, 0)
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetUsers(s *Session, page, count int) ([]*User, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*User, 0)
	if err := s.OrderBy("name desc").Limit(count, (page-1)*count).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

const (
	APP_TYPE_TEMPLATE = "template"
	APP_TYPE_REAL     = "real"
)

type App struct {
	Key      string `xorm:"key TEXT PK NOT NULL" json:"key"`
	UserKey  string `xorm:"user_key TEXT NOT NULL" json:"user_key"`
	Name     string `xorm:"name TEXT not NULL" json:"name"`
	Type     string `xorm:"type TEXT not NULL" json:"type"`
	DataSign string `xorm:"data_sign TEXT NOT NULL" json:"data_sign"`
}

func (*App) TableName() string {
	return "app"
}

func (m *App) UniqueCond() (string, []interface{}) {
	return "key=?", []interface{}{m.Key}
}

func GetAllApp(s *Session) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*App, 0)
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetAppsByUserKey(s *Session, userKey string) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*App, 0)
	if err := s.Where("user_key=?", userKey).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetAllApps(s *Session, page int, count int) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*App, 0)
	if err := s.OrderBy("name desc").Limit(count, (page-1)*count).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func IsValidAppType(typ string) bool {
	return typ == APP_TYPE_REAL || typ == APP_TYPE_TEMPLATE
}

const (
	CONF_V_TYPE_STRING   = "string"
	CONF_V_TYPE_INT      = "int"
	CONF_V_TYPE_FLOAT    = "float"
	CONF_V_TYPE_CODE     = "code"
	CONF_V_TYPE_TEMPLATE = "template"
)

type Config struct {
	Key    string `xorm:"key TEXT PK NOT NULL" json:"key"`
	AppKey string `xorm:"app_key TEXT NOT NULL" json:"app_key"`
	K      string `xorm:"k TEXT NOT NULL" json:"k"`
	V      string `xorm:"v TEXT NOT NULL" json:"v"`
	VType  string `xorm:"v_type TEXT NOT NULL" json:"v_type"`
}

func (*Config) TableName() string {
	return "config"
}

func (m *Config) UniqueCond() (string, []interface{}) {
	return "key=?", []interface{}{m.Key}
}

func GetAllConfig(s *Session) ([]*Config, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*Config, 0)
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetConfigsByAppKey(s *Session, appKey string) ([]*Config, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*Config, 0)
	if err := s.Where("app_key=?", appKey).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func IsValidConfType(typ string) bool {
	return typ == CONF_V_TYPE_CODE ||
		typ == CONF_V_TYPE_FLOAT ||
		typ == CONF_V_TYPE_INT ||
		typ == CONF_V_TYPE_STRING ||
		typ == CONF_V_TYPE_TEMPLATE
}

const (
	NODE_TYPE_MASTER = "master"
	NODE_TYPE_SLAVE  = "slave"
)

type Node struct {
	URL            string `xorm:"url TEXT PK NOT NULL" json:"url"`
	NodeURL        string `xorm:"node_url TEXT PK NOT NULL" json:"node_url"`
	Type           string `xorm:"type TEXT NOT NULL" json:"type"`
	CreatedUTC     int    `xorm:"created_utc UTC NOT NULL" json:"created_utc"`
	LastCheckUTC   int    `xorm:"last_check_utc INT NOT NULL" json:"last_check_utc"`
	DataVersionStr string `xorm:"data_version TEXT NOT NULL" json:"data_version"` // json string to store DataVersion in db

	DataVersion   *DataVersion `xorm:"-"`
	SchemeVersion string       `xorm:"-"`
}

func (*Node) TableName() string {
	return "node"
}

func (m *Node) UniqueCond() (string, []interface{}) {
	return "url=?", []interface{}{m.URL}
}

func GetAllNode(s *Session) ([]*Node, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*Node, 0)
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func IsValidNodeType(typ string) bool {
	return typ == NODE_TYPE_MASTER || typ == NODE_TYPE_SLAVE
}

type DataVersion struct {
	Version int    `xorm:"version INT NOT NULL"`
	Sign    string `xorm:"sign TEXT NOT NULL"`
	OldSign string `xorm:"old_sign TEXT NOT NULL"`
}

func (*DataVersion) TableName() string {
	return "data_version"
}

func UpdateDataVersion(s *Session, ver *DataVersion) error {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	sql := fmt.Sprintf("update data_version set version=%d, sign='%s', old_sign='%s'", ver.Version, ver.Sign, ver.OldSign)
	_, err := s.Exec(sql)
	return err
}

func GetDataVersion(s *Session) (*DataVersion, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*DataVersion, 0)
	err := s.Find(&res)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, NoDataVerError
	}

	return res[0], nil
}

const (
	CONFIG_UPDATE_KIND_NEW    = "new"
	CONFIG_UPDATE_KIND_UPDATE = "up"
	CONFIG_UPDATE_KIND_DELETE = "del"
)

type ConfigUpdateHistory struct {
	Id         string `xorm:"id PK TEXT NOT NULL" json:"id"`
	ConfigKey  string `xorm:"config_key TEXT NOT NULL" json:"config_key"`
	Kind       string `xorm:"kind TEXT NOT NULL" json:"kind"`
	K          string `xorm:"k TEXT NOT NULL" json:"k"`
	OldV       string `xorm:"old_v TEXT NOT NULL" jon:"old_v"`
	OldVType   string `xorm:"old_v_type TEXT NOT NULL" json:"old_v_type"`
	NewV       string `xorm:"new_v TEXT NOT NULL" json:"new_v"`
	NewVType   string `xorm:"new_v_type TEXT NOT NULL" json:"new_v_type"`
	UserKey    string `xorm:"user_key TEXT NOT NULL" json:"user_key"`
	CreatedUTC int    `xorm:"created_utc INT NOT NULL" json:"created_utc"`
}

func (*ConfigUpdateHistory) TableName() string {
	return "config_update_history"
}

func (m *ConfigUpdateHistory) UniqueCond() (string, []interface{}) {
	return "id=?", []interface{}{m.Id}
}

func GetConfigUpdateHistory(s *Session, configKey string) ([]*ConfigUpdateHistory, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*ConfigUpdateHistory, 0)
	err := s.Where("config_key=?", configKey).OrderBy("created_utc desc").Find(&res)

	return res, err
}

func GetAllConfigUpdateHistory(s *Session) ([]*ConfigUpdateHistory, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*ConfigUpdateHistory, 0)
	err := s.Find(&res)

	return res, err
}

func ClearModeData(s *Session) error {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	sql := "delete from user; delete from app; delete from config; delete from node;update data_version set version=0;delete from config_update_history"
	_, err := s.Exec(sql)

	return err
}
