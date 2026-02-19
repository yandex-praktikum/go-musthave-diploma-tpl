package gophermart

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Raime-34/gophermart.git/internal/consts"
	"github.com/Raime-34/gophermart.git/internal/dto"
	mock "github.com/Raime-34/gophermart.git/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGophermart_RegisterUser_UserAlreadyExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	calc := mock.NewMockaccrualCalculator(ctrl)

	g := &Gophermart{repositories: repos, accrualCalculator: calc}
	ctx := context.Background()
	creds := dto.UserCredential{Login: "u", Password: "p"}

	repos.EXPECT().GetUser(ctx, creds).Return(&dto.UserData{}, nil)

	err := g.RegisterUser(ctx, creds)
	require.ErrorIs(t, err, ErrUserAlreadyExist)
}

func TestGophermart_RegisterUser_RegisterCalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	calc := mock.NewMockaccrualCalculator(ctrl)

	g := &Gophermart{repositories: repos, accrualCalculator: calc}
	ctx := context.Background()
	creds := dto.UserCredential{Login: "u", Password: "p"}

	repos.EXPECT().GetUser(ctx, creds).Return(nil, errors.New("not found"))
	repos.EXPECT().RegisterUser(ctx, creds).Return(nil)

	require.NoError(t, g.RegisterUser(ctx, creds))
}

func TestGophermart_LoginUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	creds := dto.UserCredential{Login: "u", Password: "p"}

	repos.EXPECT().GetUser(ctx, creds).Return(nil, errors.New("no rows"))

	ud, err := g.LoginUser(ctx, creds)
	require.Nil(t, ud)
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestGophermart_LoginUser_IncorrectPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	creds := dto.UserCredential{Login: "u", Password: "p"}

	repos.EXPECT().GetUser(ctx, creds).Return(&dto.UserData{Login: "u", Password: "other"}, nil)

	ud, err := g.LoginUser(ctx, creds)
	require.Nil(t, ud)
	require.ErrorIs(t, err, ErrIncorrectPassword)
}

func TestGophermart_LoginUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	creds := dto.UserCredential{Login: "u", Password: "p"}
	user := &dto.UserData{Login: "u", Password: "p"}

	repos.EXPECT().GetUser(ctx, creds).Return(user, nil)

	got, err := g.LoginUser(ctx, creds)
	require.NoError(t, err)
	require.Equal(t, user, got)
}

func TestGophermart_InsertOrder_UserIdString(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	calc := mock.NewMockaccrualCalculator(ctrl)
	g := &Gophermart{repositories: repos, accrualCalculator: calc}

	ctx := context.WithValue(context.Background(), consts.UserIdKey, "user-1")
	order := "123"

	calc.EXPECT().AddToMonitoring(order, "user-1")
	repos.EXPECT().RegisterOrder(ctx, order).Return(nil)

	require.NoError(t, g.InsertOrder(ctx, order))
}

func TestGophermart_InsertOrder_UserIdUUID_CurrentCodePassesEmptyString(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	calc := mock.NewMockaccrualCalculator(ctrl)
	g := &Gophermart{repositories: repos, accrualCalculator: calc}

	u := uuid.New()
	ctx := context.WithValue(context.Background(), consts.UserIdKey, u)
	order := "123"

	// из-за userIdStr, _ := userId.(string) -> будет ""
	calc.EXPECT().AddToMonitoring(order, "")
	repos.EXPECT().RegisterOrder(ctx, order).Return(nil)

	require.NoError(t, g.InsertOrder(ctx, order))
}

func TestGophermart_InsertOrder_InvalidUserIdType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	calc := mock.NewMockaccrualCalculator(ctrl)
	g := &Gophermart{repositories: repos, accrualCalculator: calc}

	ctx := context.WithValue(context.Background(), consts.UserIdKey, 123) // invalid
	err := g.InsertOrder(ctx, "123")
	require.Error(t, err)
}

func TestGophermart_GetUserOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	orders := []*dto.OrderInfo{
		{Number: "1", Status: "NEW", Accrual: 10},
		{Number: "2", Status: "PROCESSED", Accrual: 20},
	}

	repos.EXPECT().GetOrders(ctx).Return(orders, nil)

	resp, err := g.GetUserOrders(ctx)
	require.NoError(t, err)
	require.Len(t, resp, 2)
	require.Equal(t, "1", resp[0].Number)
	require.Equal(t, "2", resp[1].Number)
}

func TestGophermart_GetUserBalance_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()

	repos.EXPECT().GetOrders(ctx).Return([]*dto.OrderInfo{
		{Accrual: 100},
		{Accrual: 50},
	}, nil)

	repos.EXPECT().GetWithdraws(ctx).Return([]*dto.WithdrawInfo{
		{Sum: 30},
	}, nil)

	bal, err := g.GetUserBalance(ctx)
	require.NoError(t, err)
	require.Equal(t, float64(120), bal.Current)
	require.Equal(t, 30, bal.Withdraw)
}

func TestGophermart_GetUserBalance_GetOrdersErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	repos.EXPECT().GetOrders(ctx).Return(nil, errors.New("db"))

	bal, err := g.GetUserBalance(ctx)
	require.Nil(t, bal)
	require.Error(t, err)
}

func TestGophermart_ProcessWithdraw_NotEnoughBonuses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	req := dto.WithdrawRequest{Order: "1", Sum: 10}

	repos.EXPECT().GetOrders(ctx).Return([]*dto.OrderInfo{
		{Accrual: 50},
	}, nil)
	repos.EXPECT().GetWithdraws(ctx).Return([]*dto.WithdrawInfo{
		{Sum: 100}, // withdrawals > bonuses => ErrNotEnoughBonuses (как в коде)
	}, nil)

	err := g.ProcessWithdraw(ctx, req)
	require.ErrorIs(t, err, ErrNotEnoughBonuses)
}

func TestGophermart_ProcessWithdraw_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	req := dto.WithdrawRequest{Order: "1", Sum: 10}

	repos.EXPECT().GetOrders(ctx).Return([]*dto.OrderInfo{
		{Accrual: 100},
	}, nil)
	repos.EXPECT().GetWithdraws(ctx).Return([]*dto.WithdrawInfo{
		{Sum: 20},
	}, nil)
	repos.EXPECT().RegisterWithdraw(ctx, req).Return(nil)

	require.NoError(t, g.ProcessWithdraw(ctx, req))
}

func TestGophermart_GetWithdraws(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ctx := context.Background()
	ws := []*dto.WithdrawInfo{{Order: "1", Sum: 10}}

	repos.EXPECT().GetWithdraws(ctx).Return(ws, nil)

	got, err := g.GetWithdraws(ctx)
	require.NoError(t, err)
	require.Equal(t, ws, got)
}

func TestGophermart_handleOrderState_CallsUpdateOrderAndSetsUserIdInCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := mock.NewMockrepositoriesInt(ctrl)
	g := &Gophermart{repositories: repos}

	ch := make(chan *dto.AccrualCalculatorDTO, 1)
	state := &dto.AccrualCalculatorDTO{Order: "1", Status: "PROCESSED", Accrual: 10}
	state.AddUserId("user-1")

	repos.EXPECT().
		UpdateOrder(gomock.Any(), *state).
		DoAndReturn(func(ctx context.Context, _ dto.AccrualCalculatorDTO) error {
			require.Equal(t, "user-1", ctx.Value(consts.UserIdKey))
			return nil
		})

	ch <- state
	close(ch)

	var wg sync.WaitGroup
	g.handleOrderState(t.Context(), ch, &wg)
}
