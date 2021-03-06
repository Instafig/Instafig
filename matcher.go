package main

import (
	"strconv"

	"github.com/Instafig/Instafig/models"
)

type ClientData struct {
	AppKey     string `json:"app_key"`
	OSType     string `json:"os_type"`
	OSVersion  string `json:"os_version"`
	AppVersion string `json:"app_version"`
	Ip         string `json:"ip"`
	Lang       string `json:"lang"`
	DeviceId   string `json:"device_id"`
	DataSign   string `json:"data_sign"`
	TimeZone   string `json:"timezone"`
	NetWork    string `json:"network"`
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
		sexp, err := JsonToSexpString(m.V)
		if err != nil {
			logger.Error(map[string]interface{}{
				"type":  "bad_code_config",
				"json":  m.V,
				"error": err.Error(),
			})
			return config
		}
		logger.Debug(map[string]interface{}{
			"type": "code_config",
			"json": m.V,
			"sexp": sexp,
		})
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

func getMatchConfWithKey(matchData *ClientData, configs []*Config, key string) map[string]interface{} {
	res := make(map[string]interface{}, 0)
	for _, config := range configs {
		if config.Status != models.CONF_STATUS_ACTIVE || config.K != key {
			continue
		}
		switch config.VType {
		case models.CONF_V_TYPE_CODE:
			res[config.K], _ = EvalDynVal(config.V.(*DynVal), matchData)
		case models.CONF_V_TYPE_TEMPLATE:
			res[config.K] = getAppMatchConfWithKey(config.V.(string), matchData, key)
		default:
			res[config.K] = config.V
		}

		// key is unique, so quit here
		return res
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

func getAppMatchConfWithKey(appKey string, clientData *ClientData, key string) map[string]interface{} {
	appConfigs := getAppMemConfig(appKey)
	if appConfigs == nil {
		return map[string]interface{}{}
	}

	return getMatchConfWithKey(clientData, appConfigs, key)
}
