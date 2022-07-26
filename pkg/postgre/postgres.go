package postgre

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

func NewClient(dsn string) (pool *pgxpool.Pool, err error) {
	pool, err = pgxpool.Connect(context.Background(), dsn)

	if err != nil {
		log.Print("error do with tries postgresql")
	}
	q := `CREATE TABLE users(
	id serial primary key,
    username VARCHAR(100),
	password VARCHAR(100)
	);`
	_, err = pool.Exec(context.Background(), q)
	if err != nil {
		log.Print("ТАБЛИЦА НЕ СОЗДАНА")
		log.Print(err)
	}

	q = `CREATE UNIQUE INDEX username_unique
  ON users
 USING btree(username);
`
	_, err = pool.Exec(context.Background(), q)
	if err != nil {
		log.Print("UNIQUE НЕ СОЗДАНА")
		log.Print(err)
	}
	return pool, nil
}
