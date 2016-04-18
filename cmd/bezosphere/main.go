package main

import (
	"github.com/opsee/bezosphere/service"
	"github.com/opsee/bezosphere/store"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("bezosphere")
	viper.AutomaticEnv()

	db, err := store.NewPostgres(
		viper.GetString("postgres_conn"),
	)

	if err != nil {
		log.Fatal("failed to initialize postgres: ", err)
	}

	server, err := service.New(service.Config{
		SpanxAddress: viper.GetString("spanx_address"),
		Db:           db,
	})

	if err != nil {
		log.Fatal("failed to create new service: ", err)
	}

	log.Fatal(server.Start(
		viper.GetString("address"),
		viper.GetString("cert"),
		viper.GetString("cert_key"),
	))
}
