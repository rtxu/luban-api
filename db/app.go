package db

type App struct {
	// ID is constraint by NOT NULL AUTO_INCREMENT
	// marked as "omitempty", so ID will be auto-generated when insert
	ID      uint32 `db:"id,omitempty" json:"id"`
	OwnerID uint32 `db:"owner_id" json:"ownerId"`
}
