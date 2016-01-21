package main

import (
	"log"
	"net/http"

	"time"

	"os"
	"path/filepath"
	"strconv"

	"github.com/appwilldev/Instafig/conf"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

func main() {
	if conf.IsEasyDeployMode() && !conf.IsMasterNode() {
		if err := slaveCheckMaster(); err != nil {
			log.Panicf("slave node failed to check master: %s", err.Error())
		}

		go func() {
			for {
				time.Sleep(60 * time.Second)
				slaveCheckMaster()
			}
		}()
	}

	wd, _ := os.Getwd()
	pidFile, err := os.OpenFile(filepath.Join(wd, "instafig.pid"), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Panicf("failed to create pid file: %s", err.Error())
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

	// static
	ginIns.Static("web", "./web")

	// misc api
	miscAPIGroup := ginIns.Group("/misc")
	{
		miscAPIGroup.GET("/version", VersionHandler)
	}

	// client api
	clientAPIGroup := ginIns.Group("/client")
	{
		clientAPIGroup.GET("/config", ClientReqData)
	}

	// op api
	opAPIGroup := ginIns.Group("/op")
	{
		opAPIGroup.POST("/login", Login)

		opAPIGroup.GET("/users/:page", OpAuth, GetUsers)
		opAPIGroup.POST("/user", OpAuth, ConfWriteCheck, NewUser)
		opAPIGroup.POST("/user/init", ConfWriteCheck, InitUser)

		opAPIGroup.GET("/apps/user/:user_key", OpAuth, GetApps)
		opAPIGroup.GET("/apps/all/:page", OpAuth, GetAllApps)
		opAPIGroup.POST("/app", OpAuth, ConfWriteCheck, NewApp)
		opAPIGroup.PUT("/app", OpAuth, ConfWriteCheck, UpdateApp)

		opAPIGroup.GET("/configs/:app_key", OpAuth, GetConfigs)
		opAPIGroup.POST("/config", OpAuth, ConfWriteCheck, NewConfig)
		opAPIGroup.PUT("/config", OpAuth, ConfWriteCheck, UpdateConfig)
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
