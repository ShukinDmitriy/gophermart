package models

import (
	"testing"
	"time"

	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMapUserToUserLoginResponse(t *testing.T) {
	type args struct {
		user *entities.User
	}
	tests := []struct {
		name string
		args args
		want UserLoginResponse
	}{
		{
			name: "Convert user",
			args: args{
				user: &entities.User{
					Model: gorm.Model{
						ID:        123,
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
						DeletedAt: gorm.DeletedAt{
							Time:  time.Time{},
							Valid: false,
						},
					},
					LastName:   "LastName",
					FirstName:  "FirstName",
					MiddleName: "MiddleName",
					Login:      "Login",
					Password:   "Password",
					Email:      "Email",
				},
			},
			want: UserLoginResponse{
				LastName:   "LastName",
				FirstName:  "FirstName",
				MiddleName: "MiddleName",
				Login:      "Login",
				Email:      "Email",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MapUserToUserLoginResponse(tt.args.user), "MapUserToUserLoginResponse(%v)", tt.args.user)
		})
	}
}
