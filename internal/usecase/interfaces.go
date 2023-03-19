// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"github.com/SETTER2000/gofermart/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go

type (
	// Gofermart -.
	Gofermart interface {
		Register(context.Context, *entity.Authentication) error
		Shorten(context.Context, *entity.Gofermart) (string, error)
		LongLink(context.Context, *entity.Gofermart) (string, error)
		ShortLink(context.Context, *entity.Gofermart) (*entity.Gofermart, error)
		UserAllLink(ctx context.Context, u *entity.User) (*entity.User, error)
		UserDelLink(ctx context.Context, u *entity.User) error
		ReadService() error
		SaveService() error
	}

	// GofermartRepo -.
	GofermartRepo interface {
		Registry(context.Context, *entity.Authentication) error
		Post(context.Context, *entity.Gofermart) error
		Put(context.Context, *entity.Gofermart) error
		Get(context.Context, *entity.Gofermart) (*entity.Gofermart, error)
		GetAll(context.Context, *entity.User) (*entity.User, error)
		Delete(context.Context, *entity.User) error
		Read() error
		Save() error
	}
)
