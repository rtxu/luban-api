package db

import (
	"encoding/json"
	"errors"

	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

// UserService encapsulate the operations on the `user` table
type UserService interface {
	Find(username string) (User, error)
	FindByGithubUserName(username string) (User, error)

	Insert(user User) error
	NewUser(user *User) error
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

func (s *userService) NewUser(user *User) error {
	return s.table.InsertReturning(user)
}

func (s *userService) Update(username string, toUpdate map[string]interface{}) error {
	res := s.table.Find(db.Cond{"username": username})
	return res.Update(toUpdate)
}

type memUserService struct {
	id    uint32
	table map[uint32]*User
}

// Used under unit-test enviroment
func NewMemUserService() UserService {
	return &memUserService{
		table: make(map[uint32]*User),
	}
}

func (s *memUserService) find(username string) *User {
	var result *User
	for _, v := range s.table {
		if v.UserName == username {
			return v
		}
	}
	return result
}
func (s *memUserService) Find(username string) (User, error) {
	user := s.find(username)
	if user == nil {
		return User{}, ErrNotFound
	} else {
		return *user, nil
	}
}

func (s *memUserService) FindByGithubUserName(username string) (User, error) {
	var result *User
	for _, v := range s.table {
		if v.GithubUserName != nil && *v.GithubUserName == username {
			result = v
			break
		}
	}
	if result == nil {
		return User{}, ErrNotFound
	} else {
		return *result, nil
	}
}

func (s *memUserService) Insert(user User) error {
	s.table[s.id] = &user
	s.id++
	return nil
}

func (s *memUserService) NewUser(user *User) error {
	user.ID = s.id
	s.table[user.ID] = user
	s.id++
	return nil
}

func (s *memUserService) Update(username string, toUpdate map[string]interface{}) error {
	result := s.find(username)
	if result == nil {
		return ErrNotFound
	}
	for k, v := range toUpdate {
		switch k {
		case "root_dir":
			result.RootDir = v.(json.RawMessage)
		default:
			panic("Not Implemented")
		}
	}
	return nil
}
