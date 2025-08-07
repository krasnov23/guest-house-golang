package main

import (
	"bookings-udemy/internal/config"
	"fmt"
	"github.com/go-chi/chi"
	"testing"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:

	default:
		t.Error(fmt.Sprintf("type is not *chi.Mux type is %T", v))
	}
}
