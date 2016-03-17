package main

import (
	"testing"

	"reflect"

	"github.com/Instafig/Instafig/models"
	"github.com/Instafig/Instafig/utils"
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

	userData := &newUserData{
		Name:     "rahuahua",
		PassCode: "huahua",
	}

	assert.True(t, verifyNewUserData(userData) == nil)

	user, err := newUserWithNewUserData(userData, "1234567", "1234567")
	assert.True(t, err == nil, "must correctly add new user")
	assert.True(t, len(memConfUsers) == 1, "must only one user")
	assert.True(t, user.Key == memConfUsersByName["rahuahua"].Key, "must the same user")

	badUserData := &newUserData{
		Name:     "rahuahua",
		PassCode: "huahua",
	}
	assert.True(t, verifyNewUserData(badUserData) != nil)

	badUserData.Name = "12"
	assert.True(t, verifyNewUserData(badUserData) != nil)

	badUserData.Name = "non-exists"
	badUserData.PassCode = "12"
	assert.True(t, verifyNewUserData(badUserData) != nil)

	userData = &newUserData{
		Name:     "rahuahua2",
		PassCode: "huahua22",
	}
	user, err = newUserWithNewUserData(userData, "12345678", "1234567")
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

	userData := &newUserData{
		Name:     "rahuahua",
		PassCode: "huahua",
	}
	user, err := newUserWithNewUserData(userData, "1234567", "1234567")
	assert.True(t, err == nil, "must correctly add new user")

	userData = &newUserData{
		Name:     "rahuahua2",
		PassCode: "huahua",
	}
	newUser, err := newUserWithNewUserData(userData, utils.GenerateKey(), user.Key)
	assert.True(t, err == nil, "must correctly add new user")
	assert.True(t, newUser.CreatorKey == user.Key && newUser.CreatedUTC >= user.CreatedUTC)

	updateData := &updateUserData{
		Name:    "rahuahua333",
		AuxInfo: "1234",
	}

	assert.True(t, verifyUpdateUserData(updateData, user.Key) == nil)

	oldUser := *user
	user, err = updateUserWithUpdateData(updateData, user.Key)
	assert.True(t, err == nil)
	assert.True(t, user.Key == oldUser.Key)
	assert.True(t, user.Name == "rahuahua333")

	badData := &updateUserData{
		Name: "rahuahua2",
	}
	assert.True(t, verifyUpdateUserData(badData, user.Key) != nil)
	badData.Name = "12"
	assert.True(t, verifyUpdateUserData(badData, user.Key) != nil)

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

func TestSearchApps(t *testing.T) {
	user, err := updateUser(&models.User{
		Name: "rahuahua",
		Key:  utils.GenerateKey()}, nil)

	_, err = updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "iconfreecn",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, err == nil, "must correctly add new app")

	_, err = updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "xianyouvideo",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, err == nil, "must correctly add new app")

	_, err = updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "hdfreecn",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, err == nil, "must correctly add new app")

	_, err = updateApp(&models.App{
		Key:     utils.GenerateKey(),
		UserKey: user.Key,
		Name:    "phoneplay",
		Type:    models.APP_TYPE_REAL}, nil)
	assert.True(t, err == nil, "must correctly add new app")

	apps, err := searchApps("free", 0)
	assert.True(t, err == nil)
	assert.True(t, len(apps) == 2)

	apps, err = searchApps("video", 0)
	assert.True(t, err == nil)
	assert.True(t, len(apps) == 1)

	apps, err = searchApps("phoneplay", 0)
	assert.True(t, err == nil)
	assert.True(t, len(apps) == 1)

	apps, err = searchApps("non-exist", 0)
	assert.True(t, err == nil)
	assert.True(t, len(apps) == 0)
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

	user, app, config, err := initOneConfig("rahuahua", "iconfreecn", models.APP_TYPE_REAL, "config1", "1", models.CONF_V_TYPE_INT)
	assert.True(t, err == nil, "must correctly add new config")
	assert.True(t, memConfAppConfigs[app.Key][0].Key == config.Key, "must the same config")
	assert.True(t, len(memConfAppConfigs[app.Key]) == 1, "must one config for app")

	oldAppDataSign := memConfApps[app.Key].DataSign
	newData := &newConfigData{
		K:      "float_conf",
		V:      "1.2",
		VType:  models.CONF_V_TYPE_FLOAT,
		AppKey: app.Key,
	}

	err = verifyNewConfigData(newData)
	assert.True(t, err == nil)

	config, err = newConfigWithNewConfigData(newData, user.Key)
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
	assert.True(t, config.UpdateTimes == oldConfig.UpdateTimes+1)
	assert.True(t, config.LastUpdateId != oldConfig.LastUpdateId)

	_clearModelData()
}

func TestVerifyNewConfigData(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	_, app, _, err := initOneConfig("rahuahua", "iconfreecn", models.APP_TYPE_REAL, "config1", "1", models.CONF_V_TYPE_INT)
	assert.True(t, err == nil, "must correctly add new config")

	newData := &newConfigData{
		K:      "float_conf",
		V:      "1.2",
		VType:  models.CONF_V_TYPE_FLOAT,
		AppKey: app.Key,
	}
	err = verifyNewConfigData(newData)
	assert.True(t, err == nil)

	badData := *newData
	badData.AppKey = "non-exist-app-key"
	err = verifyNewConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.K = "config1"
	err = verifyNewConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "12"
	badData.VType = models.CONF_V_TYPE_INT
	err = verifyNewConfigData(&badData)
	assert.True(t, err == nil)

	badData = *newData
	badData.V = "1.2"
	badData.VType = models.CONF_V_TYPE_INT
	err = verifyNewConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "1.223324"
	badData.VType = models.CONF_V_TYPE_FLOAT
	err = verifyNewConfigData(&badData)
	assert.True(t, err == nil)

	badData = *newData
	badData.V = "1.2a"
	badData.VType = models.CONF_V_TYPE_FLOAT
	err = verifyNewConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "non-exist-template-app-key"
	badData.VType = models.CONF_V_TYPE_TEMPLATE
	err = verifyNewConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "INVALID_CODE_VALUE"
	badData.VType = models.CONF_V_TYPE_CODE
	err = verifyNewConfigData(&badData)
	assert.True(t, err != nil)

	_clearModelData()
}

func TestVerifyUpdateConfigData(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	user, app, config, err := initOneConfig("rahuahua", "iconfreecn", models.APP_TYPE_REAL, "config1", "1", models.CONF_V_TYPE_INT)
	assert.True(t, err == nil, "must correctly add new config")

	newData := &updateConfigData{
		K:     "float_conf",
		V:     "1.2",
		VType: models.CONF_V_TYPE_FLOAT,
		Key:   config.Key,
	}
	err = verifyUpdateConfigData(newData)
	assert.True(t, err == nil)

	badData := *newData
	badData.Key = "non-exist-config-key"
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err != nil)

	_, err = newConfigWithNewConfigData(
		&newConfigData{
			K:      "already-exist-config-key",
			V:      "1.2",
			VType:  models.CONF_V_TYPE_FLOAT,
			AppKey: app.Key}, user.Key)
	assert.True(t, err == nil)
	badData = *newData
	badData.K = "already-exist-config-key"
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "12"
	badData.VType = models.CONF_V_TYPE_INT
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err == nil)

	badData = *newData
	badData.V = "1.2"
	badData.VType = models.CONF_V_TYPE_INT
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "1.223324"
	badData.VType = models.CONF_V_TYPE_FLOAT
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err == nil)

	badData = *newData
	badData.V = "1.2a"
	badData.VType = models.CONF_V_TYPE_FLOAT
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "non-exist-template-app-key"
	badData.VType = models.CONF_V_TYPE_TEMPLATE
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err != nil)

	badData = *newData
	badData.V = "INVALID_CODE_VALUE"
	badData.VType = models.CONF_V_TYPE_CODE
	err = verifyUpdateConfigData(&badData)
	assert.True(t, err != nil)

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
