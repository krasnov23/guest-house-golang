package dbrepo

import (
	"bookings-udemy/internal/config"
	"bookings-udemy/internal/repository"
	"database/sql"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresDBRepo(conn *sql.DB, app *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{App: app, DB: conn}
}

func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}
