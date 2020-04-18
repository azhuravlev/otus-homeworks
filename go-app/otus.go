package main

import (
	"github.com/gramework/gramework"
)

type StatusCheck struct {
	Status string `json:"status"`
}

func main() {
	app := gramework.New()

	app.GET("/health/", func() interface{} {
		return StatusCheck{"OK"}
	})

	app.ListenAndServe()
}