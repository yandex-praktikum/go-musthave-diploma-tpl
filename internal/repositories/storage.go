package repositories

import (
	"context"
	"github.com/botaevg/gophermart/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type DBpgx struct {
	Conn *pgxpool.Pool
}

type Storage interface {
	CreateUser(user models.User) (uint, error)
	GetUser(user models.User) (uint, error)
}

func NewDB(pool *pgxpool.Pool) *DBpgx {
	return &DBpgx{Conn: pool}
}

func (d DBpgx) CreateUser(user models.User) (uint, error) {
	q := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id;`

	rows, err := d.Conn.Query(context.Background(), q, user.Username, user.Password)
	defer rows.Close()
	if err != nil {
		log.Print(err)
		log.Print("user not created")
		return 0, err
	}
	var id uint
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Print(err)
			return 0, err
		}

	}
	return id, nil
}

func (d DBpgx) GetUser(user models.User) (uint, error) {
	q := `SELECT id FROM users WHERE username = $1 and password = $2;`
	row, err := d.Conn.Query(context.Background(), q, user.Username, user.Password)
	defer row.Close()

	if err != nil {
		log.Print(err)
		return 0, err
	}
	var id uint
	if row.Next() {
		err := row.Scan(&id)
		if err != nil {
			log.Print(err)
			return 0, err
		}
		log.Print("user found")
		return id, nil
	}
	log.Print("user no found")
	return 0, nil
}
