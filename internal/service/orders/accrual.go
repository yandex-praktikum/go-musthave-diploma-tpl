package orders

import "database/sql"

func (s *Service) Save(query string, args ...interface{}) error {
	return s.db.Save(query, args...)
}

func (s *Service) Get(query string, args ...interface{}) (*sql.Row, error) {
	return s.db.Get(query, args...)
}
func (s *Service) Gets(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Gets(query, args...)
}
