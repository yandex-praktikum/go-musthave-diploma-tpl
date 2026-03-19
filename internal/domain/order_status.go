package domain

type OrderStatus string

func (s OrderStatus) String() string {
	return string(s)
}
