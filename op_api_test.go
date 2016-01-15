package main

import (
	"testing"

	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
	"github.com/stretchr/testify/assert"
)

// !!!!NOTE!!!! only used in UT
func _clearModelData() error {
	sql := "delete from config; delete from app; delete from user;delete from node;update data_version set ver=0;"
	s := models.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		s.Rollback()
		return err
	}
	if _, err := s.Exec(sql); err != nil {
		s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		s.Rollback()
		return err
	}

	return nil
}

func TestNewUser(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey(),
	})
	assert.True(t, err == nil, "must correctly add new user")
	assert.True(t, len(memConfUsers) == 1, "must only one user")
	assert.True(t, user.Key == memConfUsersByName["rahuahua"].Key, "must the same user")

	_, err = updateUser(&models.User{
		Name: "rahuahua2",
		Key:  utils.GenerateKey(),
	})
	assert.True(t, err == nil, "must correctly add new user")
	assert.True(t, len(memConfUsers) == 2, "must two users")
}

func TestNewApp(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey(),
	})

	app, err := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL,
	})
	assert.True(t, err == nil, "must correctly add new app")
	assert.True(t, len(memConfApps) == 1, "must only one app")
	assert.True(t, app.Key == memConfAppsByName["iconfreecn"][0].Key, "must the same app")

	_, err = updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "hdfreecn",
		Type:    models.APP_TYPE_REAL,
	})
	assert.True(t, err == nil, "must correctly add new app")
	assert.True(t, len(memConfApps) == 2, "must two apps")
}

func TestNewConfig(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey(),
	})

	app, err := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL,
	})

	config, err := updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "int_conf",
		V:      "1",
		VType:  models.CONF_V_TYPE_INT,
	})
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, memConfAppConfigs[app.Key][0].Key == config.Key, "must the same config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 1, "must one config for app")

	_, err = updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "float_conf",
		V:      "1.2",
		VType:  models.CONF_V_TYPE_FLOAT,
	})
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 2, "must two configs for app")
}

func TestDataVersion(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()

	assert.True(t, memConfDataVersion == 0, "init data version must be 0")

	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey(),
	})
	assert.True(t, memConfDataVersion == 1, "data version must be 1")

	app, _ := updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL,
	})
	assert.True(t, memConfDataVersion == 2, "data version must be 2")

	updateConfig(&models.Config{
		Key:    utils.GenerateKey(),
		AppKey: app.Key,
		K:      "int_conf",
		V:      "1",
		VType:  models.CONF_V_TYPE_INT,
	})
	assert.True(t, memConfDataVersion == 3, "data version must be 3")
}
