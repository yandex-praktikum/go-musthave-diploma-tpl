package repositories

type OperationRepositoryInterface interface {
	CreateAccrual(accountID uint, orderNumber string, sum float32) error
}
