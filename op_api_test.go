package main

import (
	"testing"

	"reflect"

	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

// !!!!NOTE!!!! only used in UT
func _clearModelData() error {
	return models.ClearModeData(nil)
}

func TestNewUser(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey()}, nil)
	assert.True(t, err == nil, "must correctly add new user")
	assert.True(t, len(memConfUsers) == 1, "must only one user")
	assert.True(t, user.Key == memConfUsersByName["rahuahua"].Key, "must the same user")

	_, err = updateUser(&models.User{
		Name: "rahuahua2",
		Key:  utils.GenerateKey()}, nil)
	assert.True(t, err == nil, "must correctly add new user")
	assert.True(t, len(memConfUsers) == 2, "must two users")

	_clearModelData()
}

func TestUpdateUser(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey()}, nil)

	user, err = updateUser(&models.User{
		Name:    "rahuahua2",
		Key:     user.Key,
		AuxInfo: "guaji"}, nil)
	assert.True(t, err == nil, "must correctly add new user")
	assert.True(t, len(memConfUsers) == 1, "must only one user")
	assert.True(t, memConfUsersByName["rahuahua"] == nil, "old-name user must not exist")
	assert.True(t, memConfUsersByName["rahuahua2"].AuxInfo == "guaji", "aux_info must be updated")

	_clearModelData()
}

func TestNewApp(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey()}, nil)

	app, err := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, err == nil, "must correctly add new app")
	assert.True(t, len(memConfApps) == 1, "must only one app")
	assert.True(t, app.Key == memConfAppsByName["iconfreecn"].Key, "must the same app")

	_, err = updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "hdfreecn",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, err == nil, "must correctly add new app")
	assert.True(t, len(memConfApps) == 2, "must two apps")

	_clearModelData()
}

func TestUpdateApp(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey()}, nil)
	app, err := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL}, nil)

	app, err = updateApp(&models.App{
		Key:     app.Key,
		UserKey: user.Key,
		Name:    "hdfreecn",
		Type:    models.APP_TYPE_REAL,
		AuxInfo: "guaji"}, nil)
	assert.True(t, err == nil, "must correctly add new app")
	assert.True(t, len(memConfApps) == 1, "must only one app")
	assert.True(t, memConfAppsByName["iconfreecn"] == nil, "old-name app must not exist")
	assert.True(t, memConfAppsByName["hdfreecn"].AuxInfo == "guaji", "aux_info must be updated")

	_clearModelData()
}

func initOneConfig(userName, appName, appType, configK, configV, configVType string) (*models.User, *models.App, *models.Config, error) {
	user, err := updateUser(&models.User{
		Name: userName,
		Key:  utils.GenerateKey()}, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	app, err := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    appName,
		Type:    appType}, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	config, err := updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      configK,
		V:      configV,
		VType:  configVType,
		Status: models.CONF_STATUS_ACTIVE}, "", nil)

	return user, app, config, err
}

func TestNewConfig(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	_, app, config, err := initOneConfig("rahuahua", "iconfreecn", models.APP_TYPE_REAL, "config1", "1", models.CONF_V_TYPE_INT)
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, memConfAppConfigs[app.Key][0].Key == config.Key, "must the same config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 1, "must one config for app")

	oldAppDataSign := memConfApps[app.Key].DataSign
	_, err = updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "float_conf",
		V:      "1.2",
		VType:  models.CONF_V_TYPE_FLOAT,
		Status: models.CONF_STATUS_ACTIVE}, "", nil)
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 2, "must two configs for app")
	assert.True(t, oldAppDataSign != memConfApps[app.Key].DataSign, "app's data_sign must update when update app config")

	_clearModelData()
}

func TestUpdateConfig(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	user, _, config, err := initOneConfig("rahuahua", "iconfreecn", models.APP_TYPE_REAL, "config1", "1", models.CONF_V_TYPE_INT)
	assert.True(t, err == nil, "must correctly add new config")

	updateData := &updateConfigData{
		Key:    config.Key,
		K:      "new_config1",
		V:      "2",
		VType:  models.CONF_V_TYPE_STRING,
		Status: models.CONF_STATUS_INACTIVE,
	}
	err = verifyUpdateConfigData(updateData)
	assert.True(t, err == nil, "all update data is valid")

	oldConfig := *config
	config, err = updateConfigWithUpdateData(updateData, user.Key)
	assert.True(t, err == nil, "all update data is valid")
	assert.True(t, config.K == "new_config1")
	assert.True(t, config.V == "2")
	assert.True(t, config.VType == models.CONF_V_TYPE_STRING)
	assert.True(t, config.Status == models.CONF_STATUS_INACTIVE)
	log.Println("=========", config.UpdateTimes)
	log.Println("=========", oldConfig.UpdateTimes)
	assert.True(t, config.UpdateTimes == oldConfig.UpdateTimes+1)
	assert.True(t, config.LastUpdateId != oldConfig.LastUpdateId)

	_clearModelData()
}

func TestDataVersion(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	assert.True(t, memConfDataVersion.Version == 0, "init data version must be 0")

	oldVersion := *memConfDataVersion
	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey()}, nil)
	assert.True(t, memConfDataVersion.Version == 1, "data version must be 1")
	assert.True(t, memConfDataVersion.OldSign == oldVersion.Sign)
	assert.True(t, memConfDataVersion.Sign != oldVersion.Sign)

	oldVersion = *memConfDataVersion
	app, _ := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, memConfDataVersion.Version == 2, "data version must be 2")
	assert.True(t, memConfDataVersion.OldSign == oldVersion.Sign)
	assert.True(t, memConfDataVersion.Sign != oldVersion.Sign)

	oldVersion = *memConfDataVersion
	updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "int_conf",
		V:      "1",
		VType:  models.CONF_V_TYPE_INT,
		Status: models.CONF_STATUS_ACTIVE}, "", nil)
	assert.True(t, memConfDataVersion.Version == 3, "data version must be 3")
	assert.True(t, memConfDataVersion.OldSign == oldVersion.Sign)
	assert.True(t, memConfDataVersion.Sign != oldVersion.Sign)

	_clearModelData()
}

func TestTemplateApp(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	user, _ := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey()}, nil)

	templateApp, _ := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "template_app",
		Type:    models.APP_TYPE_TEMPLATE}, nil)
	templateConfig, _ := updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: templateApp.Key,
		K:      "template_int_conf",
		V:      "233",
		VType:  models.CONF_V_TYPE_INT,
		Status: models.CONF_STATUS_ACTIVE}, "", nil)

	app, _ := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL}, nil)

	updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "int_conf",
		V:      "1",
		VType:  models.CONF_V_TYPE_INT,
		Status: models.CONF_STATUS_ACTIVE}, "", nil)

	_, err = updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "template_conf",
		V:      templateApp.Key,
		VType:  models.CONF_V_TYPE_TEMPLATE,
		Status: models.CONF_STATUS_ACTIVE}, "", nil)
	assert.True(t, err == nil, "must correctly add template conf")
	appConfig := getAppMatchConf(app.Key, &ClientData{AppKey: app.Key})
	assert.True(t, reflect.TypeOf(appConfig["template_conf"]).Kind() == reflect.Map)

	appOldDataSign := memConfApps[app.Key].DataSign
	oldTemplateDataSign := memConfApps[templateApp.Key].DataSign
	updateConfig(templateConfig, "", nil)
	assert.True(t, appOldDataSign != memConfApps[app.Key].DataSign, "app's data_sign must update after update config")
	assert.True(t, oldTemplateDataSign != memConfApps[templateApp.Key].DataSign, "app's data_sign must update after update config")

	_clearModelData()
}
