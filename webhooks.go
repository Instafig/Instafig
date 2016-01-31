package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/appwilldev/Instafig/models"
	"io/ioutil"
	"net/http"
	"net/url"
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
