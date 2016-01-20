package main

import (
	"github.com/appwilldev/Instafig/conf"
	"github.com/gin-gonic/gin"
)

func ClientReqData(c *gin.Context) {
	clientData := &ClientData{
		AppKey:     c.Query("app_key"),
		OSType:     c.Query("os_type"),
		OSVersion:  c.Query("os_version"),
		AppVersion: c.Query("app_version"),
		Ip:         c.Query("ip"),
		Lang:       c.Query("lang"),
		DeviceId:   c.Query("device_id"),
		DataSign:   c.Query("data_sign"),
	}

	if conf.IsEasyDeployMode() {
		if !conf.IsMasterNode() {
			//todo: to check node sync status
		}

		memConfMux.RLock()
		nodes := memConfNodes
		needConf := memConfApps[clientData.AppKey] != nil && clientData.DataSign != memConfApps[clientData.AppKey].DataSign
		memConfMux.RUnlock()

		var confData map[string]interface{}
		newDataSign := ""
		if needConf {
			confData = getAppMatchConf(clientData.AppKey, clientData)
			memConfMux.RLock()
			newDataSign = memConfApps[clientData.AppKey].DataSign
			memConfMux.RUnlock()
		}

		nodeRes := make([]string, len(nodes))
		ix := 0
		for _, node := range nodes {
			nodeRes[ix] = node.URL
			ix++
		}

		if needConf {
			Success(c, map[string]interface{}{
				"nodes":     nodeRes,
				"configs":   confData,
				"data_sign": newDataSign,
			})
		} else {
			Success(c, map[string]interface{}{
				"nodes": nodeRes,
			})
		}
		return
	}
}
