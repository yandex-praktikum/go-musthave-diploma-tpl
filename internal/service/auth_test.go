package service

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	repo_mocks "github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository/mocks"
)

func TestAuthService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repo_mocks.NewMockAutorisation(ctrl)
	type args struct {
		user models.User
	}
	tests := []struct {
		name    string
		a       *AuthService
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "succesful creation user",
			a:    NewAuthStorage(repo),
			args: args{
				user: models.User{
					Login:    "user1",
					Password: "123",
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				repo.EXPECT().CreateUser(gomock.Any()).Return(1, nil)
			}
			got, err := tt.a.CreateUser(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthService.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AuthService.CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repo_mocks.NewMockAutorisation(ctrl)
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		a       *AuthService
		args    args
		user    models.User
		want    string
		wantErr bool
		err     error
	}{
		{
			name: "succesful creation user",
			a:    NewAuthStorage(repo),
			args: args{
				username: "user1",
				password: "123",
			},
			user: models.User{
				Login:    "user1",
				Password: generatePasswordHash("123", "SomeSalt"),
				Salt:     "SomeSalt",
			},
		},
		{

			name: "incorrect password",
			a:    NewAuthStorage(repo),
			args: args{
				username: "user1",
				password: "1234",
			},
			user: models.User{
				Login:    "user1",
				Password: generatePasswordHash("123", "SomeSalt"),
			},
			wantErr: true,
			err:     errors.New("unauthorized"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.EXPECT().GetUser(gomock.Any()).Return(tt.user, nil)
			got, err := tt.a.GenerateToken(tt.args.username, tt.args.password)

			if (err != nil) != tt.wantErr {
				require.Equal(t, err, tt.err)
				return
			}
			if got != tt.want {
				require.NotEqual(t, got, "")
			}
		})
	}
}
