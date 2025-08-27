package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

type userClient interface {
	RegisterAndLoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error)
	LoginUser(ctx context.Context, reg domain.Login) (domain.SessionToken, error)
}

type loyaltyClient interface {
	AddOrder(ctx context.Context, userID domain.ID, order domain.Order) error
	GetOrders(ctx context.Context, userID domain.ID) (domain.Orders, error)
	GetBalance(ctx context.Context, userID domain.ID) (domain.Balance, error)
	WithdrawPoints(ctx context.Context, userID domain.ID, o domain.Operation) error
	GetWithdrawals(ctx context.Context, userID domain.ID) ([]domain.Operation, error)
}

type Implementation struct {
	a userClient
	l loyaltyClient
}

func New(a userClient, l loyaltyClient) *Implementation {
	return &Implementation{
		a: a,
		l: l,
	}
}

type login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register POST /api/user/register — регистрация пользователя;
func (i *Implementation) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(ctx, w, serviceerrors.NewAppError(err))
		return
	}

	l := login{}
	if err = json.Unmarshal(body, &l); err != nil {
		writeError(ctx, w, serviceerrors.NewBadRequest())
		return
	}

	authInfo, err := i.a.RegisterAndLoginUser(r.Context(), domain.Login{
		Login:    l.Login,
		Password: l.Password,
	})
	if err != nil {
		writeError(ctx, w, serviceerrors.NewAppError(err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    AuthorizationKey,
		Value:   authInfo.Token,
		Expires: time.Now().Add(24 * 180 * time.Hour),
		Secure:  true,
	})

	w.Header().Set(AuthorizationKey, authInfo.Token)

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("success register and authorize")); err != nil {
		writeError(ctx, w, serviceerrors.NewAppError(err))
	}
}

// Login POST /api/user/login — аутентификация пользователя;
func (i *Implementation) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {

		return
	}

	l := login{}
	if err = json.Unmarshal(body, &l); err != nil {
		writeError(ctx, w, serviceerrors.NewBadRequest())
		return
	}

	authInfo, err := i.a.LoginUser(ctx, domain.Login{
		Login:    l.Login,
		Password: l.Password,
	})
	if err != nil {
		writeError(ctx, w, err)

		return
	}

	// TODO почему то тесты не хотят кушать куку
	http.SetCookie(w, &http.Cookie{
		Name:    AuthorizationKey,
		Value:   authInfo.Token,
		Expires: time.Now().Add(24 * 180 * time.Hour),
		Secure:  false,
	})

	w.Header().Set(AuthorizationKey, authInfo.Token)

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("success authorize")); err != nil {
		writeError(ctx, w, serviceerrors.NewAppError(err))
	}
}

// AddOrder POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
func (i *Implementation) AddOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	if isValidID := checkLuhnAlgorithm(string(body)); !isValidID {
		writeError(ctx, w,
			serviceerrors.NewUnprocessableEntity().Wrap(nil, "invalid order id"))
		return
	}

	userID, isOK := getUserID(ctx)
	if !isOK {
		writeError(ctx, w,
			serviceerrors.NewAppError(nil).
				Wrap(errors.New("invalid user id type assertion"), ""))
		return
	}

	orderID, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		writeError(ctx, w,
			serviceerrors.NewAppError(nil).Wrap(err, ""))
		return
	}

	domainOrder := domain.Order{
		ID: domain.ID{
			ID: orderID,
		},
	}

	if err = i.l.AddOrder(ctx, userID, domainOrder); err != nil {
		switch {
		case errors.Is(err, domain.ErrActionCompletedEarly):
			w.WriteHeader(http.StatusOK)
			return
		}
		writeError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

// GetOrders GET /api/user/orders — получение списка загруженных пользователем номеров заказов,
// статусов их обработки и информации о начислениях;
// TODO пагинация
func (i *Implementation) GetOrders(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	userID, isOK := getUserID(ctx)
	if !isOK {
		writeError(ctx, w,
			serviceerrors.NewAppError(nil).
				Wrap(errors.New("invalid user id type assertion"), ""))
		return
	}

	resOrders, err := i.l.GetOrders(ctx, userID)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	if resOrders == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	transportOrders := make([]order, 0, len(resOrders))
	for _, ro := range resOrders {
		val, _ := ro.AccrualAmount.Amount.Float64()

		transportOrders = append(transportOrders, order{
			Number:     strconv.FormatUint(ro.ID.ID, 10),
			Status:     string(ro.State),
			Accrual:    val,
			UploadedAt: ro.CreatedAt.Format(time.RFC3339),
		})
	}

	jsonBytes, err := json.Marshal(transportOrders)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	w.Header().Set(ContentType, ApplicationJSONType)
	_, err = w.Write(jsonBytes)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// GetBalance GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
func (i *Implementation) GetBalance(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	userID, isOK := getUserID(ctx)
	if !isOK {
		writeError(ctx, w,
			serviceerrors.NewAppError(nil).
				Wrap(errors.New("invalid user id type assertion"), ""))
		return
	}

	res, err := i.l.GetBalance(ctx, userID)
	if err != nil {
		writeError(ctx, w, err)
		return
	}
	current, _ := res.GetCurrent().Amount.Float64()
	withdrawn, _ := res.Withdrawn.Amount.Float64()
	jsonBytes, err := json.Marshal(balance{
		Current:   current,
		Withdrawn: withdrawn,
	})
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	w.Header().Set(ContentType, ApplicationJSONType)
	_, err = w.Write(jsonBytes)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type withdrawRequest struct {
	Order string `json:"order"`
	// TODO костыль полечить
	Sum float64 `json:"sum"`
}

// WithdrawalPoints POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (i *Implementation) WithdrawalPoints(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	userID, isOK := getUserID(ctx)
	if !isOK {
		writeError(ctx, w,
			serviceerrors.NewAppError(nil).
				Wrap(errors.New("invalid user id type assertion"), ""))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	req := new(withdrawRequest)
	if err = json.Unmarshal(body, req); err != nil {
		appErr := serviceerrors.NewBadRequest().LogServerError(r.Context())
		http.Error(w, appErr.String(), appErr.Code)
		return
	}

	if isValidID := checkLuhnAlgorithm(req.Order); !isValidID {
		writeError(ctx, w,
			serviceerrors.NewUnprocessableEntity().Wrap(nil, "invalid order id"))
		return
	}

	orderID, err := strconv.ParseUint(req.Order, 10, 64)
	if err != nil {
		writeError(ctx, w,
			serviceerrors.NewAppError(nil).Wrap(err, ""))
		return
	}

	if err = i.l.WithdrawPoints(
		ctx,
		userID,
		domain.Operation{
			OrderID: domain.ID{
				ID: orderID,
			},
			Amount: domain.Money{
				Currency: string(domain.GopherMarketBonuses),
				Amount:   decimal.NewFromFloat(req.Sum),
			},
		},
	); err != nil {
		writeError(ctx, w, err)
		return
	}
}

type withdrawOperation struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// GetWithdrawals GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
// TODO пагинация
func (i *Implementation) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	userID, isOK := getUserID(ctx)
	if !isOK {
		writeError(ctx, w,
			serviceerrors.NewAppError(nil).
				Wrap(errors.New("invalid user id type assertion"), ""))
		return
	}

	res, err := i.l.GetWithdrawals(ctx, userID)
	if err != nil {
		writeError(ctx, w, err)
		return
	}

	jsonResult := make([]withdrawOperation, 0, len(res))
	for _, ro := range res {
		amount, _ := ro.Amount.Amount.Float64()
		jsonResult = append(jsonResult, withdrawOperation{
			Order:       strconv.FormatUint(ro.OrderID.ID, 10),
			Sum:         amount,
			ProcessedAt: ro.CratedAt.Format(time.RFC3339),
		})
	}

	body, err := json.Marshal(jsonResult)
	if err != nil {
		writeError(ctx, w, serviceerrors.NewBadRequest())
		return
	}

	w.Header().Set(ContentType, ApplicationJSONType)
	_, err = w.Write(body)
	if err != nil {
		writeError(ctx, w, serviceerrors.NewAppError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	appErr := serviceerrors.AppErrorFromError(err).LogServerError(ctx)
	http.Error(w, appErr.String(), appErr.Code)
}

func checkLuhnAlgorithm(number string) bool {
	sum := 0
	alternate := false
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false // Некорректный символ в номере
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}
	return sum%10 == 0
}
