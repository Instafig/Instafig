package main

import (
	"log"
	"strconv"
	"sync"

	"encoding/json"

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
	Key    string
	AppKey string
	K      string
	V      interface{}
	VType  string
}

var (
	memConfUsers       map[string]*models.User
	memConfUsersByName map[string]*models.User
	memConfApps        map[string]*models.App
	memConfAppsByName  map[string][]*models.App
	memConfRawConfigs  map[string]*models.Config
	memConfAppConfigs  map[string][]*Config
	memConfNodes       map[string]*models.Node
	memConfDataVersion *models.DataVersion

	memConfMux = sync.RWMutex{}
)

func transConfig(m *models.Config) *Config {
	config := &Config{
		Key:    m.Key,
		AppKey: m.AppKey,
		K:      m.K,
		VType:  m.VType,
	}

	switch m.VType {
	case models.CONF_V_TYPE_FLOAT:
		config.V, _ = strconv.ParseFloat(m.V, 64)
	case models.CONF_V_TYPE_INT:
		config.V, _ = strconv.Atoi(m.V)
	case models.CONF_V_TYPE_STRING:
		config.V = m.V
	case models.CONF_V_TYPE_CODE:
		// TODO: trans to callable object
		config.V, _ = JsonToSexpString(m.V)
	case models.CONF_V_TYPE_TEMPLATE:
		config.V = m.V
	}

	return config
}

func loadAllData() {
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

	dataVersion, err := models.GetDataVersion(nil)
	if err != nil {
		log.Panicf("Failed to load data version info: %s", err.Error())
	}

	fillMemConfData(users, apps, configs, nodes, dataVersion)
}

func fillMemConfData(users []*models.User, apps []*models.App, configs []*models.Config, nodes []*models.Node, dataVersion *models.DataVersion) {
	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfUsers = make(map[string]*models.User)
	memConfUsersByName = make(map[string]*models.User)
	memConfApps = make(map[string]*models.App)
	memConfAppsByName = make(map[string][]*models.App)
	memConfRawConfigs = make(map[string]*models.Config)
	memConfAppConfigs = make(map[string][]*Config)
	memConfNodes = make(map[string]*models.Node)
	memConfDataVersion = dataVersion

	for _, user := range users {
		memConfUsers[user.Key] = user
		memConfUsersByName[user.Name] = user
	}

	for _, app := range apps {
		memConfApps[app.Key] = app
		memConfAppsByName[app.Name] = append(memConfAppsByName[app.Name], app)
		memConfAppConfigs[app.Key] = make([]*Config, 0)
	}

	for _, config := range configs {
		c := transConfig(config)
		memConfRawConfigs[config.Key] = config
		memConfAppConfigs[config.AppKey] = append(memConfAppConfigs[config.AppKey], c)
	}

	for _, node := range nodes {
		memConfNodes[node.URL] = node
		node.DataVersion = &models.DataVersion{}
		json.Unmarshal([]byte(node.DataVersionStr), node.DataVersion)
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
		switch config.VType {
		case models.CONF_V_TYPE_CODE:
			res[config.K] = EvalDynVal(config.V.(string), matchData)
		case models.CONF_V_TYPE_TEMPLATE:
			res[config.K] = getAppMatchConf(config.V.(string), matchData)
		default:
			res[config.K] = config.V
		}
	}

	return res
}

func getAppMatchConf(appKey string, clientData *ClientData) map[string]interface{} {
	appConfigs := getAppMemConfig(appKey)
	if appConfigs == nil {
		return nil
	}

	return getMatchConf(clientData, appConfigs)
}
