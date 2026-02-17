package orders

func insertOrderQuery() string {
	return `
	INSERT INTO orders (number, user_uuid, status, accrual)
	VALUES ($1, $2, $3, $4)
	`
}

func updateOrderQuery() string {
	return `
	UPDATE orders
	SET 
	    status = $1,
	    accrual = $2
	WHERE number = $3;
	`
}

func getOrdersQuery() string {
	return `
	SELECT number,
	       status,
	       accrual,
	       uploaded_at
	FROM orders
	WHERE user_uuid = $1
	ORDER BY uploaded_at DESC
	`
}
