package main

import (
	"testing"

	"github.com/appwilldev/Instafig/models"
	"github.com/stretchr/testify/assert"
)

func clearModelData() error {
	return models.ClearData()
}

func TestNewUser(t *testing.T) {
	err := clearModelData()
	assert.True(t, err == nil, "must correct clear data")

	loadAllData()

	user, err := newUser(&newUserData{
		Name: "rahuahua",
	})
	assert.True(t, err == nil, "must correct add new user")
	assert.True(t, len(memConfUsers) == 1, "must only one user")
	assert.True(t, user.Key == memConfUsersByName["rahuahua"].Key, "must the same user")

	_, err = newUser(&newUserData{
		Name: "rahuahua2",
	})
	assert.True(t, err == nil, "must correct add new user")
	assert.True(t, len(memConfUsers) == 2, "must two users")
}

func TestNewApp(t *testing.T) {
	err := clearModelData()
	assert.True(t, err == nil, "must correct clear data")

	loadAllData()

	user, err := newUser(&newUserData{
		Name: "rahuahua",
	})

	app, err := newApp(&newAppData{
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL,
	})
	assert.True(t, err == nil, "must correct add new app")
	assert.True(t, len(memConfApps) == 1, "must only one app")
	assert.True(t, app.Key == memConfAppsByName["iconfreecn"][0].Key, "must the same app")

	_, err = newApp(&newAppData{
		UserKey: user.Key,
		Name:    "hdfreecn",
		Type:    models.APP_TYPE_REAL,
	})
	assert.True(t, err == nil, "must correct add new app")
	assert.True(t, len(memConfApps) == 2, "must two apps")
}

func TestNewConfig(t *testing.T) {
	err := clearModelData()
	assert.True(t, err == nil, "must correct clear data")

	loadAllData()

	user, err := newUser(&newUserData{
		Name: "rahuahua",
	})

	app, err := newApp(&newAppData{
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL,
	})

	config, err := newConfig(&newConfigData{
		AppKey: app.Key,
		K:      "int_conf",
		V:      "1",
		VType:  models.CONF_V_TYPE_INT,
	})
	assert.True(t, err == nil, "must correct add new config")
	assert.True(t, memConfAppConfigs[app.Key][0].Key == config.Key, "must the same config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 1, "must one config for app")

	_, err = newConfig(&newConfigData{
		AppKey: app.Key,
		K:      "float_conf",
		V:      "1.2",
		VType:  models.CONF_V_TYPE_FLOAT,
	})
	assert.True(t, err == nil, "must correct add new config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 2, "must two configs for app")
}
