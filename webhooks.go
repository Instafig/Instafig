package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
	"github.com/gin-gonic/gin"
)

func configUpdateHistoryToNotificationText(m *models.ConfigUpdateHistory, app *models.App) string {
	text := "Unkown Action"
	if m.UserName == "" {
		memConfMux.RLock()
		m.UserName = memConfUsers[m.UserKey].Name
		memConfMux.RUnlock()
	}
	switch m.Kind {
	case models.CONFIG_UPDATE_KIND_NEW:
		text = fmt.Sprintf(
			"Your config [%s] for App [%s] is just created by %s, value: %s, type: %s",
			m.K, app.Name, m.UserName, m.NewV, m.NewVType,
		)
	case models.CONFIG_UPDATE_KIND_UPDATE:
		text = fmt.Sprintf(
			`
            Your config [%s] for App [%s] is just updated by %s,
            before updated: (value = %s, type = %s),
            after updated: (value = %s, type = %s)`,
			m.K, app.Name, m.UserName, m.OldV, m.OldVType, m.NewV, m.NewVType,
		)
	case models.CONFIG_UPDATE_KIND_HIDE:
		text = fmt.Sprintf(
			"Your config [%s] for App [%s] is just hidden by %s",
			m.K, app.Name, m.UserName,
		)
	case models.CONFIG_UPDATE_KIND_RECOVER:
		text = fmt.Sprintf(
			"Your config [%s] for App [%s] is just recovered by %s",
			m.K, app.Name, m.UserName,
		)
	case models.CONFIG_UPDATE_KIND_DELETE:
		text = fmt.Sprintf(
			"Your config [%s] for App [%s] is just deleted by %s",
			m.K, app.Name, m.UserName,
		)
	}
	return text
}

func sendNotificationToPubu(targetURL string, m *models.ConfigUpdateHistory, app *models.App) error {
	text := configUpdateHistoryToNotificationText(m, app)
	data := make(map[string]interface{})
	user := make(map[string]string)
	user["name"] = "Instafig"
	user["avatarUrl"] = `https://avatars0.githubusercontent.com/u/1274781`
	data["text"] = text
	data["displayUser"] = user
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := http.Post(targetURL, "application/json", bytes.NewReader(json))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = ioutil.ReadAll(resp.Body)
	return nil
}

func sendNotificationToSlack(targetURL string, m *models.ConfigUpdateHistory, app *models.App) error {
	text := configUpdateHistoryToNotificationText(m, app)
	data := make(map[string]interface{})
	data["name"] = "Instafig"
	data["icon_url"] = `https://avatars0.githubusercontent.com/u/1274781`
	data["text"] = text
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}
	v := url.Values{}
	v.Set("payload", string(json))
	resp, err := http.PostForm(targetURL, v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = ioutil.ReadAll(resp.Body)
	return nil
}

func TriggerWebHooks(m *models.ConfigUpdateHistory, app *models.App) {
	var globalHooks, appHooks []*models.WebHook
	globalHooks, _ = models.GetGlobalWebHooks(nil)
	appHooks, _ = models.GetWebHooksByAppKey(nil, app.Key)

	for _, hook := range append(globalHooks, appHooks...) {
		switch hook.Target {
		case models.WEBHOOK_TARGET_PUBU:
			sendNotificationToPubu(hook.URL, m, app)
		case models.WEBHOOK_TARGET_SLACK:
			sendNotificationToSlack(hook.URL, m, app)
		}
	}
}

// Rest API

func GetGlobalWebHooks(c *gin.Context) {
	hooks, err := models.GetGlobalWebHooks(nil)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	Success(c, hooks)
}

func GetAppWebHooks(c *gin.Context) {
	appKey := c.Param("app_key")
	apps, err := models.GetWebHooksByAppKey(nil, appKey)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}
	Success(c, apps)
}

func NewWebHook(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	var data struct {
		AppKey string `json:"app_key" binding:"required"`
		Scope  int    `json:"scope"`
		Target string `json:"target" binding:"required"`
		URL    string `json:"url" binding:"required"`
		Status int    `json:"status"`
	}
	if err := c.BindJSON(&data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if data.Scope != models.WEBHOOK_SCOPE_GLOBAL && data.Scope != models.WEBHOOK_SCOPE_APP {
		Error(c, BAD_REQUEST, "unknown webHook scope: "+string(data.Scope))
		return
	}

	if data.Target != models.WEBHOOK_TARGET_PUBU && data.Target != models.WEBHOOK_TARGET_SLACK {
		Error(c, BAD_REQUEST, "unsupported webHook target: "+data.Target)
		return
	}

	webHook := &models.WebHook{
		Key:    utils.GenerateKey(),
		AppKey: data.AppKey,
		Scope:  data.Scope,
		Target: data.Target,
		URL:    data.URL,
		Status: data.Status,
	}
	if _, err := updateWebHook(webHook, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(webHook, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}

func UpdateWebHook(c *gin.Context) {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	var data struct {
		Key    string `json:"key" binding:"required"`
		AppKey string `json:"app_key" binding:"required"`
		Scope  int    `json:"scope"`
		Target string `json:"target" binding:"required"`
		URL    string `json:"url" binding:"required"`
		Status int    `json:"status"`
	}

	if err := c.BindJSON(&data); err != nil {
		Error(c, BAD_POST_DATA, err.Error())
		return
	}

	if data.Scope != models.WEBHOOK_SCOPE_GLOBAL && data.Scope != models.WEBHOOK_SCOPE_APP {
		Error(c, BAD_REQUEST, "unknown webHook scope: "+string(data.Scope))
		return
	}

	if data.Target != models.WEBHOOK_TARGET_PUBU && data.Target != models.WEBHOOK_TARGET_SLACK {
		Error(c, BAD_REQUEST, "unsupported webHook target: "+data.Target)
		return
	}

	var oldHook *models.WebHook = nil
	if data.Scope == models.WEBHOOK_SCOPE_GLOBAL {
		for _, hook := range memConfGlobalWebHooks {
			if hook.Key == data.Key {
				oldHook = hook
				break
			}
		}
	} else if data.Scope == models.WEBHOOK_SCOPE_APP {
		_, ok := memConfAppWebHooks[data.AppKey]
		if !ok {
			Error(c, BAD_REQUEST, "app key not exists: "+data.AppKey)
			return
		}
		for _, hook := range memConfAppWebHooks[data.AppKey] {
			if hook.Key == data.Key {
				oldHook = hook
				break
			}
		}
	}
	if oldHook == nil {
		Error(c, BAD_REQUEST, "webHook key not exists: "+data.Key)
		return
	}

	webHook := *oldHook
	webHook.Target = data.Target
	webHook.URL = data.URL
	webHook.Status = data.Status
	if _, err := updateWebHook(&webHook, nil); err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	failedNodes := syncData2SlaveIfNeed(&webHook, getOpUserKey(c))
	if len(failedNodes) > 0 {
		Success(c, map[string]interface{}{"failed_nodes": failedNodes})
	} else {
		Success(c, nil)
	}
}
