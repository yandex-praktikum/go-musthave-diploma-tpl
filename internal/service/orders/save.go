package orders

func (s *Service) Save(query string, args ...interface{}) error {
	return s.db.Save(query, args...)
}

func (s *Service) SaveTableUser(login, passwordHash string) error {
	return s.db.SaveTableUser(login, passwordHash)
}

func (s *Service) SaveTableUserAndUpdateToken(login, accessToken string) error {
	return s.db.SaveTableUserAndUpdateToken(login, accessToken)
}
