package main

import (
	"net/http"

	"github.com/Instafig/Instafig/conf"
	"github.com/Instafig/Instafig/utils"
	"github.com/gin-gonic/gin"
)

func ClientConf(c *gin.Context) {
	clientData := &ClientData{
		AppKey:     c.Query("app_key"),
		OSType:     c.Query("os_type"),
		OSVersion:  c.Query("os_version"),
		AppVersion: c.Query("app_version"),
		Ip:         c.Query("ip"),
		Lang:       c.Query("lang"),
		DeviceId:   c.Query("device_id"),
		DataSign:   c.Query("data_sign"),
		TimeZone:   c.Query("timezone"),
		NetWork:    c.Query("network"),
	}

	if clientData.AppKey == "" {
		memConfMux.RLock()
		app := memConfAppsByName[c.Query("app")]
		memConfMux.RUnlock()

		if app != nil {
			clientData.AppKey = app.Key
			clientData.OSType = "ios"
			clientData.OSVersion = c.Query("osv")
			clientData.AppVersion = c.Query("v")
			clientData.DeviceId = c.Query("ida")
			clientData.Ip = c.Request.RemoteAddr
		}

		setClientData(c, clientData)

		c.JSON(http.StatusOK, getAppMatchConf(clientData.AppKey, clientData))
		return
	}

	setClientData(c, clientData)
	memConfMux.RLock()
	if !conf.IsMasterNode() && conf.DataExpires > 0 {
		if memConfNodes[conf.ClientAddr].LastCheckUTC < utils.GetNowSecond()-conf.DataExpires {
			memConfMux.RUnlock()
			Error(c, DATA_EXPIRED)
			return
		}
	}

	nodes := []string{}
	if !conf.ProxyDeployed {
		nodes = make([]string, len(memConfNodes))
		ix := 0
		for _, node := range memConfNodes {
			nodes[ix] = node.URL
			ix++
		}
	}

	// do not support app data_sign in server-side, always return app configs
	needConf := true || (memConfApps[clientData.AppKey] != nil && clientData.DataSign != memConfApps[clientData.AppKey].DataSign)
	memConfMux.RUnlock()

	if needConf {
		var dataSign string
		configs := getAppMatchConf(clientData.AppKey, clientData)
		if len(configs) > 0 {
			memConfMux.RLock()
			dataSign = memConfApps[clientData.AppKey].DataSign
			memConfMux.RUnlock()
		}

		Success(c, map[string]interface{}{
			"nodes":     nodes,
			"configs":   configs,
			"data_sign": dataSign,
		})
	} else {
		Success(c, map[string]interface{}{
			"nodes": nodes,
		})
	}
	return
}
