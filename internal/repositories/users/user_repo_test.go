package repositoriesusers

import (
	"regexp"
	"testing"

	"github.com/Raime-34/gophermart.git/internal/dto"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserRepo_RegisterUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	userData := dto.UserData{
		Uuid:     "3f3b0e7e-5f8a-4d2a-9b66-1c8a2b6e1f7a",
		Login:    "Test login",
		Password: "Test password",
	}
	userCreds := dto.UserCredential{
		Login:    userData.Login,
		Password: userData.Password,
	}

	mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery())).
		WithArgs(
			userData.Login,
			userData.Password,
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"uuid"}).
				AddRow(userData.Uuid),
		)

	repo := NewUserRepo(t.Context(), mock)
	err = repo.RegisterUser(
		t.Context(),
		userCreds,
	)
	assert.Nil(t, err)
	gotedUserData, found := repo.getCachedUser(userCreds)
	assert.True(t, found)
	assert.NotNil(t, gotedUserData)
	assert.Equal(t, userData, *gotedUserData)
}

func TestUserRepo_GetUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(t)
	}

	userData := dto.UserData{
		Uuid:     "3f3b0e7e-5f8a-4d2a-9b66-1c8a2b6e1f7a",
		Login:    "Test login",
		Password: "Test password",
	}
	userCreds := dto.UserCredential{
		Login:    userData.Login,
		Password: userData.Password,
	}

	// Получение несуществующего юзера
	mock.ExpectQuery(regexp.QuoteMeta(getUserQuery())).
		WithArgs(userData.Login).
		WillReturnRows(
			pgxmock.NewRows([]string{"uuid", "login", "password"}),
		)

	repo := NewUserRepo(t.Context(), mock)
	gotedUserData, err := repo.GetUser(t.Context(), userCreds)
	assert.NotNil(t, err)
	assert.Nil(t, gotedUserData)

	// Получаем существующего юзера
	mock.ExpectQuery(regexp.QuoteMeta(insertUserQuery())).
		WithArgs(
			userData.Login,
			userData.Password,
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"uuid"}).
				AddRow(userData.Uuid),
		)
	repo.RegisterUser(t.Context(), userCreds)

	gotedUserData, err = repo.GetUser(t.Context(), userCreds)
	assert.Nil(t, err)
	assert.NotNil(t, gotedUserData)
	assert.Equal(t, userData, *gotedUserData)
}
