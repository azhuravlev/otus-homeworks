package main

import (
	"github.com/chenjiandongx/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func initMetrics(router *gin.Engine) {
	router.Use(ginprom.PromMiddleware(nil))

	// register the `/metrics` route.
	router.GET("/metrics", ginprom.PromHandler(promhttp.Handler()))
}
