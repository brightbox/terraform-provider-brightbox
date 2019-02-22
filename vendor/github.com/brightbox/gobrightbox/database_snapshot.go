package brightbox

import (
	"time"
)

// DatabaseSnapshot represents a snapshot of a database server.
// https://api.gb1.brightbox.com/1.0/#database_snapshot
type DatabaseSnapshot struct {
	Id              string
	Name            string
	Description     string
	Status          string
	Account         Account
	DatabaseEngine  string `json:"database_engine"`
	DatabaseVersion string `json:"database_version"`
	Size            int
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
	Locked          bool
}
