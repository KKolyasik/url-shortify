package storage

import (
	"database/sql"

	"github.com/KKolyasik/url-shortify/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(url string) *PostgresDB {
	db, err := sql.Open("pgx", url)
	if err != nil {
		logger.Log.Sugar().Fatalw("Неудалось подключиться к БД", "err", err)
	}
	return &PostgresDB{
		db: db,
	}
}

func (p *PostgresDB) Ping() error {
	return p.db.Ping()
}