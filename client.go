package main

import (
	"net/http"

	"sort"

	"github.com/Instafig/Instafig/conf"
	"github.com/Instafig/Instafig/models"
	"github.com/Instafig/Instafig/utils"
	"github.com/gin-gonic/gin"
)

var (
	clientQueryParamCh = make(chan interface{}, 128)
)

func init() {
	recordClientQueryParam()
}

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

		clientData = uniformClientParams(clientData)
		sendChanAsync(clientQueryParamCh, clientData)
		setClientData(c, clientData)

		c.JSON(http.StatusOK, getAppMatchConf(clientData.AppKey, clientData))
		return
	}

	clientData = uniformClientParams(clientData)
	sendChanAsync(clientQueryParamCh, clientData)
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
	nodes = make([]string, len(memConfNodes))
	ix := 0
	for _, node := range memConfNodes {
		nodes[ix] = node.URL
		ix++
	}

	needConf := memConfApps[clientData.AppKey] != nil && clientData.DataSign != memConfApps[clientData.AppKey].DataSign
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
			"data_sign": clientData.DataSign,
			"nodes":     nodes,
		})
	}

	return
}

func recordClientQueryParam() {
	doEverTask(func() {
		for {
			i := <-clientQueryParamCh
			cdata := i.(*ClientData)

			if cdata.Lang != "" && !memConfClientLang[cdata.Lang] {
				err := models.InsertRow(
					nil,
					&models.ClientReqeustData{AppKey: "", Symbol: GLISP_SYMBOL_TYPE_LANG, Value: cdata.Lang})
				if err == nil {
					memConfClientMux.Lock()
					memConfClientLang[cdata.Lang] = true
					memConfClientMux.Unlock()
				}
			}

			if cdata.OSType != "" && !memConfClientOSType[cdata.OSType] {
				err := models.InsertRow(
					nil,
					&models.ClientReqeustData{AppKey: "", Symbol: GLISP_SYMBOL_TYPE_OS_TYPE, Value: cdata.OSType})
				if err == nil {
					memConfClientMux.Lock()
					memConfClientOSType[cdata.OSType] = true
					memConfClientMux.Unlock()
				}
			}

			if cdata.OSVersion != "" && !memConfClientOSV[cdata.OSVersion] {
				err := models.InsertRow(
					nil,
					&models.ClientReqeustData{AppKey: "", Symbol: GLISP_SYMBOL_TYPE_OS_VERSION, Value: cdata.OSVersion})
				if err == nil {
					memConfClientMux.Lock()
					memConfClientOSV[cdata.OSVersion] = true
					memConfClientMux.Unlock()
				}
			}

			if cdata.TimeZone != "" && !memConfClientTimezone[cdata.TimeZone] {
				err := models.InsertRow(
					nil,
					&models.ClientReqeustData{AppKey: "", Symbol: GLISP_SYMBOL_TYPE_TIMEZONE, Value: cdata.TimeZone})
				if err == nil {
					memConfClientMux.Lock()
					memConfClientTimezone[cdata.TimeZone] = true
					memConfClientMux.Unlock()
				}
			}

			if cdata.NetWork != "" && !memConfClientNetwork[cdata.NetWork] {
				err := models.InsertRow(
					nil,
					&models.ClientReqeustData{AppKey: "", Symbol: GLISP_SYMBOL_TYPE_NETWORK, Value: cdata.NetWork})
				if err == nil {
					memConfClientMux.Lock()
					memConfClientNetwork[cdata.NetWork] = true
					memConfClientMux.Unlock()
				}
			}

			if _, ok := memConfClientAppVersion[cdata.AppKey]; !ok {
				memConfClientMux.Lock()
				memConfClientAppVersion[cdata.AppKey] = map[string]bool{}
				memConfClientMux.Unlock()
			}
			if cdata.AppVersion != "" && !memConfClientAppVersion[cdata.AppKey][cdata.AppVersion] {
				err := models.InsertRow(
					nil,
					&models.ClientReqeustData{AppKey: cdata.AppKey, Symbol: GLISP_SYMBOL_TYPE_APP_VERSION, Value: cdata.AppVersion})
				if err == nil {
					memConfClientMux.Lock()
					memConfClientAppVersion[cdata.AppKey][cdata.AppVersion] = true
					memConfClientMux.Unlock()
				}
			}
		}
	})
}

func GetClientSymbols(c *gin.Context) {
	memConfClientMux.RLock()
	defer memConfClientMux.RUnlock()

	res := []string{}
	var t map[string]bool
	switch c.Param("symbol") {
	case GLISP_SYMBOL_TYPE_LANG:
		t = memConfClientLang
	case GLISP_SYMBOL_TYPE_OS_TYPE:
		t = memConfClientOSType
	case GLISP_SYMBOL_TYPE_OS_VERSION:
		t = memConfClientOSV
	case GLISP_SYMBOL_TYPE_TIMEZONE:
		t = memConfClientTimezone
	case GLISP_SYMBOL_TYPE_NETWORK:
		t = memConfClientNetwork
	case GLISP_SYMBOL_TYPE_APP_VERSION:
		t = memConfClientAppVersion[c.Query("app_key")]
	default:
		Error(c, BAD_REQUEST, "unknown symbol type: "+c.Query("symbol"))
		return
	}

	for v, _ := range t {
		res = append(res, v)
	}

	sort.Strings(res)

	Success(c, res)
}
