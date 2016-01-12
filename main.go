package Instafig

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

	s := http.Server{Addr:conf.HttpAddr, Handler: ginIns}

	ginInsNode := gin.New()
	ginInsNode.Use(gin.Recovery())

	ginInsNode.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello from gin node")
	})

	s2 := http.Server{Addr:conf.NodeAddr, Handler: ginInsNode}

	gracehttp.Serve(s, s2)
}

