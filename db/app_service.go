package db

type appService struct{}

// AppService encapsulate the operations on the `app` table
var AppService = &appService{}

const kTableName = "app"

func (s *appService) NewApp(app *App) error {
	tbl := defaultClient.Collection(kTableName)
	err := tbl.InsertReturning(app)
	return err
}
