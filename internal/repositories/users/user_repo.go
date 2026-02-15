package repositoriesusers

import (
	"context"
	"fmt"
	"sync"

	"github.com/Raime-34/gophermart.git/internal/cache"
	"github.com/Raime-34/gophermart.git/internal/dto"
	dbinterface "github.com/Raime-34/gophermart.git/internal/repositories/db_interface"
)

type UserRepo struct {
	db          dbinterface.DbIface
	cachedUsers *cache.Cache[*dto.UserData]
	mu          sync.RWMutex
}

func NewUserRepo(pool dbinterface.DbIface) *UserRepo {
	return &UserRepo{
		db:          pool,
		cachedUsers: cache.NewCache[*dto.UserData](),
	}
}

func (r *UserRepo) GetUser(ctx context.Context, userInfo dto.UserCredential) (*dto.UserData, error) {
	if userUuid, ok := r.cachedUsers.Get(userInfo.Login); ok {
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
	r.cachedUsers.Set(login, userData)

	return userData, nil
}

func (r *UserRepo) RegisterUser(ctx context.Context, userInfo dto.UserCredential) error {
	row := r.db.QueryRow(ctx, insertUserQuery(), userInfo.Login, userInfo.Password)
	var uuid string
	if err := row.Scan(&uuid); err != nil {
		return fmt.Errorf("RegisterUser - failed to insert user: %v", err)
	}

	userData := userInfo.ToUserData(uuid)
	r.cachedUsers.Set(userData.Login, userData)

	return nil
}
