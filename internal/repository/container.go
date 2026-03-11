package repository

// RepositoryContainer — интерфейс для контейнера репозиториев.
type RepositoryContainer interface {
	UserRepository() UserRepository
	OrderRepository() OrderRepository
	WithdrawalRepository() WithdrawalRepository
}
