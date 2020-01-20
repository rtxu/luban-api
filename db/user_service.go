package db

import (
	"errors"

	"upper.io/db.v3"
)

type userService struct{}

// UserService encapsulate the operations on the `user` table
var UserService = &userService{}

const tableName = "user"

func (s *userService) Find(username string) (User, error) {
	tbl := defaultClient.Collection(tableName)
	res := tbl.Find(db.Cond{"username": username})
	var user User
	err := res.One(&user)
	if errors.Is(err, db.ErrNoMoreRows) {
		return user, ErrNotFound
	}
	return user, err
}

func (s *userService) FindByGithubUserName(username string) (User, error) {
	tbl := defaultClient.Collection(tableName)
	res := tbl.Find(db.Cond{"github_username": username})
	var user User
	err := res.One(&user)
	if errors.Is(err, db.ErrNoMoreRows) {
		return user, ErrNotFound
	}
	return user, err
}

func (s *userService) Insert(user User) error {
	tbl := defaultClient.Collection(tableName)
	_, err := tbl.Insert(user)
	return err
}
