package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/brisk84/gofemart/domain"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func (s *storage) genToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	return token, nil
}

func (s *storage) Register(ctx context.Context, user domain.User) (string, error) {
	sql1 := `insert into users (login, password) values ($1, $2)`
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	user.Hash = string(hash)

	_, err = tx.Exec(sql1, user.Login, user.Hash)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return "", domain.ErrLoginIsBusy
			}
		}
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	user.Token, err = s.genToken()
	if err != nil {
		return "", err
	}

	s.userMtx.Lock()
	defer s.userMtx.Unlock()
	s.users[user.Token] = user

	return user.Token, nil
}

func (s *storage) Login(ctx context.Context, user domain.User) (bool, string, error) {
	sql1 := `select password from users where login = $1`

	row := s.db.QueryRow(sql1, user.Login)
	err := row.Scan(&user.Hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", nil
		}
		return false, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(user.Password))
	if err != nil {
		return false, "", nil
	}

	user.Token, err = s.genToken()
	if err != nil {
		return false, "", err
	}

	s.userMtx.Lock()
	defer s.userMtx.Unlock()
	s.users[user.Token] = user
	return true, user.Token, nil
}

func (s *storage) Auth(ctx context.Context, token string) (*domain.User, error) {
	s.userMtx.RLock()
	defer s.userMtx.RUnlock()
	if user, ok := s.users[token]; ok {
		return &user, nil
	}
	return nil, nil
}

func (s *storage) UserOrders(ctx context.Context, user domain.User, order int) error {
	var userId string
	sql1 := `select user_id from orders where id = $1`
	sql2 := `insert into orders (id, user_id, cr_dt, status) values ($1, $2, $3, $4)`

	row := s.db.QueryRow(sql1, order)
	err := row.Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = s.db.Exec(sql2, order, user.Login, time.Now(), "NEW")
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	if userId == user.Login {
		return domain.ErrLoadedByThisUser
	}
	return domain.ErrLoadedByAnotherUser
}

func (s *storage) UserOrdersGet(ctx context.Context, user domain.User) ([]domain.Order, error) {
	sql1 := `select id, cr_dt, status from orders where user_id = $1 order by cr_dt`
	ret := []domain.Order{}

	rows, err := s.db.Query(sql1, user.Login)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		order := domain.Order{}
		err = rows.Scan(&order.ID, &order.CrDt, &order.Status)
		if err != nil {
			return nil, err
		}
		ret = append(ret, order)
	}
	return ret, nil
}

func (s *storage) UserBalanceWithdraw(ctx context.Context, user domain.User, withdraw domain.Withdraw) error {
	sql1 := `insert into withdrawals (order_id, user_id, total, cr_dt, status) values ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(sql1, withdraw.Order, user.Login, withdraw.Sum, time.Now(), "NEW")
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) CreateUser(_ context.Context, user domain.User) (int64, error) {
	// s.userMtx.Lock()
	// defer s.userMtx.Unlock()

	// id := s.currentUserID
	// s.currentUserID++
	// user.ID = id
	// if _, ok := s.users[id]; ok {
	// 	return 0, fmt.Errorf("user with id %d already exists", id)
	// }
	// s.users[id] = user
	// return id, nil
	return 0, nil
}

func (s *storage) GetUser(_ context.Context, userID int64) (domain.User, error) {
	// s.userMtx.RLock()
	// defer s.userMtx.RUnlock()

	// user, ok := s.users[userID]
	// if !ok {
	// 	return domain.User{}, fmt.Errorf("iser with id %d is not exists", userID)
	// }
	// return user, nil
	return domain.User{}, nil
}
