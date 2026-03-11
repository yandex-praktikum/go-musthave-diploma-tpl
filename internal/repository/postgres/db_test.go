package postgres_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/MaxRadzey/go-musthave-diploma-tpl/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	testDSN string
	testDB  *sql.DB
)

const defaultTestDSN = "postgres://shortener:shortener@localhost:5432/shortener?sslmode=disable"

func TestMain(m *testing.M) {
	testDSN = os.Getenv("TEST_DATABASE_DSN")
	if testDSN == "" {
		testDSN = defaultTestDSN
	}

	if err := storage.RunMigrations(testDSN, "../../../migrations"); err != nil {
		os.Stderr.WriteString("migrations: " + err.Error() + "\n")
		os.Exit(1)
	}

	var err error
	testDB, err = sql.Open("pgx", testDSN)
	if err != nil {
		os.Stderr.WriteString("open db: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer testDB.Close()

	os.Exit(m.Run())
}
