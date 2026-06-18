package postgres

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/n0f4ph4mst3r/goshort/internal/storage"
)

type Storage struct {
	db *sql.DB
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func New(connStr string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to open database: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: unable to connect to database: %w", op, err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("%s: failed to apply migrations: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	fmt.Println("Running migrations...")
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}

func (s *Storage) SaveURL(ctx context.Context, u, alias string) error {
	const op = "storage.postgres.SaveUrl"

	const query = `
		INSERT INTO url (alias, origin)
		VALUES ($1, $2);
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, query, alias, u)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, storage.ErrUrlExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	eventPayload := map[string]any{
		"alias":  alias,
		"origin": u,
	}

	if err := s.commitWithOutboxEvent(ctx, tx, "UrlSaved", eventPayload); err != nil {
		return fmt.Errorf("%s: failed to commit outbox event: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.GetUrl"

	var u string
	err := s.db.QueryRowContext(ctx, `
		SELECT origin
		FROM url
		WHERE alias = $1;
	`, alias).Scan(&u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUrlNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.DeleteURL"

	const query = `
		DELETE FROM url
		WHERE alias = $1
		RETURNING origin;
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var u string
	if err = tx.QueryRowContext(ctx, query, alias).Scan(&u); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUrlNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	eventPayload := map[string]any{
		"alias":  alias,
		"origin": u,
	}

	if err = s.commitWithOutboxEvent(ctx, tx, "UrlDeleted", eventPayload); err != nil {
		return "", fmt.Errorf("%s: failed to commit outbox event: %w", op, err)
	}

	return u, nil
}

func (s *Storage) commitWithOutboxEvent(
	ctx context.Context,
	tx *sql.Tx,
	eventType string,
	eventPayload map[string]any,
) error {
	const query = `
		INSERT INTO outbox_events (event_type, payload)
		VALUES ($1, $2)
	`

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, eventType, payloadBytes)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
