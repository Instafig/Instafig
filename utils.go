package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	SERVER_ERROR = iota
	BAD_REQUEST
	BAD_POST_DATA
	LOGIN_NEEDED
	LOGIN_FAILED
	NOT_PERMITTED
	DATA_EXPIRED
	DATA_SYNCING
)

var (
	errorStr = map[int][2]string{
		SERVER_ERROR:  [2]string{"server_error", "服务器错误"},
		BAD_REQUEST:   [2]string{"bad_request", "客户端请求错误"},
		BAD_POST_DATA: [2]string{"bad_post_data", "客户端请求体错误"},
		LOGIN_NEEDED:  [2]string{"login_needed", "未登录"},
		LOGIN_FAILED:  [2]string{"login_failed", "登录失败"},
		NOT_PERMITTED: [2]string{"not_permitted", "无权进行此次操作"},
		DATA_EXPIRED:  [2]string{"data_expired", "该节点数据可能已过期"},
		DATA_SYNCING:  [2]string{"data_expired", "该节点正在同步数据"},
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
