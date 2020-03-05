package db

import (
	"errors"

	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

// AppService encapsulate the operations on the `app` table
type AppService interface {
	NewApp(app *App) error
	Find(ownerId, appId uint32) (App, error)
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

func (s *appService) Find(ownerId, appId uint32) (App, error) {
	res := s.table.Find("owner_id", ownerId).And("id", appId)
	var app App
	err := res.One(&app)
	if errors.Is(err, db.ErrNoMoreRows) {
		return app, ErrNotFound
	}
	return app, err
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

func (s *memAppService) Find(ownerId, appId uint32) (App, error) {
	var app App
	for id, app := range s.table {
		if id == appId && app.OwnerID == ownerId {
			return *app, nil
		}
	}
	return app, ErrNotFound
}
