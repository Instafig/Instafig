package models

import "fmt"

const SCHEME_VERSION = "0.1"

var (
	NoDataVerError = fmt.Errorf("no data version")
)

type User struct {
	Key  string `xorm:"key TEXT PK NOT NULL" json:"key"`
	Name string `xorm:"name TEXT NOT NULL UNIQUE" json:"name"`
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
	Key     string `xorm:"key TEXT PK NOT NULL" json:"key"`
	UserKey string `xorm:"user_key TEXT NOT NULL" json:"user_key"`
	Name    string `xorm:"name TEXT not NULL" json:"name"`
	Type    string `xorm:"type TEXT not NULL" json:"type"`
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
	URL          string `xorm:"url TEXT PK NOT NULL" json:"url"`
	Type         string `xorm:"type TEXT NOT NULL" json:"type"`
	DataVersion  int    `xorm:"data_version INT NOT NULL" json:"data_version"`
	CreatedUTC   int    `xorm:"created_utc UTC NOT NULL" json:"created_utc"`
	LastCheckUTC string `xorm:"last_check_utc INT NOT NULL" json:"last_check_utc"`

	SchemeVersion string `xorm:"-"`
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
	Ver int `xorm:"ver INT NOT NULL"`
}

func (*DataVersion) TableName() string {
	return "data_version"
}

func UpdateDataVersion(s *Session, ver int) error {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	sql := fmt.Sprintf("update data_version set ver=%d", ver)
	_, err := s.Exec(sql)
	return err
}

func GetDataVersion(s *Session) (int, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*DataVersion, 0)
	err := s.Find(&res)
	if err != nil {
		return 0, err
	}

	if len(res) == 0 {
		return 0, NoDataVerError
	}

	return res[0].Ver, nil
}
