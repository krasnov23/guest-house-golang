package dbrepo

import (
	"database/sql"
	"github.com/krasnov23/guest-house-golang/internal/config"
	"github.com/krasnov23/guest-house-golang/internal/repository"
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
