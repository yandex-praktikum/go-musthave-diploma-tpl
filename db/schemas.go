package db

const (
	queryCreateTableUsers = "CREATE TABLE IF NOT EXISTS users (" +
		"user_id uuid PRIMARY KEY," +
		"username VARCHAR(255) NOT NULL UNIQUE," +
		"password VARCHAR(255) NOT NULL" +
		")"
	queryCreateTableOrders = "CREATE TABLE IF NOT EXISTS orders (" +
		"user_id uuid references users(user_id)," +
		"order_id TEXT NOT NULL UNIQUE," +
		"status int, " +
		"owner_lock int DEFAULT 0," +
		"accrual int DEFAULT 0," +
		"uploaded_at timestamp with time zone DEFAULT NOW()" +
		")"

	queryCreateTableUserBalance = "CREATE TABLE IF NOT EXISTS user_balance (" +
		"user_id uuid UNIQUE references users(user_id)," +
		"balance money DEFAULT 0" +
		")"

	queryCreateTableWithdraws = "CREATE TABLE IF NOT EXISTS withdraws (" +
		"order_id TEXT NOT NULL," +
		"user_id uuid references users(user_id)," +
		"sum money NOT NULL," +
		"processed_at timestamp with time zone DEFAULT NOW()" +
		")"
)

var table = []string{
	queryCreateTableUsers,
	queryCreateTableOrders,
	queryCreateTableUserBalance,
	queryCreateTableWithdraws,
}
