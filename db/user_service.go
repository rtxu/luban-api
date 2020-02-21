package db

import (
	"errors"

	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

// UserService encapsulate the operations on the `user` table
type UserService interface {
	Find(username string) (User, error)
	FindByGithubUserName(username string) (User, error)

	Insert(user User) error
	Update(username string, toUpdate map[string]interface{}) error
}

func NewUserService(dbConn sqlbuilder.Database) UserService {
	const kTableName = "user"
	return &userService{
		table: dbConn.Collection(kTableName),
	}
}

type userService struct {
	table db.Collection
}

func (s *userService) Find(username string) (User, error) {
	res := s.table.Find(db.Cond{"username": username})
	var user User
	err := res.One(&user)
	if errors.Is(err, db.ErrNoMoreRows) {
		return user, ErrNotFound
	}
	return user, err
}

func (s *userService) FindByGithubUserName(username string) (User, error) {
	res := s.table.Find(db.Cond{"github_username": username})
	var user User
	err := res.One(&user)
	if errors.Is(err, db.ErrNoMoreRows) {
		return user, ErrNotFound
	}
	return user, err
}

func (s *userService) Insert(user User) error {
	_, err := s.table.Insert(user)
	return err
}

func (s *userService) Update(username string, toUpdate map[string]interface{}) error {
	res := s.table.Find(db.Cond{"username": username})
	return res.Update(toUpdate)
}
