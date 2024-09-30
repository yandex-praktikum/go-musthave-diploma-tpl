package orders

func (s *Service) GetAccrual(addressAccrual string) {
	s.db.GetAccrual(addressAccrual)
	return
}
