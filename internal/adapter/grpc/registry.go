package grpc

import (
	"database/sql"
	"net/http"

	"gorm.io/gorm"
)

// Deps holds shared dependencies used by service registrars.
// Note: prefer Gorm when using MySQL repositories implemented with GORM.
type Deps struct {
	// Deprecated: kept for legacy code that still expects *sql.DB.
	MySQL *sql.DB
	// Preferred ORM handle for MySQL-backed repositories.
	Gorm *gorm.DB
}

// Registrar registers handlers onto the mux using provided deps.
type Registrar func(mux *http.ServeMux, deps Deps)

var registrars []Registrar

// Add registers a Registrar to be called by RegisterAll.
func Add(r Registrar) { registrars = append(registrars, r) }

// RegisterAll invokes all registered Registrars.
func RegisterAll(mux *http.ServeMux, deps Deps) {
	for _, r := range registrars {
		r(mux, deps)
	}
}
