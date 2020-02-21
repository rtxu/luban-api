package main

import (
	"log"
	"net/http"

	"github.com/rtxu/luban-api/config"
	"github.com/rtxu/luban-api/server"
	"upper.io/db.v3/mysql"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}
	// fmt.Printf("=== app config ===\n%+v\n", conf)

	dbConn, err := mysql.Open(mysql.ConnectionURL{
		User:     conf.Mysql.User,
		Host:     conf.Mysql.Host,
		Password: conf.Mysql.Password,
		Database: conf.Mysql.Database,
	})
	if err != nil {
		return err
	}
	defer dbConn.Close()

	svr := server.New(conf)
	svr.SetupDBService(dbConn)
	return http.ListenAndServe(":9090", svr)
}
