package grpc

import (
	"database/sql"
	"net/http"
)

// Deps holds shared dependencies used by service registrars.
type Deps struct {
	MySQL *sql.DB
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
