package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	// Импорт с _ означает, что мы подключаем пакет только для его побочных эффектов (например, регистрация драйвера БД), но не используем его функции напрямую.
	// Что будет, если убрать _? При вызове sql.Open("pgx", dsn) получите ошибку: sql: unknown driver "pgx" (драйвер не зарегистрирован).
)

type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10              // Максимум открытых соединений
const maxIdleDbConn = 5               // Сколько соединений держать в режиме ожидания
const maxDbLifetime = 5 * time.Minute // Макс. время жизни соединения

func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDataBase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	// пробрасываем наше соединение в свойство SQL типа DB
	dbConn.SQL = d

	err = testDB(d)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// testDB tries to ping the database
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

func NewDataBase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
