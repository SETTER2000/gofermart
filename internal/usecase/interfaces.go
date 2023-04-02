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
		OrderBalanceWithdrawAdd(context.Context, *entity.Withdraw) error
		OrderFindByID(context.Context, *entity.Order) (*entity.OrderResponse, error)
		OrderListUserID(ctx context.Context, u *entity.User) (*entity.OrderList, error)
		OrderListAll(ctx context.Context) (*entity.OrderList, error)
		FindWithdrawalsList(ctx context.Context) (*entity.WithdrawalsList, error)
		FindBalance(ctx context.Context) (*entity.Balance, error)
		OrderUpdate(ctx context.Context, ls *entity.LoyaltyStatus) error
		OrderUpdateUserID(ctx context.Context, ls *entity.LoyaltyStatus) error
		ReadService() error
		SaveService() error
	}

	// GofermartRepo -.
	GofermartRepo interface {
		Registry(context.Context, *entity.Authentication) error
		GetByLogin(context.Context, string) (*entity.Authentication, error)
		GetByID(context.Context, string) (*entity.Authentication, error)
		OrderGetByNumber(context.Context, *entity.Order) (*entity.OrderResponse, error)
		OrderIn(context.Context, *entity.Order) error
		OrderPostBalanceWithdraw(context.Context, *entity.Withdraw) error
		OrderListGetUserID(context.Context, *entity.User) (*entity.OrderList, error)
		OrderListGetStatus(context.Context) (*entity.OrderList, error)
		BalanceGetAll(context.Context) (*entity.WithdrawalsList, error)
		Balance(context.Context) (*entity.Balance, error)
		UpdateOrder(ctx context.Context, ls *entity.LoyaltyStatus) error
		UpdateOrderUserID(ctx context.Context, ls *entity.LoyaltyStatus) error
		Read() error
		Save() error
	}
)
