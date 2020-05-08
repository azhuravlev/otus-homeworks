package main

import (
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"strings"
)

func initMetrics(router *gin.Engine) {
	prom := ginprometheus.NewPrometheus("gin")

	prom.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, param := range c.Params {
			if param.Key == "id" {
				url = strings.Replace(url, param.Value, ":id", 1)
				break
			}
		}
		return url
	}

	prom.Use(router)
}
