package storage

import "fmt"

const (
	usersTableName  = "users"
	ordersTableName = "orders"
)

var (
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

	insertOrderTemplate = fmt.Sprintf(`
	INSERT INTO %s (order_number, user_id)
	VALUES ($1, (SELECT user_id FROM users WHERE username = $2))
	ON CONFLICT (order_number) DO NOTHING
	RETURNING order_id
	`, ordersTableName)

	getOrderTemplate = fmt.Sprintf(`
	SELECT u.username 
	FROM %s o
	JOIN %s u ON u.user_id = o.user_id
	WHERE o.order_number = $1
	`, ordersTableName, usersTableName)
)
