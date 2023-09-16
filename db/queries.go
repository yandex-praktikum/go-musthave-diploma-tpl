package db

const (
	QueryInsertNewUserInUsers       = "INSERT INTO users (user_id, username, password) VALUES ($1, $2, $3)"
	QueryInsertNewUserInUserBalance = "INSERT INTO user_balance (user_id) VALUES ($1)"
	QueryGetUserIDForExistsUser     = "SELECT user_id FROM users WHERE username=$1 and password=$2"
	QueryInsertNewOrder             = "INSERT INTO orders (user_id, order_id, status) VALUES ($1, $2, $3)"
	QueryGetUserIDForOrder          = "SELECT user_id FROM orders WHERE order_id=$1"
	QueryGetOrdersForUser           = "SELECT order_id,  status, accrual, uploaded_at FROM orders WHERE user_id=$1"
	QueryGetUserBalance             = "SELECT balance::money::numeric::float8 FROM user_balance WHERE user_id=$1"
	QueryGetUserWithdraw            = "SELECT COALESCE(sum(sum)::money::numeric::float8, 0.0) FROM withdraws WHERE user_id=$1"

	// @TODO add user_id check
	QueryCheckExistsOrderID  = "SELECT 1 FROM orders WHERE order_id = $1"
	QueryCheckUserHasMoney   = "SELECT 1 FROM user_balance WHERE user_id=$1 AND balance>=$2"
	QueryAddWithdraw         = "INSERT INTO withdraws (order_id, user_id, sum) VALUES ($1, $2, $3)"
	QueryUpdateUserBalance   = "UPDATE user_balance SET balance=balance-$1 WHERE user_id=$2"
	QueryGetWithdrawsForUser = "SELECT order_id, sum::money::numeric::float8, processed_at FROM withdraws WHERE user_id=$1"
	QueryGetLimitedNewOrders = "SELECT order_id FROM orders WHERE status=1 AND owner_lock=0 LIMIT $1"
	QueryLockOrders          = "UPDATE orders SET owner_lock=$1 WHERE order_id=$2"
	QueryGetLockedNewOrders  = "SELECT order_id FROM orders WHERE status=1 AND owner_lock=$1"

	QueryUpdateProcessedOrder = "UPDATE orders SET status=$1, owner_lock=0, accrual=%2 WHERE order_id=$3"
)
