package storage

var RegistrationPostgres string = "INSERT INTO personal_account (login, password) VALUES ($1, $2)"
var CheckLoginPostgres = "SELECT password FROM personal_account WHERE login = $1"
var CheckHashPasswordPostgres string = "SELECT password FROM personal_account WHERE login = $1"
var CheckUserOrderPostgres = "SELECT user_id FROM orders WHERE order_number = $1"
var CreateNewOrderPostgres = "INSERT INTO orders (order_number, user_id, status) VALUES ($1, $2,'NEW')"
var LoginIDPostgres string = "SELECT id FROM personal_account WHERE login = $1"
var GetUserOrdersQuery string = `
	SELECT order_number, status, accrual, uploaded_at
	FROM orders
	WHERE user_id = $1
	ORDER BY uploaded_at DESC;
`
var UpdateOrderStatusPostgres string = `
UPDATE orders
SET status = $1, accrual = $2
WHERE order_number = $3
`
var GetOrdersForAccrual string = `
SELECT order_number, accrual, uploaded_at 
FROM orders 
WHERE status IN ('NEW', 'PROCESSING', 'REGISTERED')
`
var GetBalanceIncome string = `
SELECT SUM(accrual)
FROM orders
WHERE status = 'PROCESSED' AND user_id = $1
`
var GetBalanceWithDrawn string = `
SELECT SUM(accrual)
FROM orders
WHERE status = 'WITHDRAWN' AND user_id = $1
`
var AddWithdrawOrderPostgres string = `
INSERT INTO orders (user_id, order_number, accrual, status)
VALUES ($1, $2, $3, 'WITHDRAWN')
`
var GetAllWithDrawals string = `
SELECT order_number, accrual, uploaded_at 
FROM orders
WHERE user_id = $1 AND status = 'WITHDRAWN'
ORDER BY uploaded_at DESC;
`
