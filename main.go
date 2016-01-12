package main

import (
	"net/http"

	"github.com/appwilldev/Instafig/conf"
	"github.com/gin-gonic/gin"
	"github.com/facebookgo/grace/gracehttp"
)

func main() {
	ginIns := gin.New()
	ginIns.Use(gin.Recovery())

	ginIns.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello from gin")
	})

	if conf.DebugMode {
		ginIns.Use(gin.Logger())
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &http.Server{Addr:conf.HttpAddr, Handler: ginIns}

	ginInsNode := gin.New()
	ginInsNode.Use(gin.Recovery())

	ginInsNode.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello from gin node")
	})

	if conf.DebugMode {
		ginInsNode.Use(gin.Logger())
	}

	s2 := &http.Server{Addr:conf.NodeAddr, Handler: ginInsNode}

	gracehttp.Serve(s, s2)
}

