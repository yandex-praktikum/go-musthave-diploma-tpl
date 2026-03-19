package domain

type OrderNumber string

func (n OrderNumber) String() string {
	return string(n)
}
