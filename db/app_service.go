package db

import (
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

// AppService encapsulate the operations on the `app` table
type AppService interface {
	NewApp(app *App) error
}

type appService struct {
	table db.Collection
}

func NewAppService(dbConn sqlbuilder.Database) AppService {
	const kTableName = "app"
	return &appService{
		table: dbConn.Collection(kTableName),
	}
}

func (s *appService) NewApp(app *App) error {
	return s.table.InsertReturning(app)
}

type memAppService struct {
	id    uint32
	table map[uint32]*App
}

// Used under unit-test enviroment
func NewMemAppService() AppService {
	return &memAppService{
		table: make(map[uint32]*App),
	}
}

func (s *memAppService) NewApp(app *App) error {
	app.ID = s.id
	s.id++
	s.table[app.ID] = app
	return nil
}
