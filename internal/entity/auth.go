package entity

import (
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/scripts"
	validation "github.com/go-ozzo/ozzo-validation"
)

type Authentication struct {
	ID              string `json:"id"`
	Login           string `json:"login"  validate:"required"`
	Password        string `json:"password,omitempty"`
	EncryptPassword string `json:"-"`
	*config.Config  `json:"-"`
}

func (a *Authentication) Validate() error {
	return validation.ValidateStruct(
		a,
		validation.Field(&a.Login, validation.Required),
		validation.Field(&a.Password, validation.By(requiredIf(a.EncryptPassword == "")), validation.Length(1, 64)),
	)
	//return validation.ValidateStruct(a, validation.Field(&a.Email, validation.Required, is.Email))
}

// Sanitize очищает поля, для того чтоб они не возвращались в ответе
func (a *Authentication) Sanitize() {
	a.Password = ""
}

func (a *Authentication) BeforeCreate() error {
	enc, err := scripts.EncryptString(a.Password)
	if err != nil {
		return err
	}
	a.EncryptPassword = enc
	return nil
}

func (a *Authentication) ComparePassword(password string) bool {
	if len(password) < 1 {
		fmt.Printf("Is Empty pass: %v\n", password)
		return false
	}
	enc, err := scripts.EncryptString(password)
	if err != nil {
		return false
	}
	if enc == a.EncryptPassword {
		return true
	}
	return false
}
