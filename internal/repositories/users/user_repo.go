package repositoriesusers

import (
	"context"
	"fmt"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/Raime-34/gophermart.git/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db          *pgxpool.Conn
	cachedUsers map[string]*dto.UserData
	mu          sync.RWMutex
}

func NewUserRepo(ctx context.Context, pool *pgxpool.Pool) *UserRepo {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		logger.Fatal("Error while acquiring connection from the database pool")
	}

	return &UserRepo{
		db:          conn,
		cachedUsers: make(map[string]*dto.UserData),
	}
}

func (r *UserRepo) GetUser(ctx context.Context, userInfo dto.UserCredential) (*dto.UserData, error) {
	if userUuid, ok := r.getCachedUser(userInfo); ok {
		return userUuid, nil
	}

	var (
		uuid, login, password string
	)
	row := r.db.QueryRow(ctx, getUserQuery(), userInfo.Login)
	if err := row.Scan(&uuid, &login, &password); err != nil {
		return nil, fmt.Errorf("GetUser - failed to scan query result")
	}

	userData := userInfo.ToUserData(uuid)
	r.setCachedUser(userData)

	return userData, nil
}

func (r *UserRepo) getCachedUser(userInfo dto.UserCredential) (*dto.UserData, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userData, ok := r.cachedUsers[userInfo.Login]
	if userData == nil {
		return nil, false
	}

	return userData, ok
}

func (r *UserRepo) setCachedUser(userInfo *dto.UserData) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cachedUsers[userInfo.Login] = userInfo
}

func (r *UserRepo) RegisterUser(ctx context.Context, userInfo dto.UserCredential) error {
	row := r.db.QueryRow(ctx, insertUserQuery(), userInfo.Login, userInfo.Password)
	var uuid string
	if err := row.Scan(&uuid); err != nil {
		return fmt.Errorf("RegisterUser - failed to insert user: %v", err)
	}

	userData := userInfo.ToUserData(uuid)
	r.setCachedUser(userData)

	return nil
}
