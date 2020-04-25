package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type User struct {
	Id        int
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func initUsersEndpoints(router *gin.Engine) {
	router.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, StatusCheck{"OK"})
	})
}
