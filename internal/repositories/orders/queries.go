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
