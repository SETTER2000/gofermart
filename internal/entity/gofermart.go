package entity

import (
	"github.com/SETTER2000/gofermart/config"
)

// CorrelationOrigin -.
type CorrelationOrigin []Batch
type Response []GoferResponse
type Gofermarts []Gofermart

// Gofermart -.
type Gofermart struct {
	Slug           string `json:"slug,omitempty" example:"1674872720465761244B_5"`             // Строковый идентификатор
	URL            string `json:"url,omitempty" example:"https://example.com/go/to/home.html"` // URL для сокращения
	UserID         string `json:"user_id,omitempty"`
	Order          string `json:"order,omitempty"`
	Del            bool   `json:"del"`
	*config.Config `json:"-"`
	//*CorrelationOrigin `json:"correlation_origin,omitempty"`
}
type List struct {
	Slug string `json:"short_url" example:"1674872720465761244B_5"`                 // Строковый идентификатор
	URL  string `json:"original_url" example:"https://example.com/go/to/home.html"` // URL для сокращения
}
type OrderList []OrderResponse

type Order struct {
	Number         int     `json:"number,omitempty"`
	Status         string  `json:"status,omitempty"`
	Accrual        float32 `json:"accrual,omitempty"`
	UploadedAt     string  `json:"uploaded_at"`
	UserID         string  `json:"user_id,omitempty"`
	*config.Config `json:"-"`
}
type OrderResponse struct {
	Number         string  `json:"number,omitempty"`
	Status         string  `json:"status,omitempty"`
	Accrual        float32 `json:"accrual,omitempty"`
	UploadedAt     string  `json:"uploaded_at"`
	UserID         string  `json:"user_id,omitempty"`
	*config.Config `json:"-"`
}

type WithdrawalsList []WithdrawResponse

type Withdraw struct {
	NumOrder    string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
	*Order      `json:"-"`
}

type LoyaltyStatus struct {
	Order   string  `json:"order,omitempty"`
	Status  string  `json:"status,omitempty"`
	Accrual float32 `json:"accrual,omitempty"`
}

type WithdrawResponse struct {
	NumOrder    string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
	*Order      `json:"-"`
}
type User struct {
	UserID  string `json:"user_id" example:"1674872720465761244B_5"`
	Urls    []List
	DelLink []string
}

type Balance struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdraw"`
}

// GofermartResponse  --
type GofermartResponse struct {
	URL string `json:"result"` // URL для сокращения
}

type Batch struct {
	Slug string `json:"correlation_id" example:"1674872720465761244B_5"`            // Строковый идентификатор
	URL  string `json:"original_url" example:"https://example.com/go/to/home.html"` // URL для сокращения
}

type GoferResponse struct {
	Slug string `json:"correlation_id" example:"1674872720465761244B_5"`        // Строковый идентификатор
	URL  string `json:"short_url" example:"https://example.com/correlation_id"` // URL для сокращения
}

type Short interface{}
