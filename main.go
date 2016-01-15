package main

import (
	"net/http"

	"github.com/appwilldev/Instafig/conf"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

func main() {
	ginIns := gin.New()
	ginIns.Use(gin.Recovery())

	ginInsNode := gin.New()
	ginInsNode.Use(gin.Recovery())

	ginInsNode.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello from gin node")
	})

	if conf.DebugMode {
		ginIns.Use(gin.Logger())
		ginInsNode.Use(gin.Logger())
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// client api
	clientAPIGroup := ginIns.Group("/client")
	{
		clientAPIGroup.GET("/conf", ClientReqData)
	}

	// op api
	opAPIGroup := ginIns.Group("/op")
	{
		opAPIGroup.GET("/users/:page", GetUsers)
		opAPIGroup.POST("/user", ConfWriteCheck, NewUser)

		opAPIGroup.GET("/apps/:user_key", GetApps)
		opAPIGroup.POST("/app", ConfWriteCheck, NewApp)

		opAPIGroup.GET("/configs/:app_key", GetConfigs)
		opAPIGroup.POST("/config", ConfWriteCheck, NewConfig)
	}

	gracehttp.Serve(
		&http.Server{Addr: conf.HttpAddr, Handler: ginIns},
		&http.Server{Addr: conf.NodeAddr, Handler: ginInsNode})
}
