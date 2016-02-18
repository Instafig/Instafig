package main

import (
	"net/http"

	"github.com/appwilldev/Instafig/conf"
	"github.com/gin-gonic/gin"
)

const (
	SERVER_ERROR = iota
	BAD_REQUEST
	BAD_POST_DATA
	NOT_PERMITTED
	DATA_EXPIRED
	DATA_SYNCING
	DATA_VERSION_ERROR

	NOT_LOGIN
	USER_NOT_EXIST
	USER_NOT_INIT
	PASS_CODE_ERROR
)

var (
	errorStr = map[int][2]string{
		SERVER_ERROR:       [2]string{"server_error", "server error"},
		BAD_REQUEST:        [2]string{"bad_request", "bad requeset"},
		BAD_POST_DATA:      [2]string{"bad_post_data", "bad request body"},
		NOT_PERMITTED:      [2]string{"not_permitted", "not permitted"},
		DATA_EXPIRED:       [2]string{"data_expired", "conf data expired, try from anthor node"},
		DATA_SYNCING:       [2]string{"data_syncing", "conf data syncing, try from anthor node"},
		DATA_VERSION_ERROR: [2]string{"data_verison_error", "data version error"},
		NOT_LOGIN:          [2]string{"not_login", "need login"},
		USER_NOT_EXIST:     [2]string{"user_not_exist", "user not exist"},
		USER_NOT_INIT:      [2]string{"user_not_init", "need init user first"},
		PASS_CODE_ERROR:    [2]string{"pass_code_error", "user passcode wrong"},
	}
)

func Success(c *gin.Context, data interface{}) {
	res := gin.H{"status": true}
	if data != nil {
		res["data"] = data
	}

	setServiceStatus(c, true)
	c.JSON(http.StatusOK, res)
}

func Error(c *gin.Context, errorCode int, data ...interface{}) {
	var (
		errCodeStr = errorStr[errorCode][0]
		errMsg     = errorStr[errorCode][1]
	)

	if len(data) >= 1 {
		if data[0] != nil {
			errMsg = data[0].(string)
		}
	}

	setServiceStatus(c, false)
	c.JSON(http.StatusOK, gin.H{"status": false, "code": errCodeStr, "msg": errMsg})
}

func setServiceStatus(c *gin.Context, status bool) {
	c.Set("_service_status_", status)
}

func getServiceStatus(c *gin.Context) bool {
	i, exists := c.Get("_service_status_")
	if !exists || i == nil {
		return false
	}

	data := i.(bool)

	return data
}

func setOpUserKey(c *gin.Context, key string) {
	c.Set("_user_key_", key)
}

func getOpUserKey(c *gin.Context) string {
	i, exists := c.Get("_user_key_")
	if !exists || i == nil {
		return ""
	}

	data := i.(string)

	return data
}

// misc handlers

func VersionHandler(c *gin.Context) {
	c.String(200, conf.VersionString())
}
