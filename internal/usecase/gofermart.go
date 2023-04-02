package usecase

import (
	"context"
	"errors"
	"github.com/SETTER2000/gofermart/internal/entity"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrBadRequest    = errors.New("bad request")
)

// GofermartUseCase -.
type GofermartUseCase struct {
	repo GofermartRepo
}

// New -.
func New(r GofermartRepo) *GofermartUseCase {
	return &GofermartUseCase{
		repo: r,
	}
}

func (uc *GofermartUseCase) UserFindByLogin(ctx context.Context, s string) (*entity.Authentication, error) {
	//auth.Login = ctx.Value(auth.Cookie.AccessTokenName).(string)
	a, err := uc.repo.GetByLogin(ctx, s)
	if err != nil {
		return nil, err
	}
	return a, nil
}
func (uc *GofermartUseCase) UserFindByID(ctx context.Context, s string) (*entity.Authentication, error) {
	//auth.Login = ctx.Value(auth.Cookie.AccessTokenName).(string)
	a, err := uc.repo.GetByID(ctx, s)
	if err != nil {
		return nil, err
	}
	return a, nil
}
func (uc *GofermartUseCase) Register(ctx context.Context, auth *entity.Authentication) error {
	//auth.Login = ctx.Value(auth.Cookie.AccessTokenName).(string)
	err := uc.repo.Registry(ctx, auth)
	if err != nil {
		return err
	}
	return nil
}

// OrderAdd добавить ордер
func (uc *GofermartUseCase) OrderAdd(ctx context.Context, o *entity.Order) (*entity.Order, error) {
	err := uc.repo.OrderIn(ctx, o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// BalanceWithdraw запрос на списание средств
//func (uc *GofermartUseCase) BalanceWithdraw(ctx context.Context, wd *entity.Withdraw) error {
//	err := uc.repo.BalanceWriteOff(ctx, wd)
//	//err := uc.repo.OrderIn(ctx, o)
//	if err != nil {
//		return err
//	}
//	return nil
//}

// OrderListUserID возвращает все заказы пользователя
func (uc *GofermartUseCase) OrderListUserID(ctx context.Context, u *entity.User) (*entity.OrderList, error) {
	ol, err := uc.repo.OrderListGetUserID(ctx, u)
	if err == nil {
		return ol, nil
	}
	return nil, ErrBadRequest
}

// OrderListAll возвращает все заказы в соответствие со статусом
func (uc *GofermartUseCase) OrderListAll(ctx context.Context) (*entity.OrderList, error) {
	ol, err := uc.repo.OrderListGetStatus(ctx)
	if err == nil {
		return ol, nil
	}
	return nil, ErrBadRequest
}

// OrderFindByID поиск заказа по номеру заказа
func (uc *GofermartUseCase) OrderFindByID(ctx context.Context, o *entity.Order) (*entity.OrderResponse, error) {
	//o.UserID = ctx.Value("access_token").(string)
	or, err := uc.repo.OrderGetByNumber(ctx, o)
	if err == nil {
		return or, nil
	}
	return nil, ErrBadRequest
}

// OrderBalanceWithdrawAdd запрос на списание средств
func (uc *GofermartUseCase) OrderBalanceWithdrawAdd(ctx context.Context, wd *entity.Withdraw) error {
	wd.UserID = ctx.Value("access_token").(string)
	err := uc.repo.OrderPostBalanceWithdraw(ctx, wd)
	if err != nil {
		return err
	}
	return nil
}

// FindWithdrawalsList получение информации о выводе средств
func (uc *GofermartUseCase) FindWithdrawalsList(ctx context.Context) (*entity.WithdrawalsList, error) {
	ol, err := uc.repo.BalanceGetAll(ctx)
	if err == nil {
		return ol, nil
	}
	return nil, ErrBadRequest
}

// FindBalance получение текущего баланса пользователя
func (uc *GofermartUseCase) FindBalance(ctx context.Context) (*entity.Balance, error) {
	b, err := uc.repo.Balance(ctx)
	if err == nil {
		return b, nil
	}
	return nil, ErrBadRequest
}

// OrderUpdate обновить состояние заказа
func (uc *GofermartUseCase) OrderUpdate(ctx context.Context, ls *entity.LoyaltyStatus) error {
	err := uc.repo.UpdateOrder(ctx, ls)
	if err != nil {
		return err
	}
	return nil
}

// OrderUpdateUserID обновить состояние заказа по ID пользователя
func (uc *GofermartUseCase) OrderUpdateUserID(ctx context.Context, ls *entity.LoyaltyStatus) error {
	err := uc.repo.UpdateOrderUserID(ctx, ls)
	if err != nil {
		return err
	}
	return nil
}

func (uc *GofermartUseCase) SaveService() error {
	err := uc.repo.Save()
	if err != nil {
		return err
	}
	return nil
}
func (uc *GofermartUseCase) ReadService() error {
	err := uc.repo.Read()
	if err != nil {
		return err
	}
	return nil
}
