package service

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/models"
	repo_mocks "github.com/tanya-mtv/go-musthave-diploma-tpl.git/internal/repository/mocks"
)

func TestAccountService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repo_mocks.NewMockBalance(ctrl)

	type args struct {
		userID   int
		withdraw models.Withdraw
	}
	tests := []struct {
		name    string
		b       *AccountService
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			b:    NewAccountService(repo),
			args: args{
				userID: 1,
				withdraw: models.Withdraw{
					Order: "371449635398431",
					Sum:   100,
				},
			},
		},
		{
			name: "PreconditionFailed",
			b:    NewAccountService(repo),
			args: args{
				userID: 1,
				withdraw: models.Withdraw{
					Order: "371449635398431a",
					Sum:   100,
				},
			},
			wantErr: true,
		},
		{
			name: "luhns check",
			b:    NewAccountService(repo),
			args: args{
				userID: 1,
				withdraw: models.Withdraw{
					Order: "3714496353984315",
					Sum:   100,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				repo.EXPECT().DoWithdraw(gomock.Any(), gomock.Any()).Return(nil)
			}

			if err := tt.b.Withdraw(tt.args.userID, tt.args.withdraw); (err != nil) != tt.wantErr {
				t.Errorf("AccountService.Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
