package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/appwilldev/Instafig/models"
)

var (
	memConfUsers          map[string]*models.User
	memConfUsersByName    map[string]*models.User
	memConfApps           map[string]*models.App
	memConfAppsByName     map[string]*models.App
	memConfGlobalWebHooks []*models.WebHook
	memConfAppWebHooks    map[string][]*models.WebHook
	memConfRawConfigs     map[string]*models.Config
	memConfAppConfigs     map[string][]*Config
	memConfNodes          map[string]*models.Node
	memConfDataVersion    *models.DataVersion

	memConfMux = sync.RWMutex{}
)

func loadAllData() {
	users, err := models.GetAllUser(nil)
	if err != nil {
		log.Panicf("Failed to load user info: %s", err.Error())
	}

	apps, err := models.GetAllApps(nil)
	if err != nil {
		log.Panicf("Failed to load app info: %s", err.Error())
	}

	webHooks, err := models.GetAllWebHooks(nil)
	if err != nil {
		log.Panicf("Failed to load webHook info: %s", err.Error())
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

	fillMemConfData(users, apps, webHooks, configs, nodes, dataVersion)
}

func fillMemConfData(
	users []*models.User, apps []*models.App,
	webHooks []*models.WebHook, configs []*models.Config,
	nodes []*models.Node, dataVersion *models.DataVersion) {
	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfUsers = make(map[string]*models.User)
	memConfUsersByName = make(map[string]*models.User)
	memConfApps = make(map[string]*models.App)
	memConfAppsByName = make(map[string]*models.App)
	memConfRawConfigs = make(map[string]*models.Config)
	memConfAppConfigs = make(map[string][]*Config)
	memConfNodes = make(map[string]*models.Node)
	memConfAppWebHooks = make(map[string][]*models.WebHook)
	memConfDataVersion = dataVersion

	for _, user := range users {
		memConfUsers[user.Key] = user
		memConfUsersByName[user.Name] = user
	}

	for _, app := range apps {
		memConfApps[app.Key] = app
		memConfAppsByName[app.Name] = app
		memConfAppConfigs[app.Key] = make([]*Config, 0)
	}

	for _, hook := range webHooks {
		switch hook.Scope {
		case models.WEBHOOK_SCOPE_GLOBAL:
			memConfGlobalWebHooks = append(memConfGlobalWebHooks, hook)
		case models.WEBHOOK_SCOPE_APP:
			memConfAppWebHooks[hook.AppKey] = append(memConfAppWebHooks[hook.AppKey], hook)
		}
	}

	for _, config := range configs {
		memConfRawConfigs[config.Key] = config
		memConfAppConfigs[config.AppKey] = append(memConfAppConfigs[config.AppKey], transConfig(config))
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
