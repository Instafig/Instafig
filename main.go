package main

import (
	"net/http"

	"encoding/json"

	"github.com/appwilldev/Instafig/conf"
	"github.com/appwilldev/Instafig/models"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

func main() {
	ginIns := gin.New()
	ginIns.Use(gin.Recovery())

	ginIns.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello from gin")
	})

	ginIns.GET("/users", func(c *gin.Context) {
		users, err := models.GetAllUser(nil)
		if err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}

		bs, _ := json.Marshal(users)
		c.String(http.StatusOK, string(bs))
	})

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

	gracehttp.Serve(
		&http.Server{Addr: conf.HttpAddr, Handler: ginIns},
		&http.Server{Addr: conf.NodeAddr, Handler: ginInsNode})
}
