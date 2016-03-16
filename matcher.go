package main

import (
	"strconv"

	"github.com/appwilldev/Instafig/models"
)

type ClientData struct {
	AppKey     string `json:"app_key" binding:"required"`
	OSType     string `json:"os_type" binding:"required"`
	OSVersion  string `json:"os_version" binding:"required"`
	AppVersion string `json:"app_version" binding:"required"`
	Ip         string `json:"ip" binding:"required"`
	Lang       string `json:"lang" binding:"required"`
	DeviceId   string `json:"device_id"`
	DataSign   string `json:"data_sign"`
	TimeZone   string `json:"time_zone"`
	NetWork    string `json:"net_work"`
}

type Config struct {
	Key    string
	AppKey string
	K      string
	V      interface{}
	VType  string
	Status int
}

func transConfig(m *models.Config) *Config {
	config := &Config{
		Key:    m.Key,
		AppKey: m.AppKey,
		K:      m.K,
		VType:  m.VType,
		Status: m.Status,
	}

	switch m.VType {
	case models.CONF_V_TYPE_FLOAT:
		config.V, _ = strconv.ParseFloat(m.V, 64)
	case models.CONF_V_TYPE_INT:
		config.V, _ = strconv.Atoi(m.V)
	case models.CONF_V_TYPE_STRING:
		config.V = m.V
	case models.CONF_V_TYPE_CODE:
		sexp, _ := JsonToSexpString(m.V)
		config.V = NewDynValFromSexpStringDefault(sexp)
	case models.CONF_V_TYPE_TEMPLATE:
		config.V = m.V
	}

	return config
}

func getMatchConf(matchData *ClientData, configs []*Config) map[string]interface{} {
	res := make(map[string]interface{}, 0)
	for _, config := range configs {
		if config.Status != models.CONF_STATUS_ACTIVE {
			continue
		}
		switch config.VType {
		case models.CONF_V_TYPE_CODE:
			res[config.K], _ = EvalDynVal(config.V.(*DynVal), matchData)
		case models.CONF_V_TYPE_TEMPLATE:
			res[config.K] = getAppMatchConf(config.V.(string), matchData)
		default:
			res[config.K] = config.V
		}
	}

	return res
}

func getAppMatchConf(appKey string, clientData *ClientData) map[string]interface{} {
	appConfigs := getAppMemConfig(appKey)
	if appConfigs == nil {
		return map[string]interface{}{}
	}

	return getMatchConf(clientData, appConfigs)
}
