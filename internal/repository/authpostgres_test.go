package repository

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository/mocks"
)

func randFunc(n int) string {
	var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, n)
	for i := range b {
		// randomly select 1 character from given charset
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func TestAuthPostgres_CreateUser(t *testing.T) {
	type args struct {
		user models.User
	}
	tests := []struct {
		name    string
		r       *AuthPostgres
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "success_create_user",
			args: args{
				models.User{
					Login:    "testuser",
					Password: "123456",
					Salt:     randFunc(20),
				},
			},
			want:    1,
			wantErr: false,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthStorage := mocks.NewMockAutorisation(ctrl)
	mockAuthStorage.EXPECT().CreateUser(gomock.Any()).Return(int64(-1), nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockAutorisation{}
			if !tt.wantErr {

				repo.EXPECT().Return(tt.want, nil)
			} else {
				repo.EXPECT("CreateUser", tt.args.user).Return(tt.want, errors.New("Failed to create user"))
			}
			got, err := repo.CreatePayment(tt.args.payment)
			// if (err != nil) != tt.wantErr {
			//     t.Errorf("Repository.CreatePayment() error = %v, wantErr %v", err, tt.wantErr)
			//     return
			// }
			// if got != tt.want {
			//     t.Errorf("Repository.CreatePayment() = %v, want %v", got, tt.want)
			// }
			// got, err := tt.r.CreateUser(tt.args.user)
			// fmt.Println("11111111", got)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("AuthPostgres.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }
			// if got != tt.want {
			// 	t.Errorf("AuthPostgres.CreateUser() = %v, want %v", got, tt.want)
			// }
		})
	}
}
