package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/appwilldev/Instafig/conf"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

func main() {
	wd, _ := os.Getwd()
	pidFile, err := os.OpenFile(filepath.Join(wd, "instafig.pid"), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("failed to create pid file: %s", err.Error())
		os.Exit(1)
	}
	pidFile.WriteString(strconv.Itoa(os.Getpid()))
	pidFile.Close()

	if conf.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginIns := gin.New()
	ginIns.Use(gin.Recovery())
	if conf.DebugMode {
		ginIns.Use(gin.Logger())
	}

	if conf.WebDebugMode {
		// static
		ginIns.Static("/web", "./web")
	} else {
		// bin static
		ginIns.GET("/web/*file",
			func(c *gin.Context) {
				fileName := c.Param("file")
				if fileName == "/" {
					fileName = "/index.html"
				}
				data, err := Asset("web" + fileName)
				if err != nil {
					c.String(http.StatusNotFound, err.Error())
					return
				}

				switch {
				case strings.LastIndex(fileName, ".html") == len(fileName)-5:
					c.Header("Content-Type", "text/html; charset=utf-8")
				case strings.LastIndex(fileName, ".css") == len(fileName)-4:
					c.Header("Content-Type", "text/css")
				}
				c.String(http.StatusOK, string(data))
			})
	}

	// misc api
	miscAPIGroup := ginIns.Group("/misc")
	{
		miscAPIGroup.GET("/version", VersionHandler)
	}

	// client api
	clientAPIGroup := ginIns.Group("/client")
	{
		clientAPIGroup.GET("/config", StatisticHandler, ClientConf)
	}
	// compatible with old awconfig
	ginIns.GET("/conf", StatisticHandler, ClientConf)

	// op api
	opAPIGroup := ginIns.Group("/op")
	{
		opAPIGroup.POST("/login", Login)
		opAPIGroup.POST("/logout", OpAuth, Logout)

		opAPIGroup.GET("/users/:page/:count", InitUserCheck, OpAuth, GetUsers)
		opAPIGroup.POST("/user", OpAuth, ConfWriteCheck, NewUser, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.PUT("/user", OpAuth, ConfWriteCheck, UpdateUser, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.PUT("/user/status", OpAuth, ConfWriteCheck, UpdateUserStatus, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.PUT("/user/passcode", OpAuth, ConfWriteCheck, UpdateUserPassCode, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.POST("/user/init", ConfWriteCheck, InitUser, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.GET("/user/info", OpAuth, GetLoginUserInfo)

		opAPIGroup.GET("/apps/user/:user_key", OpAuth, GetApps)
		opAPIGroup.GET("/apps/all/:page/:count", OpAuth, GetAllApps)
		opAPIGroup.GET("/app/:app_key", OpAuth, GetApp)
		opAPIGroup.GET("/apps/search", OpAuth, SearchApps)
		opAPIGroup.GET("/apps/search/hint", OpAuth, SearchAppsHint)
		opAPIGroup.POST("/app", OpAuth, ConfWriteCheck, NewApp, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.PUT("/app", OpAuth, ConfWriteCheck, UpdateApp, UpdateMasterLastDataUpdateUTC)

		opAPIGroup.GET("/webhooks/global", OpAuth, GetGlobalWebHooks)
		opAPIGroup.GET("/webhooks/app/:app_key", OpAuth, GetAppWebHooks)
		opAPIGroup.POST("/webhook", OpAuth, ConfWriteCheck, NewWebHook, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.PUT("/webhook", OpAuth, ConfWriteCheck, UpdateWebHook)

		opAPIGroup.GET("/configs/:app_key", OpAuth, GetConfigs)
		opAPIGroup.POST("/config", OpAuth, ConfWriteCheck, NewConfig, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.PUT("/config", OpAuth, ConfWriteCheck, UpdateConfig, UpdateMasterLastDataUpdateUTC)
		opAPIGroup.GET("/config/history/:config_key", OpAuth, GetConfigUpdateHistory)
		opAPIGroup.GET("/config/apphistory/:app_key/:page/:count", OpAuth, GetAppConfigUpdateHistory)
		opAPIGroup.GET("/config/userhistory/:user_key/:page/:count", OpAuth, GetConfigUpdateHistoryOfUser)
		opAPIGroup.GET("/config/by/:config_key", OpAuth, GetConfigByKey)

		opAPIGroup.GET("/nodes", OpAuth, GetNodes)

		// for statistics
		opAPIGroup.GET("/stat/latest-config-device-count/:app_key", OpAuth, StatCheck, GetDeviceCountOfAppLatestConfig)
		opAPIGroup.GET("/stat/app-config-response/:app_key", OpAuth, StatCheck, GetAppConfigResponseData)
		opAPIGroup.GET("/stat/node-config-response/:node_url", OpAuth, StatCheck, GetNodeConfigResponseData)
	}

	if conf.IsEasyDeployMode() {
		ginInsNode := gin.New()
		if conf.DebugMode {
			ginInsNode.Use(gin.Logger())
		}
		ginInsNode.Use(gin.Recovery())
		ginInsNode.POST("/node/req/:req_type", NodeRequestHandler)

		err = gracehttp.Serve(
			&http.Server{Addr: fmt.Sprintf(":%d", conf.Port), Handler: ginIns},
			&http.Server{Addr: conf.NodeAddr, Handler: ginInsNode})
		if err != nil {
			log.Printf("fatal error: %s", err.Error())
		}
	} else {
		err = gracehttp.Serve(&http.Server{Addr: fmt.Sprintf(":%d", conf.Port), Handler: ginIns})
		if err != nil {
			log.Printf("fatal error: %s", err.Error())
		}
	}
}
