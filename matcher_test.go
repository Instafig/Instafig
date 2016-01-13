package main

import (
	"testing"

	"github.com/appwilldev/Instafig/models"
	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	clientData := &ClientData{
		AppKey:     "app1",
		OSType:     "ios",
		OSVersion:  "9.3",
		AppVersion: "1.1",
		Ip:         "14.32.123.23",
		Lang:       "zh",
	}

	users := []*models.User{&models.User{Key: "user1", Name: "user1"}}
	apps := []*models.App{&models.App{Key: "app1", Name: "app1", UserKey: "user1", Type: models.APP_TYPE_REAL}}
	configs := []*models.Config{
		&models.Config{
			Key:    "conf1",
			AppKey: "app1",
			K:      "time_out",
			V:      "1",
			VType:  models.CONF_V_TYPE_INT,
		},
		&models.Config{
			Key:    "conf2",
			AppKey: "app1",
			K:      "accuracy",
			V:      "1.2",
			VType:  models.CONF_V_TYPE_FLOAT,
		},
		&models.Config{
			Key:    "conf3",
			AppKey: "app1",
			K:      "dsn",
			V:      "beijing.appdao.com:8080",
			VType:  models.CONF_V_TYPE_STRING,
		},
		&models.Config{
			Key:    "conf4",
			AppKey: "app1",
			K:      "guaji",
			V: `(if (= lang "cn")
			(if (and (>= app_version "1.3.1") (< app_version "1.5")) 1 0)
			(if (and (>= app_version "1.3.1") (< app_version "1.5")) 2 3))`,
			VType: models.CONF_V_TYPE_CODE,
		},
	}

	fillMemConfData(users, apps, configs, nil)

	res := getAppMatchConf("app1", clientData)
	assert.True(t, res["time_out"].(int) == 1)
	assert.True(t, res["accuracy"].(float64) == 1.2)
	assert.True(t, res["dsn"].(string) == "beijing.appdao.com:8080")
	assert.True(t, res["guaji"] == nil)
	assert.True(t, res["no-exist-key"] == nil)
}
