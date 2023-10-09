package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
)

type BalancePostgres struct {
	db *sqlx.DB
}

func NewBalancePostgres(db *sqlx.DB) *BalancePostgres {
	return &BalancePostgres{db: db}
}

func (b *BalancePostgres) GetBalance(user_id int) (models.Balance, error) {
	balance := make([]models.Balance, 0)

	query := `SELECT -SUM(VT.withdrawn) withdrawn, (SUM(VT.accrual) + SUM(VT.withdrawn)) current   from
                    (SELECT user_id, -SUM (sum) withdrawn, 0 accrual FROM withdrawns WHERE user_id=$1 group by user_id
                        UNION ALL
                            SELECT user_id, 0 withdrawn, SUM (sum) accrual FROM orders WHERE user_id=$2 group by user_id) VT GROUP BY VT.user_id`

	err := b.db.Select(&balance, query, user_id, user_id)

	if err != nil {
		return balance[0], err
	}

	return balance[0], nil
}

func (b *BalancePostgres) DoWithdraw(user_id int, withdraw models.Withdraw) error {

	query := `INSERT INTO withdrawns (number, user_id, sum) values ($1, $2, $3)
                    ON CONFLICT (number) DO UPDATE SET number =  EXCLUDED.number, sum =  EXCLUDED.sum`
	_, err := b.db.Exec(query, withdraw.Order, user_id, withdraw.Sum)

	if err != nil {
		return err
	}

	return nil
}
