package main

import (
	"testing"

	"reflect"

	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
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
	initLocalNodeData()

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

func TestNewApp(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initLocalNodeData()

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
	assert.True(t, app.Key == memConfAppsByName["iconfreecn"][0].Key, "must the same app")

	_, err = updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "hdfreecn",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, err == nil, "must correctly add new app")
	assert.True(t, len(memConfApps) == 2, "must two apps")

	_clearModelData()
}

func TestNewConfig(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initLocalNodeData()

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

	config, err := updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "int_conf",
		V:      "1",
		VType:  models.CONF_V_TYPE_INT}, "", nil)
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, memConfAppConfigs[app.Key][0].Key == config.Key, "must the same config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 1, "must one config for app")

	oldAppDataSign := memConfApps[app.Key].DataSign
	_, err = updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "float_conf",
		V:      "1.2",
		VType:  models.CONF_V_TYPE_FLOAT}, "", nil)
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 2, "must two configs for app")
	assert.True(t, oldAppDataSign != memConfApps[app.Key].DataSign, "app's data_sign must update when update app config")

	_clearModelData()
}

func TestDataVersion(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initLocalNodeData()

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
		VType:  models.CONF_V_TYPE_INT}, "", nil)
	assert.True(t, memConfDataVersion.Version == 3, "data version must be 3")
	assert.True(t, memConfDataVersion.OldSign == oldVersion.Sign)
	assert.True(t, memConfDataVersion.Sign != oldVersion.Sign)

	_clearModelData()
}

func TestTemplateApp(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initLocalNodeData()

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
		VType:  models.CONF_V_TYPE_INT}, "", nil)

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
		VType:  models.CONF_V_TYPE_INT}, "", nil)

	_, err = updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "template_conf",
		V:      templateApp.Key,
		VType:  models.CONF_V_TYPE_TEMPLATE}, "", nil)
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
