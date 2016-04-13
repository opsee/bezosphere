package main

import (
	"github.com/opsee/bezosphere/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("bezosphere")
	viper.AutomaticEnv()

	server, err := service.New(service.Config{
		SpanxAddress: viper.GetString("spanx_address"),
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start(
		viper.GetString("address"),
		viper.GetString("cert"),
		viper.GetString("cert_key"),
	))
}
