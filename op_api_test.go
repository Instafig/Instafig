package main

import (
	"testing"

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
		VType:  models.CONF_V_TYPE_INT}, nil)
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, memConfAppConfigs[app.Key][0].Key == config.Key, "must the same config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 1, "must one config for app")

	_, err = updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "float_conf",
		V:      "1.2",
		VType:  models.CONF_V_TYPE_FLOAT}, nil)
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 2, "must two configs for app")
}

func TestDataVersion(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initLocalNodeData()

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

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
		VType:  models.CONF_V_TYPE_INT}, nil)
	assert.True(t, memConfDataVersion.Version == 3, "data version must be 3")
	assert.True(t, memConfDataVersion.OldSign == oldVersion.Sign)
	assert.True(t, memConfDataVersion.Sign != oldVersion.Sign)
}
