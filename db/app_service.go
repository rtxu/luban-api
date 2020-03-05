package db

import (
	"encoding/json"
	"errors"

	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

// AppService encapsulate the operations on the `app` table
type AppService interface {
	NewApp(app *App) error
	Find(ownerId, appId uint32) (App, error)

	UpdateContent(ownerId, appId uint32, v json.RawMessage) error
	UpdateLastPublishedContent(ownerId, appId uint32, v json.RawMessage) error
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

func (s *appService) Update(ownerId, appId uint32, toUpdate map[string]interface{}) error {
	res := s.table.Find("owner_id", ownerId).And("id", appId)
	return res.Update(toUpdate)
}
func (s *appService) UpdateContent(ownerId, appId uint32, v json.RawMessage) error {
	return s.Update(ownerId, appId, map[string]interface{}{
		"content": v,
	})
}
func (s *appService) UpdateLastPublishedContent(ownerId, appId uint32, v json.RawMessage) error {
	return s.Update(ownerId, appId, map[string]interface{}{
		"last_published_content": v,
	})
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

func (s *memAppService) find(ownerId, appId uint32) *App {
	for id, app := range s.table {
		if id == appId && app.OwnerID == ownerId {
			return app
		}
	}
	return nil
}

func (s *memAppService) Find(ownerId, appId uint32) (App, error) {
	app := s.find(ownerId, appId)
	if app == nil {
		return App{}, ErrNotFound
	} else {
		return *app, nil
	}
}

func (s *memAppService) Update(ownerId, appId uint32, k string, v interface{}) error {
	app := s.find(ownerId, appId)
	if app == nil {
		return ErrNotFound
	}
	switch k {
	case "content":
		app.Content = v.(json.RawMessage)
	case "last_published_content":
		app.LastPublishedContent = v.(json.RawMessage)
	default:
		panic("Not Implemented")
	}
	return nil
}
func (s *memAppService) UpdateContent(ownerId, appId uint32, v json.RawMessage) error {
	return s.Update(ownerId, appId, "content", v)
}
func (s *memAppService) UpdateLastPublishedContent(ownerId, appId uint32, v json.RawMessage) error {
	return s.Update(ownerId, appId, "last_published_content", v)
}
