package db

import (
	"database/sql"
)

type User struct {
	// ID is constraint by NOT NULL AUTO_INCREMENT
	// marked as "omitempty", so ID will be auto-generated when insert
	ID             uint32         `db:"id,omitempty"`
	UserName       string         `db:"username"`
	GithubUserName sql.NullString `db:"github_username"`
}
