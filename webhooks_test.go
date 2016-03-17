package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/Instafig/Instafig/models"
	"github.com/Instafig/Instafig/utils"
)

func TestPubuNotifaction(t *testing.T) {
	pubuURL := `https://hooks.pubu.im/services/q2qp7wywyebgfqd`
	hostname, _ := os.Hostname()
	configHistory := &models.ConfigUpdateHistory{
		Id:         "",
		ConfigKey:  "",
		K:          "TestKey",
		OldV:       "",
		OldVType:   "",
		NewV:       "1.0",
		NewVType:   models.CONF_V_TYPE_INT,
		Kind:       models.CONFIG_UPDATE_KIND_NEW,
		UserKey:    "",
		UserName:   fmt.Sprintf("PubuTester@%s", hostname),
		CreatedUTC: utils.GetNowSecond(),
	}
	app := &models.App{
		Name: "TestApp",
	}
	sendNotificationToPubu(pubuURL, configHistory, app)
	configHistory.UserName = fmt.Sprintf("SlackTester@%s", hostname)
	sendNotificationToSlack(pubuURL, configHistory, app)
}
