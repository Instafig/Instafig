package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/appwilldev/Instafig/models"
)

type ClientData struct {
	AppKey     string `json:"app_key" binding:"required"`
	OSType     string `json:"os_type" binding:"required"`
	OSVersion  string `json:"os_version" binding:"required"`
	AppVersion string `json:"app_version" binding:"required"`
	Ip         string `json:"ip" binding:"required"`
	Lang       string `json:"lang" binding:"required"`
	DeviceId   string `json:"device_id"`
}

type Config struct {
	ConfKey string
	Key     string
	Val     interface{}
	ValType string
}

var (
	memConfUsers      = make(map[string]*models.User)
	memConfUsersByName      = make(map[string]*models.User)
	memConfApps       = make(map[string]*models.App)
	memConfConfigs    = make(map[string]*Config)
	memConfAppConfigs = make(map[string][]*Config)
	memConfNodes      = make(map[string]*models.Node)

	memConfMux = sync.RWMutex{}
)

func init() {
	loadData()
}

func getNodeKey(node *models.Node) string {
	return fmt.Sprintf("%s:%d", node.Host, node.Port)
}

func loadData() {
	users, err := models.GetAllUser(nil)
	if err != nil {
		log.Panicf("Failed to load user info: %s", err.Error())
	}

	apps, err := models.GetAllApp(nil)
	if err != nil {
		log.Panicf("Failed to load app info: %s", err.Error())
	}

	configs, err := models.GetAllConfig(nil)
	if err != nil {
		log.Panicf("Failed to load config info: %s", err.Error())
	}

	nodes, err := models.GetAllNode(nil)
	if err != nil {
		log.Panicf("Failed to load node info: %s", err.Error())
	}

	fillMemConfData(users, apps, configs, nodes)
}

func fillMemConfData(users []*models.User, apps []*models.App, configs []*models.Config, nodes []*models.Node) {
	memConfMux.Lock()
	defer memConfMux.Unlock()

	for _, user := range users {
		memConfUsers[user.Key] = user
		memConfUsersByName[user.Name] = user
	}

	for _, app := range apps {
		memConfApps[app.Key] = app
		memConfAppConfigs[app.Key] = make([]*Config, 0)
	}

	for _, config := range configs {
		memConfConfigs[config.Key] = &Config{
			ConfKey: config.Key,
			Key:     config.K,
			ValType: config.VType,
		}

		switch config.VType {
		case models.CONF_V_TYPE_FLOAT:
			memConfConfigs[config.Key].Val, _ = strconv.ParseFloat(config.V, 64)
		case models.CONF_V_TYPE_INT:
			memConfConfigs[config.Key].Val, _ = strconv.Atoi(config.V)
		case models.CONF_V_TYPE_STRING:
			memConfConfigs[config.Key].Val = config.V
		case models.CONF_V_TYPE_CODE:
			//TODO: trans to callable object
			memConfConfigs[config.Key].Val = config.V
		}

		memConfAppConfigs[config.AppKey] = append(memConfAppConfigs[config.AppKey], memConfConfigs[config.Key])
	}

	for _, node := range nodes {
		memConfNodes[getNodeKey(node)] = node
	}
}

// read only, DO NOT change field value
func getAppMemConfig(appKey string) []*Config {
	memConfMux.RLock()
	defer memConfMux.RUnlock()

	return memConfAppConfigs[appKey]
}

func getMatchConf(matchData *ClientData, configs []*Config) map[string]interface{} {
	res := make(map[string]interface{}, 0)
	for _, config := range configs {
		switch config.ValType {
		case models.CONF_V_TYPE_CODE:
			res[config.Key] = EvalDynVal(config.Val.(string), matchData)
			continue
		default:
			res[config.Key] = config.Val
		}
	}

	return res
}

func getAppMatchConf(appKey string, matchData *ClientData) map[string]interface{} {
	appConfigs := getAppMemConfig(appKey)
	if appConfigs == nil {
		return nil
	}

	return getMatchConf(matchData, appConfigs)
}
