package orders

func (s *Service) Save(query string, args ...interface{}) error {
	return s.db.Save(query, args...)
}
