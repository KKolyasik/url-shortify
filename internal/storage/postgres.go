package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(url string) *PostgresDB {
	db, err := sql.Open("pgx", url)
	if err != nil {
		logger.Log.Sugar().Fatalw("Не удалось подключиться к БД", "err", err)
	}
	err = runMigrations(db)
	if err != nil {
		logger.Log.Sugar().Fatalw("Не удалось выполнить миграции", "err", err)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.ConstraintName {
			case "urls_original_url_key":
				existingID, lookupErr := p.GetIDByURL(ctx, u)
				if lookupErr != nil {
					return lookupErr
				}
				return &domainerr.URLAlreadyExistsError{ShortCode: existingID}
			case "urls_short_code_key":
				return domainerr.ErrShortCodeCollision
			}
		}
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

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}
