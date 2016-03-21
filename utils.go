package main

import (
	"net/http"

	"github.com/Instafig/Instafig/conf"
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
	USER_INACTIVE
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
		USER_INACTIVE:      [2]string{"user_inactive", "user is inactive"},
		PASS_CODE_ERROR:    [2]string{"pass_code_error", "user passcode wrong"},
	}
)

type RequestLogData struct {
	Status bool   `json:"status"`
	Error  string `json:"err"`
	Msg    string `json:"msg"`
}

func Success(c *gin.Context, data interface{}) {
	res := gin.H{"status": true}
	if data != nil {
		res["data"] = data
	}

	setServiceStatus(c, true)
	setRequestLogData(c, &RequestLogData{Status: true})
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

	setRequestLogData(c, &RequestLogData{Status: false, Error: errCodeStr, Msg: errMsg})
	setServiceStatus(c, false)
	setServiceErrorCode(c, errCodeStr)
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

func setServiceErrorCode(c *gin.Context, code string) {
	c.Set("_error_code_", code)
}

func getServiceErrorCode(c *gin.Context) string {
	i, exists := c.Get("_error_code_")
	if !exists || i == nil {
		return ""
	}

	data := i.(string)

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

func setClientData(c *gin.Context, data *ClientData) {
	c.Set("_client_request_data_", data)
}

func getClientData(c *gin.Context) *ClientData {
	i, exists := c.Get("_client_request_data_")
	if !exists || i == nil {
		return nil
	}

	data := i.(*ClientData)

	return data
}

func setRequestLogData(c *gin.Context, data *RequestLogData) {
	if conf.RequestLogEnable {
		c.Set("_request_log_", data)
	}
}

func getRequestLogDataFromContext(c *gin.Context) *RequestLogData {
	i, exists := c.Get("_request_log_")
	if !exists || i == nil {
		return nil
	}

	data := i.(*RequestLogData)

	return data
}

// misc handlers

func VersionHandler(c *gin.Context) {
	c.String(200, conf.VersionString())
}

func sendChanAsync(ch chan interface{}, i interface{}) {
	select {
	case ch <- i: //
	default: //
	}
}

func doEverTask(f func()) {
	exitCh := make(chan struct{})

	go func() {
		for {
			go func() {
				defer func() { exitCh <- struct{}{} }()
				f()
			}()
			<-exitCh
		}
	}()
}
