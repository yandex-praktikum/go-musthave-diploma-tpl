package pgx

import (
	"context"
	"errors"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (st *storage) RegisterUser(ctx context.Context, ld *domain.LoginData) (int, error) {
	st.logger.Infow("storage.RegisterUser", "status", "start")

	var userID int
	if err := st.pPool.QueryRow(ctx,
		"insert into userInfo(login, hash, salt) values ($1, $2, $3) returning userId",
		ld.Login,
		ld.Hash,
		ld.Salt).Scan(&userID); err == nil {
		st.logger.Infow("storage.RegisterUser", "status", "success", "userID", userID)
		return userID, nil
	} else {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				st.logger.Infow("storage.RegisterUser", "err", "login is busy")
				return -1, domain.ErrLoginIsBusy
			}
		}
		st.logger.Infow("storage.RegisterUser", "err", err.Error())
		return -1, domain.ErrServerInternal
	}
}

func (st *storage) GetUserData(ctx context.Context, login string) (*domain.LoginData, error) {
	st.logger.Infow("storage.GetUserData", "status", "start")

	var data domain.LoginData
	err := st.pPool.QueryRow(ctx, "select userId, login, hash, salt from userInfo where login = $1", login).Scan(&data.UserID, &data.Login, &data.Hash, &data.Salt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			st.logger.Infow("storage.GetUserData", "status", "not found", "login", login)
			return nil, nil
		}
		st.logger.Errorw("storage.GetUserData", "err", err.Error())
		return nil, domain.ErrServerInternal
	}

	return &data, nil
}
