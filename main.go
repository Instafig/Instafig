package main

import (
	"log"
	"net/http"

	"time"

	"os"
	"path/filepath"
	"strconv"

	"strings"

	"github.com/appwilldev/Instafig/conf"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

func main() {
	if conf.IsEasyDeployMode() && !conf.IsMasterNode() {
		if err := slaveCheckMaster(); err != nil {
			log.Printf("slave node failed to check master: %s", err.Error())
			os.Exit(1)
		}

		go func() {
			for {
				time.Sleep(time.Duration(conf.CheckMasterInerval) * time.Second)
				slaveCheckMaster()
			}
		}()
	}

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
		clientAPIGroup.GET("/config", ClientConf)
	}
	// compatible with old awconfig
	ginIns.GET("/conf", ClientConf)

	// op api
	opAPIGroup := ginIns.Group("/op")
	{
		opAPIGroup.POST("/login", Login)
		opAPIGroup.POST("/logout", Logout)

		opAPIGroup.GET("/users/:page/:count", InitUserCheck, OpAuth, GetUsers)
		opAPIGroup.POST("/user", OpAuth, ConfWriteCheck, NewUser)
		opAPIGroup.PUT("/user", OpAuth, ConfWriteCheck, UpdateUser)
		opAPIGroup.POST("/user/init", ConfWriteCheck, InitUser)
		opAPIGroup.GET("/user/info", OpAuth, GetLoginUserInfo)

		opAPIGroup.GET("/apps/user/:user_key", OpAuth, GetApps)
		opAPIGroup.GET("/apps/all/:page/:count", OpAuth, GetAllApps)
		opAPIGroup.GET("/app/:app_key", OpAuth, GetApp)
		opAPIGroup.POST("/app", OpAuth, ConfWriteCheck, NewApp)
		opAPIGroup.PUT("/app", OpAuth, ConfWriteCheck, UpdateApp)

		opAPIGroup.GET("/webhooks/global", OpAuth, GetGlobalWebHooks)
		opAPIGroup.GET("/webhooks/app/:app_key", OpAuth, GetAppWebHooks)
		opAPIGroup.POST("/webhook", OpAuth, ConfWriteCheck, NewWebHook)
		opAPIGroup.PUT("/webhook", OpAuth, ConfWriteCheck, UpdateWebHook)

		opAPIGroup.GET("/configs/:app_key", OpAuth, GetConfigs)
		opAPIGroup.POST("/config", OpAuth, ConfWriteCheck, NewConfig)
		opAPIGroup.PUT("/config", OpAuth, ConfWriteCheck, UpdateConfig)
		opAPIGroup.GET("/config/history/:config_key", OpAuth, GetConfigUpdateHistory)

		opAPIGroup.GET("/nodes", OpAuth, GetNodes)
	}

	if conf.IsEasyDeployMode() {
		ginInsNode := gin.New()
		if conf.DebugMode {
			ginInsNode.Use(gin.Logger())
		}
		ginInsNode.Use(gin.Recovery())
		ginInsNode.POST("/node/req/:req_type", NodeRequestHandler)

		gracehttp.Serve(
			&http.Server{Addr: conf.HttpAddr, Handler: ginIns},
			&http.Server{Addr: conf.NodeAddr, Handler: ginInsNode})
	} else {
		gracehttp.Serve(&http.Server{Addr: conf.HttpAddr, Handler: ginIns})
	}
}
