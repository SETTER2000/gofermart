package entity

import (
	"github.com/SETTER2000/gofermart/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthentication_BeforeCreate(t *testing.T) {
	type fields struct {
		Login           string
		Password        string
		EncryptPassword string
		Config          *config.Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "positive test #1",
			fields: fields{
				Login:    "bob",
				Password: "123",
			},
			wantErr: nil,
		},
		//{
		//	name: "negative test #2",
		//	fields: fields{
		//		Login:    "bob",
		//		Password: "",
		//	},
		//	wantErr: ErrBadRequest,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authentication{
				Login:           tt.fields.Login,
				Password:        tt.fields.Password,
				EncryptPassword: tt.fields.EncryptPassword,
				Config:          tt.fields.Config,
			}

			assert.NoError(t, a.BeforeCreate())
			assert.NotEmpty(t, a.EncryptPassword)
		})
	}
}

func TestAuthentication_Validate(t *testing.T) {
	type fields struct {
		Login           string
		Password        string
		EncryptPassword string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				Login:           "aa",
				Password:        "aa",
				EncryptPassword: "",
			},
			wantErr: true,
		}, {
			name: "negative test #1 empty email",
			fields: fields{
				Login:           "",
				Password:        "aa",
				EncryptPassword: "",
			},
			wantErr: false,
		}, {
			name: "negative test #2 empty password",
			fields: fields{
				Login:           "qw",
				Password:        "",
				EncryptPassword: "",
			},
			wantErr: false,
		}, {
			name: "with encrypted password",
			fields: fields{
				Login:           "qw",
				Password:        "",
				EncryptPassword: "encryptedpassword",
			},
			wantErr: true,
		}, {
			name: "positive test #2 max length password",
			fields: fields{
				Login:           "qw",
				Password:        "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
				EncryptPassword: "",
			},
			wantErr: true,
		}, {
			name: "negative test #3 max length password",
			fields: fields{
				Login:           "qw",
				Password:        "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae32",
				EncryptPassword: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authentication{
				Login:           tt.fields.Login,
				Password:        tt.fields.Password,
				EncryptPassword: tt.fields.EncryptPassword,
			}

			if tt.wantErr {
				assert.NoError(t, a.Validate())
			} else {
				assert.Error(t, a.Validate())
			}
		})
	}
}

func TestAuthentication_ComparePassword(t *testing.T) {
	type fields struct {
		ID              string
		Login           string
		Password        string
		EncryptPassword string
		Config          *config.Config
	}
	type args struct {
		password string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				EncryptPassword: "6c99f25e5ef66c3ecb370f8759f17597a3b4d4c81acc1e42c8228bd67d823339",
			},
			args: args{
				password: "123",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authentication{
				ID:              tt.fields.ID,
				Login:           tt.fields.Login,
				Password:        tt.fields.Password,
				EncryptPassword: tt.fields.EncryptPassword,
				Config:          tt.fields.Config,
			}
			assert.Equalf(t, tt.want, a.ComparePassword(tt.args.password), "ComparePassword(%v)", tt.args.password)
		})
	}
}
