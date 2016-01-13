package models

const SCHEME_VERSION = "0.1"

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

func GetAllUser(s *ModelSession) ([]*User, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*User, 0)
	if err := s.Find(&res); err != nil {
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

func GetAllApp(s *ModelSession) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*App, 0)
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func IsValidAppType(typ string) bool {
	return typ == APP_TYPE_REAL || typ == APP_TYPE_TEMPLATE
}

const (
	CONF_V_TYPE_STRING = "string"
	CONF_V_TYPE_INT    = "int"
	CONF_V_TYPE_FLOAT  = "float"
	CONF_V_TYPE_CODE   = "code"
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

func GetAllConfig(s *ModelSession) ([]*Config, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := make([]*Config, 0)
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func IsValidConfType(typ string) bool {
	return typ == CONF_V_TYPE_CODE || typ == CONF_V_TYPE_FLOAT || typ == CONF_V_TYPE_INT || typ == CONF_V_TYPE_STRING
}

const (
	NODE_TYPE_MASTER = "master"
	NODE_TYPE_SLAVE  = "slave"
)

type Node struct {
	Host         string `xorm:"Host TEXT NOT NULL UNIQUE(node_key)" json:"host"`
	Port         int    `xorm:"port INT NOT NULL UNIQUE(node_key)" json:"port"`
	Type         string `xorm:"type TEXT NOT NULL" json:"type"`
	DBVersion    int    `xorm:"db_version INT NOT NULL" json:"db_version"`
	CreatedUTC   int    `xorm:"created_utc UTC NOT NULL" json:"created_utc"`
	LastCheckUTC string `xorm:"last_check_utc INT NOT NULL" json:"last_check_utc"`
}

func (*Node) TableName() string {
	return "node"
}

func (m *Node) UniqueCond() (string, []interface{}) {
	return "host=? and port=?", []interface{}{m.Host, m.Port}
}

func GetAllNode(s *ModelSession) ([]*Node, error) {
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
