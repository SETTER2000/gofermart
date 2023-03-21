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
		FindByLogin(context.Context, string) (*entity.Authentication, error)
		FindByID(context.Context, string) (*entity.Authentication, error)
		Register(context.Context, *entity.Authentication) error
		Shorten(context.Context, *entity.Gofermart) (string, error)
		LongLink(context.Context, *entity.Gofermart) (string, error)
		OrderAdd(context.Context, *entity.Gofermart) (string, error)
		ShortLink(context.Context, *entity.Gofermart) (*entity.Gofermart, error)
		OrderFindByID(context.Context, *entity.Gofermart) (*entity.Gofermart, error)
		UserAllLink(ctx context.Context, u *entity.User) (*entity.User, error)
		UserDelLink(ctx context.Context, u *entity.User) error
		ReadService() error
		SaveService() error
	}

	// GofermartRepo -.
	GofermartRepo interface {
		GetByLogin(context.Context, string) (*entity.Authentication, error)
		GetByID(context.Context, string) (*entity.Authentication, error)
		Registry(context.Context, *entity.Authentication) error
		Post(context.Context, *entity.Gofermart) error
		Put(context.Context, *entity.Gofermart) error
		OrderIn(context.Context, *entity.Gofermart) error
		Get(context.Context, *entity.Gofermart) (*entity.Gofermart, error)
		OrderGetByID(context.Context, *entity.Gofermart) (*entity.Gofermart, error)
		GetAll(context.Context, *entity.User) (*entity.User, error)
		Delete(context.Context, *entity.User) error
		Read() error
		Save() error
	}
)
