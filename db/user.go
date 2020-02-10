package db

type User struct {
	// ID is constraint by NOT NULL AUTO_INCREMENT
	// marked as "omitempty", so ID will be auto-generated when insert
	ID             uint32  `db:"id,omitempty" json:"id"`
	UserName       string  `db:"username" json:"username"`
	GithubUserName *string `db:"github_username" json:"githubUsername"`
	AvatarUrl      *string `db:"avatar_url" json:"avatarUrl"`
}
