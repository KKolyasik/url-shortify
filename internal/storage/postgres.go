package storage

import (
	"context"
	"database/sql"

	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/google/uuid"
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

func (p *PostgresDB) GetIDByURL(ctx context.Context, u string) (string, error) {
	var id string
	query := "SELECT short_code FROM urls WHERE original_url=$1"
	row := p.db.QueryRowContext(ctx, query, u)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *PostgresDB) GetURLByID(ctx context.Context, id string) (string, error) {
	var url string
	query := "SELECT original_url FROM urls WHERE short_code=$1"
	row := p.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (p *PostgresDB) Save(ctx context.Context, id, u string) error {
	query := `INSERT INTO urls (id, original_url, short_code)
	VALUES ($1, $2, $3)`
	_, err := p.db.ExecContext(ctx, query, uuid.New(), u, id)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresDB) HasID(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM urls WHERE short_code=$1)"
	err := p.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
