package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

type HostCheck struct {
	Hostname string `json:"host"`
}

type StatusCheck struct {
	Status string `json:"status"`
}

func main() {
	parseFlags()
	serverPort := fmt.Sprintf(":%d", viper.GetInt("port"))

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		hostName, err := os.Hostname()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{ "error": err.Error() })
		} else {
			c.JSON(http.StatusOK, HostCheck{hostName})
		}
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, StatusCheck{"OK"})
	})

	router.Run(serverPort)
}