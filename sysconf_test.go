package main

import (
	"testing"

	"github.com/Instafig/Instafig/models"
	"github.com/stretchr/testify/assert"
)

func TestSysConf(t *testing.T) {
	err := _clearModelData()
	assert.True(t, err == nil, "must correctly clear data")
	loadAllData()
	initNodeData()

	_, app, _, err := initOneConfig("rahuahua", "iconfreecn", models.APP_TYPE_REAL, "config1", "1", models.CONF_V_TYPE_INT)
	assert.True(t, err == nil, "must correctly add new config")

	newData := &newConfigData{
		K:      "int_conf",
		V:      `{"cond-values":[{"condition":{"arguments":[{"symbol":"LANG"},"en"],"func":"str="},"value":999}],"default-value":100000}`,
		VType:  models.CONF_V_TYPE_CODE,
		AppKey: app.Key,
	}

	config, err := newConfigWithNewConfigData(newData, app.UserKey)
	assert.True(t, err == nil)
	assert.True(t, config.AppKey == app.Key)

	confs := getAppMatchConf(app.Key, &ClientData{Lang: "en"})
	assert.True(t, confs["int_conf"] == 999)
	confs = getAppMatchConf(app.Key, &ClientData{Lang: "en_US"})
	assert.True(t, confs["int_conf"] == 100000)

	newData = &newConfigData{
		K:      "en_US",
		V:      `en`,
		VType:  models.CONF_V_TYPE_STRING,
		AppKey: SYS_CONF_LANG,
	}
	config, err = newConfigWithNewConfigData(newData, app.UserKey)
	assert.True(t, err == nil)
	assert.True(t, config.AppKey == SYS_CONF_LANG)
	confs = getAppMatchConf(app.Key, uniformClientParams(&ClientData{Lang: "en"}))
	assert.True(t, confs["int_conf"] == 999)
	confs = getAppMatchConf(app.Key, uniformClientParams(&ClientData{Lang: "en_US"}))
	assert.True(t, confs["int_conf"] == 999)
	confs = getAppMatchConf(app.Key, uniformClientParams(&ClientData{Lang: "zh"}))
	assert.True(t, confs["int_conf"] == 100000)

	updateData := &updateConfigData{
		Key:    config.Key,
		K:      "en_US",
		V:      `en_EN`,
		VType:  models.CONF_V_TYPE_STRING,
		Status: models.CONF_STATUS_ACTIVE,
	}

	config, err = updateConfigWithUpdateData(updateData, app.UserKey)
	assert.True(t, err == nil)
	assert.True(t, config.AppKey == SYS_CONF_LANG)
	confs = getAppMatchConf(app.Key, uniformClientParams(&ClientData{Lang: "en"}))
	assert.True(t, confs["int_conf"] == 999)
	confs = getAppMatchConf(app.Key, uniformClientParams(&ClientData{Lang: "en_US"}))
	assert.True(t, confs["int_conf"] == 100000)
	confs = getAppMatchConf(app.Key, uniformClientParams(&ClientData{Lang: "zh"}))
	assert.True(t, confs["int_conf"] == 100000)
}
