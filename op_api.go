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

	if _, err := newUser(data); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	if conf.IsEasyDeployMode() {
		//TODO: sync to slave
	}

	Success(c, nil)
}

func newUser(newData *newUserData) (*models.User, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	memConfMux.RUnlock()

	if err := updateNodeDataVersion(s, node, memConfDataVersion+1); err != nil {
		s.Rollback()
		return nil, err
	}

	user := &models.User{
		Name: newData.Name,
		Key:  utils.GenerateKey()}
	if err := models.InsertDBModel(s, user); err != nil {
		s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfDataVersion++
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

	if _, err := newApp(data); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	if conf.IsEasyDeployMode() {
		//TODO: sync to slave
	}

	Success(c, nil)
}

func newApp(newData *newAppData) (*models.App, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	memConfMux.RUnlock()

	if err := updateNodeDataVersion(s, node, memConfDataVersion+1); err != nil {
		s.Rollback()
		return nil, err
	}

	app := &models.App{
		Key:     utils.GenerateKey(),
		Name:    newData.Name,
		UserKey: newData.UserKey,
		Type:    newData.Type,
	}
	if err := models.InsertDBModel(s, app); err != nil {
		return nil, err
	}
	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfDataVersion++
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

	memConfMux.RLock()
	app := memConfApps[data.V]
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

	if _, err := newConfig(data); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	if conf.IsEasyDeployMode() {
		//TODO: sync to slave
	}

	Success(c, nil)
}

func newConfig(newData *newConfigData) (*models.Config, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	memConfMux.RUnlock()

	if err := updateNodeDataVersion(s, node, memConfDataVersion+1); err != nil {
		s.Rollback()
		return nil, err
	}

	config := &models.Config{
		Key:    utils.GenerateKey(),
		AppKey: newData.AppKey,
		K:      newData.K,
		V:      newData.V,
		VType:  newData.VType,
	}
	if err := models.InsertDBModel(s, config); err != nil {
		s.Rollback()
		return nil, err
	}
	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	memConfMux.Lock()
	defer memConfMux.Unlock()

	memConfDataVersion++
	memConfRawConfigs[config.Key] = config
	memConfConfigs[config.Key] = transConfig(config)
	memConfAppConfigs[config.AppKey] = append(memConfAppConfigs[config.AppKey], transConfig(config))

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
