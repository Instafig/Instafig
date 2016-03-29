package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

//const SCHEME_VERSION = "0.1"

var (
	NoDataVerError = fmt.Errorf("no data version")
)

const (
	USER_STATUS_ACTIVE   = 0
	USER_STATUS_INACTIVE = -1
)

type User struct {
	Key        string `xorm:"key TEXT PK " json:"key"`
	PassCode   string `xorm:"pass_code TEXT " json:"pass_code"`
	Name       string `xorm:"name TEXT  UNIQUE" json:"name"`
	CreatorKey string `xorm:"creator_key TEXT " json:"creator_key"`
	CreatedUTC int    `xorm:"created_utc INT " json:"created_utc"`
	AuxInfo    string `xorm:"aux_info TEXT" json:"aux_info"`
	Status     int    `xorm:"status INT" json:"status"`

	CreatorName string `xorm:"-" json:"creator_name"`
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

	var res []*User
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetUsers(s *Session, page, count int) ([]*User, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*User
	if err := s.OrderBy("name desc").Limit(count, (page-1)*count).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetUserCount(s *Session) (int, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	count, err := s.Count(&User{})

	return int(count), err
}

const (
	APP_TYPE_TEMPLATE = "template"
	APP_TYPE_REAL     = "real"
)

type App struct {
	Key           string `xorm:"key TEXT PK " json:"key"`
	UserKey       string `xorm:"user_key TEXT " json:"creator_key"`
	Name          string `xorm:"name TEXT not NULL" json:"name"`
	Type          string `xorm:"type TEXT not NULL" json:"type"`
	DataSign      string `xorm:"data_sign TEXT " json:"data_sign"`
	CreatedUTC    int    `xorm:"created_utc INT " json:"created_utc"`
	LastUpdateId  string `xorm:"last_update_id TEXT " json:"last_update_id"`
	LastUpdateUTC int    `xorm:"last_update_utc INT " json:"last_update_utc"`
	KeyCount      int    `xorm:"key_count INT " json:"key_count"`
	UpdateTimes   int    `xorm:"update_times INT " json:"update_times"`
	AuxInfo       string `xorm:"aux_info TEXT" json:"aux_info"`

	UserName       string               `xorm:"-" json:"creator_name"`
	LastUpdateInfo *ConfigUpdateHistory `xorm:"-" json:"last_update_info"`
}

func (*App) TableName() string {
	return "app"
}

func (m *App) UniqueCond() (string, []interface{}) {
	return "key=?", []interface{}{m.Key}
}

func GetAppByKey(s *Session, key string) (*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := &App{}
	if has, err := s.Where("key=?", key).Get(res); !has || err != nil {
		return nil, err
	}

	return res, nil
}

func GetAllApps(s *Session) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*App
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetAppsByUserKey(s *Session, userKey string) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*App
	if err := s.Where("user_key=?", userKey).OrderBy("last_update_utc desc").Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetAllAppsPage(s *Session, page int, count int) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*App
	if err := s.OrderBy("last_update_utc desc").Limit(count, (page-1)*count).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetAppCount(s *Session) (int, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	count, err := s.Count(&App{})

	return int(count), err
}

func SearchAppByName(s *Session, q string, count int) ([]*App, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*App
	var err error

	if count > 0 {
		err = s.Where(fmt.Sprintf("instr(lower(name), '%s')>0", strings.ToLower(q))).OrderBy(fmt.Sprintf("instr(lower(name), '%s') asc", strings.ToLower(q))).Limit(count).Find(&res)
	} else {
		err = s.Where(fmt.Sprintf("instr(lower(name), '%s')>0", strings.ToLower(q))).OrderBy(fmt.Sprintf("instr(lower(name), '%s') asc", strings.ToLower(q))).Find(&res)
	}

	return res, err
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

	CONF_STATUS_INACTIVE = 0
	CONF_STATUS_ACTIVE   = 1
)

type Config struct {
	Key          string `xorm:"key TEXT PK " json:"key"`
	AppKey       string `xorm:"app_key TEXT INDEX" json:"app_key"`
	K            string `xorm:"k TEXT " json:"k"`
	V            string `xorm:"v TEXT " json:"v"`
	VType        string `xorm:"v_type TEXT " json:"v_type"`
	CreatorKey   string `xorm:"creator_key TEXT " json:"creator_key"`
	CreatedUTC   int    `xorm:"created_utc INT " json:"created_utc"`
	LastUpdateId string `xorm:"last_update_id TEXT " json:"last_update_id"`
	UpdateTimes  int    `xorm:"update_times INT " json:"update_times"`
	Des          string `xorm:"des TEXT " json:"des"`
	Status       int    `xorm:"status INT" json:"status"`

	CreatorName    string               `xorm:"-" json:"creator_name"`
	LastUpdateInfo *ConfigUpdateHistory `xorm:"-" json:"last_update_info"`
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

	var res []*Config
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetConfigsByAppKey(s *Session, appKey string) ([]*Config, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*Config
	if err := s.Where("app_key=?", appKey).OrderBy("k").Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func IsValidConfValueType(typ string) bool {
	return typ == CONF_V_TYPE_CODE ||
		typ == CONF_V_TYPE_FLOAT ||
		typ == CONF_V_TYPE_INT ||
		typ == CONF_V_TYPE_STRING ||
		typ == CONF_V_TYPE_TEMPLATE
}

func IsValidConfStatus(status int) bool {
	return status == CONF_STATUS_ACTIVE || status == CONF_STATUS_INACTIVE
}

const (
	NODE_TYPE_MASTER = "master"
	NODE_TYPE_SLAVE  = "slave"
)

type Node struct {
	URL            string `xorm:"url TEXT PK " json:"url"`
	NodeURL        string `xorm:"node_url TEXT UNIQUE 'node_node_url'" json:"node_url"`
	Type           string `xorm:"type TEXT " json:"type"`
	CreatedUTC     int    `xorm:"created_utc INT " json:"created_utc"`
	LastCheckUTC   int    `xorm:"last_check_utc INT " json:"last_check_utc"`
	DataVersionStr string `xorm:"data_version TEXT " json:"data_version_str"` // json string to store DataVersion in db

	DataVersion *DataVersion `xorm:"-" json:"data_version"`
	AppVersion  string       `xorm:"-" json:"app_version"`
	//SchemeVersion string       `xorm:"-"`
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

	var res []*Node
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	for _, node := range res {
		node.DataVersion = &DataVersion{}
		json.Unmarshal([]byte(node.DataVersionStr), node.DataVersion)
	}

	return res, nil
}

func GetNodeByURL(s *Session, url string) (*Node, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := &Node{}
	if has, err := s.Where("url=?", url).Get(res); !has || err != nil {
		return nil, err
	}

	res.DataVersion = &DataVersion{}
	json.Unmarshal([]byte(res.DataVersionStr), res.DataVersion)

	return res, nil
}

func IsValidNodeType(typ string) bool {
	return typ == NODE_TYPE_MASTER || typ == NODE_TYPE_SLAVE
}

type DataVersion struct {
	Version int    `xorm:"version INT" json:"version"`
	Sign    string `xorm:"sign TEXT" json:"sign"`
	OldSign string `xorm:"old_sign TEXT" json:"old_sign"`
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

	var res []*DataVersion
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
	CONFIG_UPDATE_KIND_NEW     = "new"
	CONFIG_UPDATE_KIND_UPDATE  = "up"
	CONFIG_UPDATE_KIND_HIDE    = "hide"
	CONFIG_UPDATE_KIND_RECOVER = "recover"
	CONFIG_UPDATE_KIND_DELETE  = "del"
)

type ConfigUpdateHistory struct {
	Id         string `xorm:"id PK TEXT " json:"id"`
	ConfigKey  string `xorm:"config_key TEXT INDEX" json:"config_key"`
	Kind       string `xorm:"kind TEXT " json:"kind"`
	K          string `xorm:"k TEXT " json:"k"`
	OldV       string `xorm:"old_v TEXT " json:"old_v"`
	OldVType   string `xorm:"old_v_type TEXT " json:"old_v_type"`
	NewV       string `xorm:"new_v TEXT " json:"new_v"`
	NewVType   string `xorm:"new_v_type TEXT " json:"new_v_type"`
	UserKey    string `xorm:"user_key TEXT INDEX" json:"user_key"`
	CreatedUTC int    `xorm:"created_utc INT " json:"created_utc"`

	UserName string `xorm:"-" json:"user_name"`
	App      *App   `xorm:"-" json:"app"`
}

func (*ConfigUpdateHistory) TableName() string {
	return "config_update_history"
}

func (m *ConfigUpdateHistory) UniqueCond() (string, []interface{}) {
	return "id=?", []interface{}{m.Id}
}

func GetConfigUpdateHistoryById(s *Session, id string) (*ConfigUpdateHistory, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	res := &ConfigUpdateHistory{}
	has, err := s.Where("id=?", id).Get(res)
	if !has || err != nil {
		return nil, err
	}

	return res, nil
}

func GetConfigUpdateHistory(s *Session, configKey string) ([]*ConfigUpdateHistory, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*ConfigUpdateHistory
	err := s.Where("config_key=?", configKey).OrderBy("created_utc desc").Find(&res)

	return res, err
}

func GetAppConfigUpdateHistory(s *Session, appKey string, page, count int) ([]*ConfigUpdateHistory, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*ConfigUpdateHistory
	err := s.
		Table("config_update_history").
		Join("INNER", "config", "config.key=config_update_history.config_key").
		Where("config.app_key=?", appKey).
		OrderBy("created_utc desc").
		Limit(count, (page-1)*count).
		Find(&res)
	return res, err
}

func GetAppConfigUpdateHistoryCount(s *Session, appKey string) (int, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	count, err := s.
		Join("INNER", "config", "config.key=config_update_history.config_key").
		Where("config.app_key=?", appKey).
		Count(&ConfigUpdateHistory{})
	return int(count), err
}

func GetConfigUpdateHistoryOfUser(s *Session, userKey string, page, count int) ([]*ConfigUpdateHistory, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*ConfigUpdateHistory
	err := s.Where("user_key=?", userKey).OrderBy("created_utc desc").Limit(count, (page-1)*count).Find(&res)
	return res, err
}

func GetConfigUpdateHistoryCountOfUser(s *Session, userKey string) (int, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	count, err := s.Where("user_key=?", userKey).Count(&ConfigUpdateHistory{})
	return int(count), err
}

func GetAllConfigUpdateHistory(s *Session) ([]*ConfigUpdateHistory, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*ConfigUpdateHistory
	err := s.Find(&res)

	return res, err
}

func ClearModeData(s *Session) error {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	sql := "delete from user; delete from app; delete from config; delete from node;update data_version set version=0;delete from config_update_history; delete from web_hook;"
	_, err := s.Exec(sql)

	return err
}

const (
	WEBHOOK_SCOPE_GLOBAL = 0
	WEBHOOK_SCOPE_APP    = 1

	WEBHOOK_AUTH_NONE  = 0
	WEBHOOK_AUTH_BASIC = 1

	WEBHOOK_STATUS_INACTIVE = 0
	WEBHOOK_STATUS_ACTIVE   = 1

	WEBHOOK_TARGET_PUBU  = "pubu"
	WEBHOOK_TARGET_SLACK = "slack"
)

type WebHook struct {
	Key      string `xorm:"key TEXT PK " json:"key"`
	AppKey   string `xorm:"app_key TEXT " json:"app_key"`
	Scope    int    `xorm:"scope INT" json:"scope"`
	Target   string `xorm:"target TEXT " json:"target"`
	URL      string `xorm:"url TEXT " json:"url"`
	AuthType int    `xorm:"auth_type TEXT " json:"auth_type"`
	AuthInfo string `xorm:"auth_info TEXT " json:"auth_info"`
	Status   int    `xorm:"status INT" json:"status"`
}

func (*WebHook) TableName() string {
	return "web_hook"
}

func (m *WebHook) UniqueCond() (string, []interface{}) {
	return "key=?", []interface{}{m.Key}
}

func GetAllWebHooks(s *Session) ([]*WebHook, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*WebHook
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetGlobalWebHooks(s *Session) ([]*WebHook, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*WebHook
	if err := s.Where("scope =?", WEBHOOK_SCOPE_GLOBAL).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetWebHooksByAppKey(s *Session, appKey string) ([]*WebHook, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*WebHook
	if err := s.Where("scope =? and app_key=?", WEBHOOK_SCOPE_APP, appKey).Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}

type ClientReqeustData struct {
	AppKey string `xorm:"app_key TEXT UNIQUE(uix_client_request_data)" json:"app_key"`
	Symbol string `xorm:"symbol TEXT UNIQUE(uix_client_request_data)" json:"symbol"`
	Value  string `xorm:"value TEXT UNIQUE(uix_client_request_data)" json:"value"`
}

func (*ClientReqeustData) TableName() string {
	return "client_request_data"
}

func GetAllClientRequestData(s *Session) ([]*ClientReqeustData, error) {
	if s == nil {
		s = newAutoCloseModelsSession()
	}

	var res []*ClientReqeustData
	if err := s.Find(&res); err != nil {
		return nil, err
	}

	return res, nil
}
