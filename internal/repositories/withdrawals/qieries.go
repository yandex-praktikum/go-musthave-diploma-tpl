package withdrawals

func insertWithdrawlQuery() string {
	return `
	INSERT INTO withdrawals ("order", user_uuid, sum, processed_at)
	VALUES ($1, $2, $3, $4);
	`
}

func getWithdrawalsQuery() string {
	return `
	SELECT "order",
	       sum,
	       processed_at
	FROM withdrawals
	WHERE user_uuid = $1
	ORDER BY processed_at DESC
	`
}
