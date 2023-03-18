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

// ShorturlUseCase -.
type ShorturlUseCase struct {
	repo ShorturlRepo
}

// New -.
func New(r ShorturlRepo) *ShorturlUseCase {
	return &ShorturlUseCase{
		repo: r,
	}
}

func (uc *ShorturlUseCase) Shorten(ctx context.Context, sh *entity.Gofermart) (string, error) {
	sh.UserID = ctx.Value(sh.Cookie.AccessTokenName).(string)
	err := uc.repo.Post(ctx, sh)
	if err != nil {
		return "", err
	}
	return sh.Slug, nil
}

// LongLink принимает длинный URL и возвращает короткий (PUT /api)
func (uc *ShorturlUseCase) LongLink(ctx context.Context, sh *entity.Gofermart) (string, error) {
	//sh.Slug = scripts.UniqueString()
	sh.UserID = ctx.Value("access_token").(string)
	err := uc.repo.Put(ctx, sh)
	if err != nil {
		return "", err
	}
	return sh.Slug, nil
}

// ShortLink принимает короткий URL и возвращает длинный (GET /api/{key})
func (uc *ShorturlUseCase) ShortLink(ctx context.Context, sh *entity.Gofermart) (*entity.Gofermart, error) {
	sh.UserID = ctx.Value("access_token").(string)
	sh, err := uc.repo.Get(ctx, sh)
	if err == nil {
		return sh, nil
	}
	return nil, ErrBadRequest
}

// UserAllLink принимает короткий URL и возвращает длинный (GET /user/urls)
func (uc *ShorturlUseCase) UserAllLink(ctx context.Context, u *entity.User) (*entity.User, error) {
	u, err := uc.repo.GetAll(ctx, u)
	if err == nil {
		return u, nil
	}
	return nil, ErrBadRequest
}

// UserDelLink принимает короткий URL и возвращает длинный (DELETE /api/user/urls)
func (uc *ShorturlUseCase) UserDelLink(ctx context.Context, u *entity.User) error {
	err := uc.repo.Delete(ctx, u)
	if err != nil {
		return err
	}
	return nil
}
func (uc *ShorturlUseCase) SaveService() error {
	err := uc.repo.Save()
	if err != nil {
		return err
	}
	return nil
}
func (uc *ShorturlUseCase) ReadService() error {
	err := uc.repo.Read()
	if err != nil {
		return err
	}
	return nil
}
