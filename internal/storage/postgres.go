package storage

import (
	"context"
	"errors"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/KKolyasik/url-shortify/internal/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, url string) *PostgresDB {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		logger.Log.Sugar().Fatalw("Не удалось подключиться к БД", "err", err)
	}
	err = runMigrations(pool)
	if err != nil {
		logger.Log.Sugar().Fatalw("Не удалось выполнить миграции", "err", err)
	}
	return &PostgresDB{
		pool: pool,
	}
}

func (p *PostgresDB) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *PostgresDB) GetIDByURL(ctx context.Context, u string) (model.URL, error) {
	var id string
	i := sq.Select("short_code").
		From("urls").
		Where(sq.Eq{"original_url": u}).
		PlaceholderFormat(sq.Dollar)
	query, args, err := i.ToSql()
	if err != nil {
		return model.URL{}, err
	}
	row := p.pool.QueryRow(ctx, query, args...)
	err = row.Scan(&id)
	if err == pgx.ErrNoRows {
		return model.URL{}, nil
	}
	if err != nil {
		return model.URL{}, err
	}

	return model.URL{ShortCode: id}, nil
}

func (p *PostgresDB) GetURLByID(ctx context.Context, id string) (model.URL, error) {
	var url string
	var isDeleted bool
	u := sq.Select("original_url", "is_deleted").
		From("urls").
		Where(sq.Eq{"short_code": id}).
		PlaceholderFormat(sq.Dollar)
	query, args, err := u.ToSql()
	if err != nil {
		return model.URL{}, err
	}
	row := p.pool.QueryRow(ctx, query, args...)
	err = row.Scan(&url, &isDeleted)
	if err != nil {
		return model.URL{}, err
	}
	return model.URL{
		ShortCode:   id,
		OriginalURL: url,
		IsDeleted:   isDeleted,
	}, nil
}

func (p *PostgresDB) Save(ctx context.Context, id, u string, vid uuid.UUID) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query, args, err := sq.
		Insert("urls").
		Columns("id", "original_url", "short_code", "user_id").
		Values(uuid.New(), u, id, vid).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, args...)

	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
		return err
	}

	var existingShortCode string
	query, args, err = sq.
		Select("short_code").
		From("urls").
		Where(sq.Eq{"original_url": u}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	scanErr := p.pool.QueryRow(ctx, query, args...).Scan(&existingShortCode)

	switch {
	case scanErr == nil:
		return &domainerr.URLAlreadyExistsError{ShortCode: existingShortCode}
	case errors.Is(scanErr, pgx.ErrNoRows):
		return domainerr.ErrShortCodeCollision
	default:
		return scanErr
	}
}

func (p *PostgresDB) HasID(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM urls WHERE short_code=$1)"
	err := p.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *PostgresDB) GetURLIDByUser(ctx context.Context, vid uuid.UUID) ([]model.URL, error) {
	query, args, err := sq.
		Select("original_url", "short_code").
		From("urls").
		Where(sq.Eq{"user_id": vid}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := p.pool.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	urls := make([]model.URL, 0)

	for rows.Next() {
		var url model.URL
		err = rows.Scan(&url.OriginalURL, &url.ShortCode)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return urls, nil
}

func (p *PostgresDB) BatchDelete(ctx context.Context, shortCodes ...string) error {
	logger.Log.Sugar().Infow("Начало удаление BatchDelete")
	batch := pgx.Batch{}
	query := "UPDATE urls SET is_deleted = $1 WHERE short_code = $2"
	for _, code := range shortCodes {
		batch.Queue(query, true, code)
	}

	br := p.pool.SendBatch(ctx, &batch)
	defer br.Close()
	for i := 0; i < len(shortCodes); i++ {
		_, err := br.Exec()
		if err != nil {
			logger.Log.Sugar().Infow("Ошибка выполнения запроса в batch", "err", err)
			return err
		}
	}

	logger.Log.Sugar().Infow("Успешно удалили")
	return nil
}

func runMigrations(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "pgx5", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}
