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
		UserFindByLogin(context.Context, string) (*entity.Authentication, error)
		UserFindByID(context.Context, string) (*entity.Authentication, error)
		OrderAdd(context.Context, *entity.Order) (*entity.Order, error)
		BalanceWithdraw(context.Context, *entity.Withdraw) error
		OrderFindByID(context.Context, *entity.Order) (*entity.Order, error)
		OrderList(ctx context.Context, u *entity.User) (*entity.OrderList, error)
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
		GetByLogin(context.Context, string) (*entity.Authentication, error)
		GetByID(context.Context, string) (*entity.Authentication, error)
		OrderGetByNumber(context.Context, *entity.Order) (*entity.Order, error)
		OrderIn(context.Context, *entity.Order) error
		BalanceWriteOff(context.Context, *entity.Withdraw) error
		OrderGetAll(context.Context, *entity.User) (*entity.OrderList, error)
		Post(context.Context, *entity.Gofermart) error
		Put(context.Context, *entity.Gofermart) error
		Get(context.Context, *entity.Gofermart) (*entity.Gofermart, error)
		GetAll(context.Context, *entity.User) (*entity.User, error)
		Delete(context.Context, *entity.User) error
		Read() error
		Save() error
	}
)
