package main

import (
	"strconv"
	"sync"

	"github.com/appwilldev/Instafig/conf"
	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
	"github.com/gin-gonic/gin"
)

var (
	confWriteMux = sync.Mutex{}
)

func genNewDataVersion(old *models.DataVersion) *models.DataVersion {
	return &models.DataVersion{
		Version: old.Version + 1,
		Sign:    utils.GenerateKey(),
		OldSign: old.Sign,
	}
}

func ConfWriteCheck(c *gin.Context) {
	if conf.IsEasyDeployMode() && !conf.IsMasterNode() {
		Error(c, NOT_PERMITTED, "You can not update config data as you connecting to slave node,")
		c.Abort()
	}
}

type newUserData struct {
	Name string `json:"name" binding:"required"`
}

func NewUser(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &newUserData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	memConfMux.RLock()
	if memConfUsersByName[data.Name] != nil {
		Error(c, BAD_REQUEST, "user name already exists: "+data.Name)
		memConfMux.RUnlock()
		return
	}
	memConfMux.RUnlock()

	user := &models.User{
		Name: data.Name,
		Key:  utils.GenerateKey()}

	if _, err := updateUser(user, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(user)
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func updateUser(user *models.User, newDataVersion *models.DataVersion) (*models.User, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	oldUser := memConfUsers[user.Key]
	dataVer := memConfDataVersion
	memConfMux.RUnlock()

	if newDataVersion == nil {
		newDataVersion = genNewDataVersion(dataVer)
	}
	if err := updateNodeDataVersion(s, node, newDataVersion); err != nil {
		s.Rollback()
		return nil, err
	}

	if oldUser == nil {
		if err := models.InsertDBModel(s, user); err != nil {
			s.Rollback()
			return nil, err
		}
	} else {
		if err := models.UpdateDBModel(s, user); err != nil {
			s.Rollback()
			return nil, err
		}
	}

	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfDataVersion = newDataVersion
	if oldUser != nil {
		memConfUsersByName[oldUser.Name] = nil
	}
	memConfUsers[user.Key] = user
	memConfUsersByName[user.Name] = user

	return user, nil
}

func GetUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		Error(c, BAD_REQUEST, "page not number")
		return
	}

	users, err := models.GetUsers(nil, page, 25)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	Success(c, users)
}

type newAppData struct {
	UserKey string `json:"user_key" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"`
}

func NewApp(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &newAppData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if !models.IsValidAppType(data.Type) {
		Error(c, BAD_REQUEST, "unkown app type: "+data.Type)
		return
	}

	memConfMux.RLock()
	for _, app := range memConfAppsByName[data.Name] {
		if app.UserKey == data.UserKey {
			Error(c, BAD_REQUEST, "appname already exists: "+data.Name)
			memConfMux.RUnlock()
			return
		}
	}
	memConfMux.RUnlock()

	app := &models.App{
		Key:     utils.GenerateKey(),
		Name:    data.Name,
		UserKey: data.UserKey,
		Type:    data.Type,
	}
	if _, err := updateApp(app, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(app)
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

type updateAppData struct {
	Key     string `json:"key" binding:"required"`
	UserKey string `json:"user_key" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"`
}

func UpdateApp(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &updateAppData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if !models.IsValidAppType(data.Type) {
		Error(c, BAD_REQUEST, "unkown app type: "+data.Type)
		return
	}

	memConfMux.RLock()
	if memConfApps[data.Key] == nil {
		Error(c, BAD_REQUEST, "app key not exists: "+data.Key)
		memConfMux.RUnlock()
		return
	}

	if memConfApps[data.Key].Type == models.APP_TYPE_TEMPLATE && memConfApps[data.Key].Type != data.Type {
		Error(c, BAD_REQUEST, "can not change template app to real app")
		memConfMux.RUnlock()
		return
	}

	for _, app := range memConfAppsByName[data.Name] {
		if app.UserKey == data.UserKey && app.Key != data.Key {
			Error(c, BAD_REQUEST, "appname already exists: "+data.Name)
			memConfMux.RUnlock()
			return
		}
	}
	memConfMux.RUnlock()

	app := &models.App{
		Key:     data.Key,
		Name:    data.Name,
		UserKey: data.UserKey,
		Type:    data.Type,
	}
	if _, err := updateApp(app, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(app)
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func updateApp(app *models.App, newDataVersion *models.DataVersion) (*models.App, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	oldApp := memConfApps[app.Key]
	dataVer := memConfDataVersion
	memConfMux.RUnlock()

	if newDataVersion == nil {
		newDataVersion = genNewDataVersion(dataVer)
	}
	if err := updateNodeDataVersion(s, node, newDataVersion); err != nil {
		s.Rollback()
		return nil, err
	}

	if oldApp == nil {
		if err := models.InsertDBModel(s, app); err != nil {
			s.Rollback()
			return nil, err
		}
	} else {
		if err := models.UpdateDBModel(s, app); err != nil {
			s.Rollback()
			return nil, err
		}
	}

	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfDataVersion = newDataVersion
	if oldApp != nil {
		apps := memConfAppsByName[oldApp.Name]
		memConfAppsByName[oldApp.Name] = make([]*models.App, 0)
		for _, app := range apps {
			if app.Key != oldApp.Key {
				memConfAppsByName[oldApp.Name] = append(memConfAppsByName[oldApp.Name], app)
			}
		}
	}
	memConfApps[app.Key] = app
	memConfAppsByName[app.Name] = append(memConfAppsByName[app.Name], app)

	return app, nil
}

func GetApps(c *gin.Context) {
	userKey := c.Param("user_key")
	apps, err := models.GetAppsByUserKey(nil, userKey)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	Success(c, apps)
}

type newConfigData struct {
	AppKey string `json:"app_key" binding:"required"`
	K      string `json:"k" binding:"required"`
	V      string `json:"v" binding:"required"`
	VType  string `json:"v_type" binding:"required"`
}

func NewConfig(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &newConfigData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if !models.IsValidConfType(data.VType) {
		Error(c, BAD_REQUEST, "unkown conf type: "+data.VType)
		return
	}

	if data.VType == models.APP_TYPE_TEMPLATE {
		memConfMux.RLock()
		app := memConfApps[data.V]
		memConfMux.RUnlock()
		if app == nil {
			Error(c, BAD_REQUEST, "template not found for: "+data.V)
			return
		}
		if app.Type != models.APP_TYPE_TEMPLATE {
			Error(c, BAD_REQUEST, "can not set a template conf that is a real app")
			return
		}
	}

	if data.VType == models.CONF_V_TYPE_CODE {
		if _, err := JsonToSexpString(data.V); err != nil {
			Error(c, BAD_REQUEST, "syntax error for code type value: "+err.Error())
			return
		}
	}

	memConfMux.RLock()
	app := memConfApps[data.AppKey]
	configs := getAppMemConfig(data.AppKey)
	memConfMux.RUnlock()

	if app == nil {
		Error(c, BAD_REQUEST, "app key not exists: "+data.AppKey)
		return
	}
	for _, config := range configs {
		if config.K == data.K {
			Error(c, BAD_REQUEST, "config key has existed: "+data.K)
			return
		}
	}

	config := &models.Config{
		Key:    utils.GenerateKey(),
		AppKey: data.AppKey,
		K:      data.K,
		V:      data.V,
		VType:  data.VType,
	}

	config, err := updateConfig(config, nil)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(config)
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

type updateConfigData struct {
	Key    string `json:"key" binding:"required"`
	AppKey string `json:"app_key" binding:"required"`
	K      string `json:"k" binding:"required"`
	V      string `json:"v" binding:"required"`
	VType  string `json:"v_type" binding:"required"`
}

func UpdateConfig(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &updateConfigData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if !models.IsValidConfType(data.VType) {
		Error(c, BAD_REQUEST, "unkown conf type: "+data.VType)
		return
	}

	if data.VType == models.APP_TYPE_TEMPLATE {
		memConfMux.RLock()
		app := memConfApps[data.V]
		memConfMux.RUnlock()
		if app == nil {
			Error(c, BAD_REQUEST, "template not found for: "+data.V)
			return
		}
		if app.Type != models.APP_TYPE_TEMPLATE {
			Error(c, BAD_REQUEST, "can not set a template conf that is a real app")
			return
		}
	}

	if data.VType == models.CONF_V_TYPE_CODE {
		if _, err := JsonToSexpString(data.V); err != nil {
			Error(c, BAD_REQUEST, "syntax error for code type value: "+err.Error())
			return
		}
	}

	memConfMux.RLock()
	oldConfig := memConfRawConfigs[data.Key]
	memConfMux.RUnlock()

	if oldConfig == nil {
		Error(c, BAD_REQUEST, "config key not exists: "+data.Key)
		return
	}
	if oldConfig.AppKey != data.AppKey {
		Error(c, BAD_REQUEST, "can not change config's app key")
		return
	}

	config := &models.Config{
		Key:    data.Key,
		AppKey: data.AppKey,
		K:      data.K,
		V:      data.V,
		VType:  data.VType,
	}

	config, err := updateConfig(config, nil)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(config)
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func updateConfig(config *models.Config, newDataVersion *models.DataVersion) (*models.Config, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	oldConfig := memConfRawConfigs[config.Key]
	ver := memConfDataVersion
	app := *memConfApps[config.AppKey]
	memConfMux.RUnlock()

	if newDataVersion == nil {
		newDataVersion = genNewDataVersion(ver)
	}
	if err := updateNodeDataVersion(s, node, newDataVersion); err != nil {
		s.Rollback()
		return nil, err
	}

	if oldConfig == nil {
		if err := models.InsertDBModel(s, config); err != nil {
			s.Rollback()
			return nil, err
		}
	} else {
		if err := models.UpdateDBModel(s, config); err != nil {
			s.Rollback()
			return nil, err
		}
	}

	app.DataSign = utils.GenerateKey()
	if err := models.UpdateDBModel(s, &app); err != nil {
		s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfDataVersion = newDataVersion
	memConfRawConfigs[config.Key] = config
	*memConfApps[config.AppKey] = app
	if oldConfig == nil {
		memConfAppConfigs[config.AppKey] = append(memConfAppConfigs[config.AppKey], transConfig(config))
	} else {
		for ix, _config := range memConfAppConfigs[config.AppKey] {
			if config.Key == _config.Key {
				memConfAppConfigs[config.AppKey][ix] = transConfig(config)
				break
			}
		}
	}

	return config, nil
}

func GetConfigs(c *gin.Context) {
	appKey := c.Param("app_key")
	configs, err := models.GetConfigsByAppKey(nil, appKey)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	Success(c, configs)
}
