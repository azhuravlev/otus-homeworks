package main

import (
	"github.com/gramework/gramework"
	"os"
)

type HostCheck struct {
	Hostname string `json:"host"`
}

type StatusCheck struct {
	Status string `json:"status"`
}

func main() {
	app := gramework.New()

	app.GET("/", func() interface{} {
		hostName,_ := os.Hostname()
		return HostCheck{hostName}
	})

	app.GET("/health/", func() interface{} {
		return StatusCheck{"OK"}
	})

	app.ListenAndServe()
}