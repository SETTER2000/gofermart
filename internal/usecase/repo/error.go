package repo

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound                 = errors.New("not found")
	ErrAlreadyExists            = errors.New("already exists")
	ErrInsufficientFundsAccount = errors.New("insufficient funds in the account")
)

// TimeError предназначен для ошибок с фиксацией времени возникновения.
type TimeError struct {
	Time time.Time
	Err  error
}

// Error добавляет поддержку интерфейса error для типа TimeError.
func (te *TimeError) Error() string {
	return fmt.Sprintf("%v %v", te.Time.Format(`2006/01/02 15:04:05`), te.Err)
}

// NewTimeError упаковывает ошибку err в тип TimeError c текущим временем.
func NewTimeError(err error) error {
	return &TimeError{
		Time: time.Now(),
		Err:  err,
	}
}
func (te *TimeError) Unwrap() error {
	return te.Err
}

func (te *TimeError) Is(err error) bool {
	return te.Err == err
}

// LabelError описывает ошибку с дополнительной меткой.
type LabelError struct {
	Label string // метка должна быть в верхнем регистре
	Err   error
}

// NewLabelError упаковывает ошибку err в тип LabelError.
func NewLabelError(label string, err error) error {
	return &LabelError{
		Label: strings.ToUpper(label),
		Err:   err,
	}
}

// Error добавляет поддержку интерфейса error для типа LabelError.
func (le *LabelError) Error() string {
	return fmt.Sprintf("[%s] %v", le.Label, le.Err)
	//return fmt.Sprintf("[%s] %v", le.Label, le.Err)
}

func (le *LabelError) Unwrap() error {
	return le.Err
}

func (le *LabelError) Is(err error) bool {
	return le.Err == err
}

// ConflictError описывает ошибку с дополнительной меткой и значением.
type ConflictError struct {
	Label string // метка должна быть в верхнем регистре
	URL   string // уже имеющийся сокращённый URL
	Err   error
}

// NewConflictError упаковывает ошибку err в тип LabelError.
func NewConflictError(label string, url string, err error) error {
	return &ConflictError{
		Label: strings.ToUpper(label),
		URL:   url,
		Err:   err,
	}
}

// Error добавляет поддержку интерфейса error для типа LabelError.
func (ce *ConflictError) Error() string {
	return fmt.Sprintf("%v", ce.Err)
	//return fmt.Sprintf("[%s] %s %v", ce.Label, ce.URL, ce.Err)
}

func (ce *ConflictError) Unwrap() error {
	return ce.Err
}

func (ce *ConflictError) Is(err error) bool {
	return ce.Err == err
}
