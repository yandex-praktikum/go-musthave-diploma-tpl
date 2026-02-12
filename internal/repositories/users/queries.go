package repositoriesusers

func getUserQuery() string {
	return `
	SELECT * FROM users WHERE login = $1
	`
}

func insertUserQuery() string {
	return `
	INSERT INTO users (login, password) VALUES ($1, $2)
	RETURNING uuid
	`
}
