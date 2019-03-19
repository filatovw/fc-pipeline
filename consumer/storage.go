package main

import (
	"context"

	"github.com/filatovw/fc-pipeline/libs/config"
	"github.com/filatovw/fc-pipeline/libs/queue"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type storage struct {
	db    *sqlx.DB
	query *sqlx.Stmt
}

func (s *storage) Close() error {
	return s.db.Close()
}

func newStorage(cfg config.DB) (*storage, error) {
	db, err := sqlx.Connect("postgres", cfg.ConnectionString("userdata"))
	if err != nil {
		return nil, errors.Wrap(err, "new storage, connection")
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "new storage, ping")
	}
	q, err := db.Preparex(db.Rebind(`INSERT INTO contacts (name, email) VALUES (?, ?);`))
	if err != nil {
		return nil, errors.Wrap(err, "new storage, precompile query")
	}
	return &storage{
		db:    db,
		query: q,
	}, nil
}

func (s *storage) Insert(ctx context.Context, msg queue.Message) error {
	if _, err := s.query.ExecContext(ctx, msg.Name, msg.Email); err != nil {
		return errors.Wrapf(err, "insert message")
	}
	return nil
}
