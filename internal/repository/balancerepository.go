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

func (b *BalancePostgres) GetBalance(userID int) (models.Balance, error) {

	var balance models.Balance

	query := `SELECT SUM(VT.withdrawn) withdrawn, (SUM(VT.accrual) - SUM(VT.withdrawn)) current   from
                    (SELECT user_id, SUM (sum) withdrawn, 0 accrual FROM withdrawns WHERE user_id=$1 group by user_id
                        UNION ALL
                            SELECT user_id, 0 withdrawn, SUM (sum) accrual FROM orders WHERE user_id=$2 group by user_id) VT GROUP BY VT.user_id`

	err := b.db.Get(&balance, query, userID, userID)

	if err != nil {
		return balance, err
	}

	return balance, nil
}

func (b *BalancePostgres) ExistOrder(order int) bool {
	var existOrder bool

	query := `SELECT true from orders WHERE number = $1`

	err := b.db.Get(&existOrder, query, order)

	if err != nil {
		return false
	}
	return existOrder
}

func (b *BalancePostgres) GetWithdraws(userID int) ([]models.WithdrawResponse, error) {
	var withdraws []models.WithdrawResponse

	query := `SELECT number, sum, processed from withdrawns WHERE user_id = $1`

	err := b.db.Select(&withdraws, query, userID)

	if err != nil {
		return withdraws, err
	}
	return withdraws, nil
}

func (b *BalancePostgres) DoWithdraw(userID int, withdraw models.Withdraw) error {

	query := `INSERT INTO withdrawns (number, user_id, sum) values ($1, $2, $3)
                    ON CONFLICT (number) DO UPDATE SET number =  EXCLUDED.number, sum =  EXCLUDED.sum`
	_, err := b.db.Exec(query, withdraw.Order, userID, withdraw.Sum)

	if err != nil {
		return err
	}

	return nil
}
