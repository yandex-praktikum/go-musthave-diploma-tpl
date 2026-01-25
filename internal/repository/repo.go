package repository

import (
	"context"
	"database/sql"
	"time"
)

type Repo struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repo {
	return &Repo{
		conn: conn,
	}
}

func (r *Repo) TestConnectionDB() error {
	//проверкаа подключения
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := r.conn.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
