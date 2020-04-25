package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
)

func parseFlags() {
	pflag.IntP("port", "p", 8000, "HTTP server binding port")
	pflag.StringP("database", "d", "", "DatabaseUrl")
	pflag.StringP("migrate", "m", "", "Run migrations up/down or don't run if empty")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatal(err)
	}
}
