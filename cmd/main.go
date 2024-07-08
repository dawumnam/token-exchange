package main

import (
	"log"

	"github.com/dawumnam/token-trader/cmd/api"
	"github.com/dawumnam/token-trader/config"
	database "github.com/dawumnam/token-trader/db"
	"github.com/go-sql-driver/mysql"
)

func main() {
	database.Init()

	db, err := database.NewMySQLStorage(mysql.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 config.Envs.DBAddress,
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})

	if err != nil {
		log.Fatal(err)
	}

	database.InitDatabase(db)

	server := api.NewAPIServer(":8080", db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
