package main

import (
	"crypto/sha1"
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
		Error(c, PASS_CODE_ERR)
		return
	}

	setUserKeyCookie(c, user.Key)
	Success(c, nil)
}

type newUserData struct {
	Name     string `json:"name" binding:"required"`
	PassCode string `json:"pass_code" binding:"required"`
}

func InitUser(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	data := &newUserData{}
	if err := c.BindJSON(data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	memConfMux.RLock()
	if len(memConfUsersByName) > 0 {
		Error(c, NOT_PERMITTED, "some users already exists: ")
		memConfMux.RUnlock()
		return
	}
	memConfMux.RUnlock()

	user := &models.User{
		Name:     data.Name,
		PassCode: encryptUserPassCode(data.PassCode),
		Creator:  "",
		Key:      utils.GenerateKey()}

	if _, err := updateUser(user, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(user)
	setUserKeyCookie(c, user.Key)
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

	memConfMux.RLock()
	if memConfUsersByName[data.Name] != nil {
		Error(c, BAD_REQUEST, "user name already exists: "+data.Name)
		memConfMux.RUnlock()
		return
	}
	memConfMux.RUnlock()

	user := &models.User{
		Name:     data.Name,
		PassCode: encryptUserPassCode(data.PassCode),
		Creator:  getOpUserKey(c),
		Key:      utils.GenerateKey()}

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

	res := make([]map[string]interface{}, len(users))
	for ix, user := range users {
		user.PassCode = ""
		authorName := ""
		memConfMux.RLock()
		if memConfUsers[user.Creator] != nil {
			authorName = memConfUsers[user.Creator].Name
		}
		memConfMux.RUnlock()
		res[ix] = map[string]interface{}{
			"user": user,
			"creator": map[string]interface{}{"name": authorName, "key":user.Creator},
		}
	}

	Success(c, res)
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
	if conf.IsEasyDeployMode() && memConfAppsByName[data.Name] != nil {
		Error(c, BAD_REQUEST, "appname already exists: "+data.Name)
		return
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

func GetAllApps(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		Error(c, BAD_REQUEST, "page not number")
		return
	}

	apps, err := models.GetAllApps(nil, page, 25)
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

	config, err := updateConfig(config, getOpUserKey(c), nil)
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
	if oldConfig.K != data.K {
		Error(c, BAD_REQUEST, "can not change config's key")
		return
	}

	config := &models.Config{
		Key:    data.Key,
		AppKey: data.AppKey,
		K:      data.K,
		V:      data.V,
		VType:  data.VType,
	}

	config, err := updateConfig(config, getOpUserKey(c), nil)
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

func updateConfig(config *models.Config, userKey string, newDataVersion *models.DataVersion) (*models.Config, error) {
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
	app := memConfApps[config.AppKey]
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

		configHistory := &models.ConfigUpdateHistory{
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
		if err := models.InsertDBModel(s, configHistory); err != nil {
			s.Rollback()
			return nil, err
		}
	} else {
		if err := models.UpdateDBModel(s, config); err != nil {
			s.Rollback()
			return nil, err
		}

		configHistory := &models.ConfigUpdateHistory{
			Id:         utils.GenerateKey(),
			ConfigKey:  config.Key,
			K:          config.K,
			OldV:       oldConfig.V,
			OldVType:   oldConfig.VType,
			NewV:       config.V,
			NewVType:   config.VType,
			Kind:       models.CONFIG_UPDATE_KIND_UPDATE,
			UserKey:    userKey,
			CreatedUTC: utils.GetNowSecond(),
		}
		if err := models.InsertDBModel(s, configHistory); err != nil {
			s.Rollback()
			return nil, err
		}
	}

	toUpdateApps := make([]*models.App, 0)
	if app.Type == models.APP_TYPE_REAL {
		toUpdateApps = append(toUpdateApps, app)
	} else {
		memConfMux.RLock()
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
		memConfMux.RUnlock()
	}

	newDataSign := utils.GenerateKey()
	for _, app := range toUpdateApps {
		memConfMux.Lock()
		app.DataSign = newDataSign
		memConfMux.Unlock()
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
	memConfRawConfigs[config.Key] = config
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

func GetConfigUpdateHistory(c *gin.Context) {
	histories, err := models.GetConfigUpdateHistory(nil, c.Param("config_key"))
	if err != nil {
		Error(c, SERVER_ERROR)
		return
	}

	res := make([]map[string]interface{}, len(histories))
	for ix, history := range histories {
		memConfMux.RLock()
		userName := memConfUsers[history.UserKey].Name
		memConfMux.RUnlock()
		res[ix] = map[string]interface{}{
			"value":  history,
			"author": map[string]string{"name": userName, "key": history.UserKey},
		}
	}

	Success(c, res)
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
		return []byte(conf.MasterAuth), nil
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
	if memConfUsers[userKey] == nil {
		memConfMux.RUnlock()
		Error(c, NOT_LOGIN, "user not exist")
		c.Abort()
		return
	}
	memConfMux.RUnlock()

	setOpUserKey(c, userKey)
}

func encryptUserPassCode(code string) string {
	s := sha1.Sum([]byte(code))
	return string(s[:sha1.Size])
}

func setUserKeyCookie(c *gin.Context, userKey string) {
	jwtIns := jwt.New(jwt.SigningMethodHS256)
	jwtIns.Claims["uky"] = userKey

	encStr, _ := jwtIns.SignedString([]byte(conf.MasterAuth))
	cookie := new(http.Cookie)
	cookie.Name = "op_user"
	cookie.Expires = time.Now().Add(time.Duration(30*86400) * time.Second)
	cookie.Value = encStr
	//	cookie.Path = "/"
	//	cookie.Domain = ""
	http.SetCookie(c.Writer, cookie)
}
