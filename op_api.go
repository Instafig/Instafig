package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/appwilldev/Instafig/conf"
	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
	"github.com/dgrijalva/jwt-go"
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

func UpdateMasterLastDataUpdateUTC(c *gin.Context) {
	if !conf.IsEasyDeployMode() || !getServiceStatus(c) ||
		(c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch && c.Request.Method != http.MethodDelete) {
		return
	}
	memConfMux.RLock()
	localNode := *memConfNodes[conf.ClientAddr] // must be master node
	memConfMux.RUnlock()

	// use master node's LastCheckUTC to store last update utc
	localNode.LastCheckUTC = utils.GetNowSecond()
	if models.UpdateDBModel(nil, &localNode) == nil {
		memConfMux.Lock()
		memConfNodes[localNode.URL] = &localNode
		memConfMux.Unlock()
	}
}

func Login(c *gin.Context) {
	var data struct {
		Name     string `json:"name" binding:"required"`
		PassCode string `json:"pass_code" binding:"required"`
	}
	if err := c.BindJSON(&data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	memConfMux.RLock()
	user := memConfUsersByName[data.Name]
	memConfMux.RUnlock()

	if user == nil {
		Error(c, USER_NOT_EXIST)
		return
	}
	if user.PassCode != encryptUserPassCode(data.PassCode) {
		Error(c, PASS_CODE_ERROR)
		return
	}

	if user.Status == models.USER_STATUS_INACTIVE {
		Error(c, USER_INACTIVE)
		return
	}

	setUserKeyCookie(c, user.Key, user.PassCode)
	Success(c, nil)
}

func Logout(c *gin.Context) {
	deleteUserKeyCookie(c)
	Success(c, nil)
}

type newUserData struct {
	Name     string `json:"name" binding:"required"`
	PassCode string `json:"pass_code" binding:"required"`
	AuxInfo  string `json:"aux_info"`
}

func InitUser(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	if len(memConfUsers) > 0 {
		Error(c, BAD_REQUEST, "users already exists")
		return
	}

	data := &newUserData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if err := verifyNewUserData(data); err != nil {
		Error(c, BAD_REQUEST, err.Error())
		return
	}

	key := utils.GenerateKey()
	user, err := newUserWithNewUserData(data, key, key)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(user, key)
	setUserKeyCookie(c, user.Key, user.PassCode)
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func NewUser(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &newUserData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if err := verifyNewUserData(data); err != nil {
		Error(c, BAD_REQUEST, err.Error())
		return
	}

	user, err := newUserWithNewUserData(data, utils.GenerateKey(), getOpUserKey(c))
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(user, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func verifyNewUserData(data *newUserData) error {
	if memConfUsersByName[data.Name] != nil {
		return fmt.Errorf("user [%s] already exists", data.Name)
	}

	if len(data.Name) < 3 {
		return fmt.Errorf("user name too short, length must bigger than 2")
	}

	if len(data.PassCode) < 6 {
		return fmt.Errorf("user passcode too short, length must bigger than 6")
	}

	return nil
}

func newUserWithNewUserData(data *newUserData, userKey, creatorKey string) (*models.User, error) {
	user := &models.User{
		Name:       data.Name,
		PassCode:   encryptUserPassCode(data.PassCode),
		CreatorKey: creatorKey,
		CreatedUTC: utils.GetNowSecond(),
		AuxInfo:    data.AuxInfo,
		Status:     models.USER_STATUS_ACTIVE,
		Key:        userKey}

	return updateUser(user, nil)
}

type updateUserData struct {
	Name    string `json:"name" binding:"required"`
	AuxInfo string `json:"aux_info"`
}

func UpdateUser(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &updateUserData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if err := verifyUpdateUserData(data, getOpUserKey(c)); err != nil {
		Error(c, BAD_REQUEST, err.Error())
		return
	}

	oldUser := memConfUsers[getOpUserKey(c)]
	if oldUser.AuxInfo == data.AuxInfo && oldUser.Name == data.Name {
		Success(c, nil)
		return
	}

	user, err := updateUserWithUpdateData(data, getOpUserKey(c))
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(user, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func verifyUpdateUserData(data *updateUserData, userKey string) error {
	if memConfUsersByName[data.Name] != nil && memConfUsersByName[data.Name].Key != memConfUsers[userKey].Key {
		return fmt.Errorf("user name [%s] already exists", data.Name)
	}

	if len(data.Name) < 3 {
		return fmt.Errorf("user name too short, length must bigger than 2")
	}

	return nil
}

func updateUserWithUpdateData(data *updateUserData, userKey string) (*models.User, error) {
	user := *memConfUsers[userKey]
	user.AuxInfo = data.AuxInfo
	user.Name = data.Name
	return updateUser(&user, nil)
}

func UpdateUserPassCode(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	var data struct {
		PassCode string `json:"pass_code" binding:"required"`
	}
	if err := c.BindJSON(&data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	user := *memConfUsers[getOpUserKey(c)]
	user.PassCode = encryptUserPassCode(data.PassCode)

	if _, err := updateUser(&user, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	setUserKeyCookie(c, user.Key, user.PassCode)
	failedNodes := syncData2SlaveIfNeed(&user, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func UpdateUserStatus(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	var data struct {
		Status  int    `json:"status"`
		UserKey string `json:"user_key" binding:"required"`
	}
	if err := c.BindJSON(&data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	user := *memConfUsers[data.UserKey]
	opUser := memConfUsers[getOpUserKey(c)]
	if opUser.CreatorKey != opUser.Key {
		Error(c, NOT_PERMITTED, "can not update user's status as current user is non-init user")
		return
	}
	if user.Key == opUser.Key {
		Error(c, NOT_PERMITTED, "can not update current login user's status")
		return
	}

	user.Status = data.Status
	if _, err := updateUser(&user, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}
	failedNodes := syncData2SlaveIfNeed(&user, getOpUserKey(c))
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

	node := *memConfNodes[conf.ClientAddr]
	oldUser := memConfUsers[user.Key]

	if newDataVersion == nil {
		newDataVersion = genNewDataVersion(memConfDataVersion)
	}
	if err := updateNodeDataVersion(s, &node, newDataVersion); err != nil {
		s.Rollback()
		return nil, err
	}

	if oldUser == nil {
		if err := models.InsertRow(s, user); err != nil {
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

	updateMemConf(user, newDataVersion, &node)

	return user, nil
}

func GetUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		Error(c, BAD_REQUEST, "page not number")
		return
	}
	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		Error(c, BAD_REQUEST, "count not number")
		return
	}

	users, err := models.GetUsers(nil, page, count)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	totalCount, err := models.GetUserCount(nil)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	memConfMux.RLock()
	for _, user := range users {
		user.PassCode = ""
		if memConfUsers[user.CreatorKey] != nil {
			user.CreatorName = memConfUsers[user.CreatorKey].Name
		}
	}
	memConfMux.RUnlock()

	Success(c, map[string]interface{}{
		"total_count": totalCount,
		"list":        users,
	})
}

func GetApp(c *gin.Context) {
	memConfMux.RLock()
	app := memConfApps[c.Param("app_key")]
	memConfMux.RUnlock()

	if app == nil {
		Success(c, nil)
		return
	}

	returnApp := *app
	returnApp.LastUpdateInfo, _ = models.GetConfigUpdateHistoryById(nil, returnApp.LastUpdateId)
	memConfMux.RLock()
	returnApp.UserName = memConfUsers[returnApp.UserKey].Name
	memConfMux.RUnlock()

	Success(c, &returnApp)
}

func SearchApps(c *gin.Context) {
	apps, err := searchApps(c.Query("q"), 0)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	memConfMux.RLock()
	for _, app := range apps {
		app.UserName = memConfUsers[app.UserKey].Name
		app.LastUpdateInfo, _ = models.GetConfigUpdateHistoryById(nil, app.LastUpdateId)
		if app.LastUpdateInfo != nil {
			app.LastUpdateInfo.UserName = memConfUsers[app.LastUpdateInfo.UserKey].Name
		}
	}
	memConfMux.RUnlock()

	Success(c, apps)
}

func SearchAppsHint(c *gin.Context) {
	apps, err := searchApps(c.Query("q"), 10)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	res := make([]map[string]string, len(apps))
	memConfMux.RLock()
	for ix, app := range apps {
		_res := map[string]string{}
		_res["name"] = app.Name
		_res["aux_info"] = app.AuxInfo
		_res["key"] = app.Key
		res[ix] = _res
	}
	memConfMux.RUnlock()

	Success(c, res)
}

func searchApps(q string, count int) ([]*models.App, error) {
	return models.SearchAppByName(nil, q, count)
}

func NewApp(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	var data struct {
		Name    string `json:"name" binding:"required"`
		Type    string `json:"type" binding:"required"`
		AuxInfo string `json:"aux_info"`
	}
	if err := c.BindJSON(&data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if !models.IsValidAppType(data.Type) {
		Error(c, BAD_REQUEST, "unknown app type: "+data.Type)
		return
	}

	if conf.IsEasyDeployMode() && memConfAppsByName[data.Name] != nil {
		Error(c, BAD_REQUEST, "appname already exists: "+data.Name)
		return
	}

	app := &models.App{
		Key:        utils.GenerateKey(),
		Name:       data.Name,
		UserKey:    getOpUserKey(c),
		Type:       data.Type,
		AuxInfo:    data.AuxInfo,
		CreatedUTC: utils.GetNowSecond(),
	}
	if _, err := updateApp(app, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(app, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func UpdateApp(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	var data struct {
		Key     string `json:"key" binding:"required"`
		Name    string `json:"name" binding:"required"`
		Type    string `json:"type" binding:"required"`
		AuxInfo string `json:"aux_info"`
	}
	if err := c.BindJSON(&data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if !models.IsValidAppType(data.Type) {
		Error(c, BAD_REQUEST, "unknown app type: "+data.Type)
		return
	}

	oldApp := memConfApps[data.Key]
	if oldApp == nil {
		Error(c, BAD_REQUEST, "app key not exists: "+data.Key)
		return
	}

	if oldApp.Name == data.Name && oldApp.AuxInfo == data.AuxInfo && oldApp.Type == data.Type {
		Success(c, nil)
		return
	}

	if oldApp.Type == models.APP_TYPE_TEMPLATE && oldApp.Type != data.Type {
		Error(c, BAD_REQUEST, "can not change template app to real app")
		return
	}

	if conf.IsEasyDeployMode() && memConfApps[data.Key].Name != data.Name {
		for _, app := range memConfApps {
			if app.Name == data.Name {
				Error(c, BAD_REQUEST, "appname already exists: "+data.Name)
				return
			}
		}
	}

	app := *oldApp
	app.Name = data.Name
	app.AuxInfo = data.AuxInfo
	if _, err := updateApp(&app, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(&app, getOpUserKey(c))
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

	node := *memConfNodes[conf.ClientAddr]
	oldApp := memConfApps[app.Key]

	if newDataVersion == nil {
		newDataVersion = genNewDataVersion(memConfDataVersion)
	}
	if err := updateNodeDataVersion(s, &node, newDataVersion); err != nil {
		s.Rollback()
		return nil, err
	}

	if oldApp == nil {
		if err := models.InsertRow(s, app); err != nil {
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

	updateMemConf(app, newDataVersion, &node)

	return app, nil
}

func GetApps(c *gin.Context) {
	userKey := c.Param("user_key")
	apps, err := models.GetAppsByUserKey(nil, userKey)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	memConfMux.RLock()
	for _, app := range apps {
		app.UserName = memConfUsers[app.UserKey].Name
		app.LastUpdateInfo, _ = models.GetConfigUpdateHistoryById(nil, app.LastUpdateId)
		if app.LastUpdateInfo != nil {
			app.LastUpdateInfo.UserName = memConfUsers[app.LastUpdateInfo.UserKey].Name
		}
	}
	memConfMux.RUnlock()

	Success(c, apps)
}

func GetAllApps(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		Error(c, BAD_REQUEST, "page not number")
		return
	}
	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		Error(c, BAD_REQUEST, "count not number")
		return
	}

	apps, err := models.GetAllAppsPage(nil, page, count)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	totalCount, err := models.GetAppCount(nil)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	memConfMux.RLock()
	for _, app := range apps {
		app.UserName = memConfUsers[app.UserKey].Name
		app.LastUpdateInfo, _ = models.GetConfigUpdateHistoryById(nil, app.LastUpdateId)
		if app.LastUpdateInfo != nil {
			app.LastUpdateInfo.UserName = memConfUsers[app.LastUpdateInfo.UserKey].Name
		}
	}
	memConfMux.RUnlock()

	Success(c, map[string]interface{}{
		"total_count": totalCount,
		"list":        apps,
	})
}

func updateWebHook(hook *models.WebHook, newDataVersion *models.DataVersion) (*models.WebHook, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	node := *memConfNodes[conf.ClientAddr]
	oldHookIdx := -1
	var hooks []*models.WebHook
	if hook.Scope == models.WEBHOOK_SCOPE_GLOBAL {
		hooks = memConfGlobalWebHooks
	} else if hook.Scope == models.WEBHOOK_SCOPE_APP {
		hooks = memConfAppWebHooks[hook.AppKey]
	}
	for idx, oldHook := range hooks {
		if hook.Key == oldHook.Key {
			oldHookIdx = idx
			break
		}
	}

	if newDataVersion == nil {
		newDataVersion = genNewDataVersion(memConfDataVersion)
	}
	if err := updateNodeDataVersion(s, &node, newDataVersion); err != nil {
		s.Rollback()
		return nil, err
	}

	if oldHookIdx == -1 {
		if err := models.InsertRow(s, hook); err != nil {
			s.Rollback()
			return nil, err
		}
	} else {
		if err := models.UpdateDBModel(s, hook); err != nil {
			s.Rollback()
			return nil, err
		}
	}

	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	updateMemConf(hook, newDataVersion, &node, oldHookIdx)

	return hook, nil
}

type newConfigData struct {
	AppKey string `json:"app_key" binding:"required"`
	K      string `json:"k" binding:"required"`
	V      string `json:"v"`
	VType  string `json:"v_type" binding:"required"`
	Des    string `json:"des"`
}

func NewConfig(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &newConfigData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if err := verifyNewConfigData(data); err != nil {
		Error(c, BAD_REQUEST, err.Error())
		return
	}

	config, err := newConfigWithNewConfigData(data, getOpUserKey(c))
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(config, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func verifyNewConfigData(data *newConfigData) error {
	if !models.IsValidConfValueType(data.VType) {
		return fmt.Errorf("unknown conf type: " + data.VType)
	}

	switch data.VType {
	case models.CONF_V_TYPE_CODE:
		if err := CheckJsonString(data.V); err != nil {
			return fmt.Errorf("syntax error for code type value: " + err.Error())
		}
	case models.CONF_V_TYPE_FLOAT:
		if _, err := strconv.ParseFloat(data.V, 64); err != nil {
			return fmt.Errorf("config Value not float")
		}
	case models.CONF_V_TYPE_INT:
		if _, err := strconv.ParseInt(data.V, 10, 64); err != nil {
			return fmt.Errorf("config Value not int")
		}
	case models.APP_TYPE_TEMPLATE:
		app := memConfApps[data.V]
		if app == nil {
			return fmt.Errorf("template not found for: " + data.V)
		}
		if app.Type != models.APP_TYPE_TEMPLATE {
			return fmt.Errorf("can not set a template conf that is a real app")
		}
	case models.CONF_V_TYPE_STRING:
	// no need check
	default:
		return fmt.Errorf("unknown config value type: " + data.VType)
	}

	if memConfApps[data.AppKey] == nil {
		return fmt.Errorf("app key not exists: " + data.AppKey)
	}

	for _, config := range memConfAppConfigs[data.AppKey] {
		if config.K == data.K {
			return fmt.Errorf("config key has existed: " + data.K)
		}
	}

	return nil
}

func newConfigWithNewConfigData(data *newConfigData, userKey string) (*models.Config, error) {
	config := &models.Config{
		Key:        utils.GenerateKey(),
		AppKey:     data.AppKey,
		K:          data.K,
		V:          data.V,
		VType:      data.VType,
		CreatedUTC: utils.GetNowSecond(),
		CreatorKey: userKey,
		Des:        data.Des,
		Status:     models.CONF_STATUS_ACTIVE,
	}

	return updateConfig(config, userKey, nil)
}

type updateConfigData struct {
	Key    string `json:"key" binding:"required"`
	K      string `json:"k" binding:"required"`
	V      string `json:"v"`
	VType  string `json:"v_type" binding:"required"`
	Des    string `json:"des"`
	Status int    `json:"status"`
}

func UpdateConfig(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &updateConfigData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if err := verifyUpdateConfigData(data); err != nil {
		Error(c, BAD_REQUEST, err.Error())
		return
	}

	oldConfig := memConfRawConfigs[data.Key]
	if oldConfig.K == data.K && oldConfig.V == data.V && oldConfig.VType == data.VType && oldConfig.Des == data.Des && oldConfig.Status == data.Status {
		Success(c, nil)
		return
	}

	config, err := updateConfigWithUpdateData(data, getOpUserKey(c))
	if err != nil {
		Error(c, SERVER_ERROR, err)
		return
	}

	failedNodes := syncData2SlaveIfNeed(config, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func verifyUpdateConfigData(data *updateConfigData) error {
	if !models.IsValidConfValueType(data.VType) {
		return fmt.Errorf("unknown conf type: " + data.VType)
	}
	if !models.IsValidConfStatus(data.Status) {
		return fmt.Errorf("unknown conf status: " + strconv.Itoa(data.Status))
	}

	switch data.VType {
	case models.CONF_V_TYPE_CODE:
		if err := CheckJsonString(data.V); err != nil {
			return fmt.Errorf("syntax error for code type value: " + err.Error())
		}
	case models.CONF_V_TYPE_FLOAT:
		if _, err := strconv.ParseFloat(data.V, 64); err != nil {
			return fmt.Errorf("config Value not float")
		}
	case models.CONF_V_TYPE_INT:
		if _, err := strconv.ParseInt(data.V, 10, 64); err != nil {
			return fmt.Errorf("config Value not int")
		}
	case models.APP_TYPE_TEMPLATE:
		app := memConfApps[data.V]
		if app == nil {
			return fmt.Errorf("template not found for: " + data.V)
		}
		if app.Type != models.APP_TYPE_TEMPLATE {
			return fmt.Errorf("can not set a template conf that is a real app")
		}
	case models.CONF_V_TYPE_STRING:
		// no need check
	default:
		return fmt.Errorf("unknown config value type: " + data.VType)
	}

	oldConfig := memConfRawConfigs[data.Key]
	if oldConfig == nil {
		return fmt.Errorf("config key not exists: " + data.Key)
	}
	if oldConfig.K != data.K {
		for _, config := range memConfAppConfigs[oldConfig.AppKey] {
			if config.K == data.K {
				return fmt.Errorf("config [%s] already exists", data.K)
			}
		}
	}

	return nil
}

func updateConfigWithUpdateData(data *updateConfigData, userKey string) (*models.Config, error) {
	config := *memConfRawConfigs[data.Key]
	config.K = data.K
	config.V = data.V
	config.VType = data.VType
	config.Des = data.Des
	config.Status = data.Status

	return updateConfig(&config, userKey, nil)
}

func updateConfig(config *models.Config, userKey string, newDataVersion *models.DataVersion) (*models.Config, error) {
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return nil, err
	}

	node := *memConfNodes[conf.ClientAddr]
	oldConfig := memConfRawConfigs[config.Key]
	app := memConfApps[config.AppKey]

	if newDataVersion == nil {
		newDataVersion = genNewDataVersion(memConfDataVersion)
	}
	if err := updateNodeDataVersion(s, &node, newDataVersion); err != nil {
		s.Rollback()
		return nil, err
	}

	var configHistory *models.ConfigUpdateHistory
	tempApp := *app

	if oldConfig == nil {
		configHistory = &models.ConfigUpdateHistory{
			Id:         utils.GenerateKey(),
			ConfigKey:  config.Key,
			K:          config.K,
			OldV:       "",
			OldVType:   "",
			NewV:       config.V,
			NewVType:   config.VType,
			Kind:       models.CONFIG_UPDATE_KIND_NEW,
			UserKey:    userKey,
			CreatedUTC: utils.GetNowSecond(),
		}
		if err := models.InsertRow(s, configHistory); err != nil {
			s.Rollback()
			return nil, err
		}

		config.LastUpdateId = configHistory.Id
		if err := models.InsertRow(s, config); err != nil {
			s.Rollback()
			return nil, err
		}

		tempApp.KeyCount++
		tempApp.LastUpdateUTC = configHistory.CreatedUTC
		tempApp.LastUpdateId = configHistory.Id
		tempApp.UpdateTimes++
	} else {
		kind := models.CONFIG_UPDATE_KIND_UPDATE
		if config.Status != oldConfig.Status {
			if config.Status == models.CONF_STATUS_ACTIVE {
				kind = models.CONFIG_UPDATE_KIND_RECOVER
			} else {
				kind = models.CONFIG_UPDATE_KIND_HIDE
			}
		}

		configHistory = &models.ConfigUpdateHistory{
			Id:         utils.GenerateKey(),
			ConfigKey:  config.Key,
			K:          config.K,
			OldV:       oldConfig.V,
			OldVType:   oldConfig.VType,
			NewV:       config.V,
			NewVType:   config.VType,
			Kind:       kind,
			UserKey:    userKey,
			CreatedUTC: utils.GetNowSecond(),
		}
		if err := models.InsertRow(s, configHistory); err != nil {
			s.Rollback()
			return nil, err
		}

		config.UpdateTimes++
		config.LastUpdateId = configHistory.Id
		if err := models.UpdateDBModel(s, config); err != nil {
			s.Rollback()
			return nil, err
		}

		tempApp.LastUpdateUTC = configHistory.CreatedUTC
		tempApp.LastUpdateId = configHistory.Id
		tempApp.UpdateTimes++
	}

	if err := models.UpdateDBModel(s, &tempApp); err != nil {
		s.Rollback()
		return nil, err
	}

	var toUpdateApps []*models.App
	if app.Type == models.APP_TYPE_REAL {
		toUpdateApps = append(toUpdateApps, app)
	} else {
		for _, app := range memConfApps {
			if app.Key == config.AppKey {
				toUpdateApps = append(toUpdateApps, app)
				continue
			}
			for _, _config := range memConfAppConfigs[app.Key] {
				if _config.VType == models.CONF_V_TYPE_TEMPLATE && _config.V == config.AppKey {
					// this app has a config refer to this template app
					toUpdateApps = append(toUpdateApps, app)
					break
				}
			}
		}
	}

	newDataSign := utils.GenerateKey()
	for _, app := range toUpdateApps {
		_app := *app
		if app.Key == tempApp.Key {
			_app = tempApp
			tempApp.DataSign = newDataSign
		}
		_app.DataSign = newDataSign
		if err := models.UpdateDBModel(s, &_app); err != nil {
			s.Rollback()
			return nil, err
		}
	}

	if err := s.Commit(); err != nil {
		s.Rollback()
		return nil, err
	}

	go TriggerWebHooks(configHistory, app)

	updateMemConf(config, newDataVersion, &node, toUpdateApps)

	return config, nil
}

func GetConfigs(c *gin.Context) {
	appKey := c.Param("app_key")
	configs, err := models.GetConfigsByAppKey(nil, appKey)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	memConfMux.RLock()
	for _, config := range configs {
		config.CreatorName = memConfUsers[config.CreatorKey].Name
		config.LastUpdateInfo, _ = models.GetConfigUpdateHistoryById(nil, config.LastUpdateId)
		config.LastUpdateInfo.UserName = memConfUsers[config.LastUpdateInfo.UserKey].Name
	}
	memConfMux.RUnlock()

	Success(c, configs)
}

func GetConfigUpdateHistory(c *gin.Context) {
	histories, err := models.GetConfigUpdateHistory(nil, c.Param("config_key"))
	if err != nil {
		Error(c, SERVER_ERROR)
		return
	}

	memConfMux.RLock()
	for _, history := range histories {
		history.UserName = memConfUsers[history.UserKey].Name
	}
	memConfMux.RUnlock()

	Success(c, histories)
}

func GetAppConfigUpdateHistory(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		Error(c, BAD_REQUEST, "page not number")
		return
	}

	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		Error(c, BAD_REQUEST, "count not number")
		return
	}

	histories, err := models.GetAppConfigUpdateHistory(nil, c.Param("app_key"), page, count)
	if err != nil {
		Error(c, SERVER_ERROR)
		return
	}

	if len(histories) == 0 {
		Success(c, map[string]interface{}{
			"total_count": 0,
			"list":        []interface{}{},
		})
		return
	}

	memConfMux.RLock()
	for _, history := range histories {
		history.UserName = memConfUsers[history.UserKey].Name
	}
	memConfMux.RUnlock()

	totalCount, err := models.GetAppConfigUpdateHistoryCount(nil, c.Param("app_key"))
	if err != nil {
		Error(c, SERVER_ERROR)
		return
	}

	Success(c, map[string]interface{}{
		"total_count": totalCount,
		"list":        histories,
	})
}

func GetConfigUpdateHistoryOfUser(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		Error(c, BAD_REQUEST, "page not number")
		return
	}

	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		Error(c, BAD_REQUEST, "count not number")
		return
	}

	histories, err := models.GetConfigUpdateHistoryOfUser(nil, c.Param("user_key"), page, count)
	if err != nil {
		Error(c, SERVER_ERROR)
		return
	}

	if len(histories) == 0 {
		Success(c, map[string]interface{}{
			"total_count": 0,
			"list":        []interface{}{},
		})
		return
	}

	memConfMux.RLock()
	userName := memConfUsers[c.Param("user_key")].Name
	for _, history := range histories {
		history.UserName = userName
		app := memConfApps[memConfRawConfigs[history.ConfigKey].AppKey]
		app.UserName = memConfUsers[app.UserKey].Name
		history.App = app
	}
	memConfMux.RUnlock()

	totalCount, err := models.GetConfigUpdateHistoryCountOfUser(nil, c.Param("user_key"))
	if err != nil {
		Error(c, SERVER_ERROR)
		return
	}

	Success(c, map[string]interface{}{
		"total_count": totalCount,
		"list":        histories,
	})
}

func GetConfigByKey(c *gin.Context) {
	memConfMux.RLock()
	configPtr := memConfRawConfigs[c.Param("config_key")]
	memConfMux.RUnlock()

	if configPtr != nil {
		config := *configPtr
		config.LastUpdateInfo, _ = models.GetConfigUpdateHistoryById(nil, config.LastUpdateId)
		memConfMux.RLock()
		config.LastUpdateInfo.UserName = memConfUsers[config.LastUpdateInfo.UserKey].Name
		memConfMux.RUnlock()
		Success(c, config)
	} else {
		Error(c, BAD_REQUEST, "config not exists")
	}
}

func GetNodes(c *gin.Context) {
	memConfMux.RLock()
	nodes := make([]*models.Node, 0)
	for _, node := range memConfNodes {
		_node := *node
		_node.DataVersionStr = ""
		_node.AppVersion = conf.VersionString()
		nodes = append(nodes, &_node)
	}
	memConfMux.RUnlock()

	Success(c, nodes)
}

func OpAuth(c *gin.Context) {
	cookie, err := c.Request.Cookie("op_user")
	if err != nil {
		Error(c, NOT_LOGIN, err.Error())
		c.Abort()
		return
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(conf.NodeAuth), nil
	})
	if err != nil {
		Error(c, NOT_LOGIN, err.Error())
		c.Abort()
		return
	}

	if !token.Valid {
		Error(c, NOT_LOGIN, "cookie token invalid")
		c.Abort()
		return
	}

	userKey := token.Claims["uky"].(string)
	memConfMux.RLock()
	defer memConfMux.RUnlock()

	user := memConfUsers[userKey]

	if user == nil {
		Error(c, NOT_LOGIN, "user not exist")
		c.Abort()
		return
	}
	if token.Claims["upc"] == nil {
		Error(c, NOT_LOGIN, "pass_code not correct")
		c.Abort()
		return
	}
	passCode := token.Claims["upc"].(string)
	if user.PassCode != passCode {
		Error(c, NOT_LOGIN, "pass_code not correct")
		c.Abort()
		return
	}

	if user.Status == models.USER_STATUS_INACTIVE {
		deleteUserKeyCookie(c)
		Error(c, USER_INACTIVE)
		c.Abort()
		return
	}

	setOpUserKey(c, userKey)
}

func InitUserCheck(c *gin.Context) {
	memConfMux.RLock()
	userCount := len(memConfUsers)
	memConfMux.RUnlock()

	if userCount == 0 {
		Error(c, USER_NOT_INIT)
		c.Abort()
	}
}

func GetLoginUserInfo(c *gin.Context) {
	memConfMux.RLock()
	user := *memConfUsers[getOpUserKey(c)]
	user.CreatorName = memConfUsers[user.CreatorKey].Name
	memConfMux.RUnlock()

	user.PassCode = ""

	Success(c, user)
}

func encryptUserPassCode(code string) string {
	hash := hmac.New(sha256.New, []byte(conf.UserPassCodeEncryptKey))
	hash.Write([]byte(code))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func setUserKeyCookie(c *gin.Context, userkey, passCode string) {
	jwtIns := jwt.New(jwt.SigningMethodHS256)
	jwtIns.Claims["uky"] = userkey
	jwtIns.Claims["upc"] = passCode

	encStr, _ := jwtIns.SignedString([]byte(conf.NodeAuth))
	cookie := new(http.Cookie)
	cookie.Name = "op_user"
	cookie.Expires = time.Now().Add(time.Duration(30*86400) * time.Second)
	cookie.Value = encStr
	cookie.Path = "/op"
	http.SetCookie(c.Writer, cookie)
}

func deleteUserKeyCookie(c *gin.Context) {
	cookie := new(http.Cookie)
	cookie.Name = "op_user"
	cookie.Value = ""
	http.SetCookie(c.Writer, cookie)
}
