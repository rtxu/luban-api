package db

import (
	"errors"
	"log"

	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"

	"github.com/rtxu/luban-api/config"
)

var settings = mysql.ConnectionURL{
	Host:     config.Mysql.Host,
	User:     config.Mysql.User,
	Password: config.Mysql.Password,
	Database: config.Mysql.Database,
}

var defaultClient sqlbuilder.Database

var (
	// errors could be handled
	ErrNotFound = errors.New(`no more rows in this result set`)
)

func init() {
	var err error
	defaultClient, err = mysql.Open(settings)
	if err != nil {
		log.Fatalf("db.Open(): %v\n", err)
	}
}
