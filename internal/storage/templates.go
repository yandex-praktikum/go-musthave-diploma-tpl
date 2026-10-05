package storage

import (
	"fmt"

	"github.com/nikitaw13/go-musthave-diploma/internal/model"
)

const (
	usersTableName       = "users"
	accrualsTableName    = "accruals_journal"
	withdrawalsTableName = "withdrawals_journal"
)

var (
	// User related
	userExistsTemplate = fmt.Sprintf(`
	SELECT EXISTS(
		SELECT 1 
		FROM %s 
		WHERE username = $1)
	`, usersTableName)

	getPasswordHashTemplate = fmt.Sprintf(`
	SELECT password 
	FROM %s 
	WHERE username = $1
	`, usersTableName)

	insertUserTemplate = fmt.Sprintf(`
	INSERT INTO %s (username, password) 
	VALUES ($1, $2)
	`, usersTableName)

	// Accrual related
	insertOrderTemplate = fmt.Sprintf(`
	INSERT INTO %s (order_number, user_id)
	VALUES ($1, (SELECT user_id FROM %s WHERE username = $2))
	ON CONFLICT (order_number) DO NOTHING
	RETURNING order_id
	`, accrualsTableName, usersTableName)

	getOrderOwnerTemplate = fmt.Sprintf(`
	SELECT u.username 
	FROM %s o
	JOIN %s u ON u.user_id = o.user_id
	WHERE o.order_number = $1
	`, accrualsTableName, usersTableName)

	listOrdersTemplate = fmt.Sprintf(`
	SELECT order_number, status, amount, uploaded_at
	FROM %s
	WHERE user_id = (SELECT user_id FROM %s WHERE username = $1)
	ORDER BY uploaded_at DESC
	`, accrualsTableName, usersTableName)

	// Balance related
	getBalanceTemplate = fmt.Sprintf(`
	SELECT
    	COALESCE(a.total_accruals, 0) - COALESCE(w.total_withdrawals, 0) AS current_balance,
    	COALESCE(w.total_withdrawals, 0) AS total_withdrawals
	FROM %s u
	LEFT JOIN (
    	SELECT user_id, SUM(amount) AS total_accruals
    	FROM %s
    	WHERE status = %s
    	GROUP BY user_id
	) a ON a.user_id = u.user_id
	LEFT JOIN (
    	SELECT user_id, SUM(amount) AS total_withdrawals
    	FROM %s
    	GROUP BY user_id
	) w ON w.user_id = u.user_id
	WHERE u.username = $1
	`, usersTableName, accrualsTableName, model.StatusProcessed, withdrawalsTableName)

	// Withdraw-related
	lockUserTemplate = fmt.Sprintf(`
	SELECT user_id 
	FROM %s 
	WHERE username = $1 FOR UPDATE
	`, usersTableName)

	withdrawTemplate = fmt.Sprintf(`
	INSERT INTO %s (user_id, order_number, amount)
	VALUES ((SELECT user_id FROM %s WHERE username = $1), $2, $3)
	ON CONFLICT (order_number) DO NOTHING
	RETURNING withdrawal_id
	`, withdrawalsTableName, usersTableName)

	getWithdrawalOwnerTemplate = fmt.Sprintf(`
	SELECT u.username 
	FROM %s o
	JOIN %s u ON u.user_id = o.user_id
	WHERE o.order_number = $1
	`, withdrawalsTableName, usersTableName)

	listWithdrawalsTemplate = fmt.Sprintf(`
	SELECT order_number, amount, processed_at
    FROM %s
    WHERE user_id = (SELECT user_id FROM %s WHERE username = $1)
    ORDER BY processed_at DESC
	`, withdrawalsTableName, usersTableName)

	// Accrual related
	fetchNewOrdersTemplate = fmt.Sprintf(`
	SELECT order_number
	FROM %s
	WHERE status in ($1, $2)
	`, accrualsTableName)

	updateOrderStatusTemplate = fmt.Sprintf(`
	UPDATE %s
	SET status = $2, amount = $3
	WHERE order_number = $1
	AND status NOT IN ($4, $5)
	`, accrualsTableName)
)
