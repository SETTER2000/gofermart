package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/entity"
	"github.com/SETTER2000/gofermart/internal/usecase"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AClient struct {
	ctx    context.Context
	client *http.Client
	cfg    *config.Config
	url    string
	s      usecase.Gofermart
}

// NewAClient - accrual client
func NewAClient(s usecase.Gofermart, cfg *config.Config) *AClient {
	return &AClient{
		ctx:    context.Background(),
		client: &http.Client{},
		cfg:    cfg,
		url:    cfg.HTTP.Accrual,
		s:      s,
	}
}

func (a *AClient) accrualLink(order string) string {
	return fmt.Sprintf("%s/api/orders/%v", a.url, order)
}

// LoyaltyFind запрос баллов лояльности заказа
func (a *AClient) LoyaltyFind(order string) (*entity.LoyaltyStatus, error) {
	order = strings.TrimSpace(order)
	if order == "" {
		return nil, fmt.Errorf("error empty arg link")
	}
	link := a.accrualLink(order)
	req, err := http.NewRequestWithContext(a.ctx, "GET", link, nil)
	if err != nil {
		return nil, fmt.Errorf("GET: шибка подключения к клиенту Accrual:: %e", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ls := entity.LoyaltyStatus{}
	ls.Order = order
	json.Unmarshal(body, &ls)
	if resp.StatusCode == 204 && ls.Status == "" {
		ls.Status = "NEW"
		ls.Accrual = 0
		return &ls, nil
	}

	return &ls, nil
}

func (a *AClient) Start() error {
	ctx := context.Background()
	// выбрать все заказы кроме закрытых (PROCESSED, INVALID)
	ol, err := a.s.OrderListAll(ctx)
	if err != nil {
		return err
	}

	var l entity.LoyaltyStatus

	for _, o := range *ol {
		time.Sleep(1 * time.Second)
		go func(o entity.OrderResponse) {
			l.Status = o.Status
			l.Accrual = o.Accrual
			l.Order = o.Number
			lst := *a.Run(l)
			a.s.OrderUpdate(ctx, &lst) // обновить заказ
		}(o)
	}
	return nil
}
func (a *AClient) Run(lst entity.LoyaltyStatus) *entity.LoyaltyStatus {
	var ls entity.LoyaltyStatus
	b := make(chan entity.LoyaltyStatus, 1)

	go func() {
		for {
			time.Sleep(time.Second)
			r, err := a.LoyaltyFind(lst.Order)
			if err != nil {
				break
			}

			if r.Status == "PROCESSED" || r.Status == "INVALID" {
				b <- *r
				break
			}
			b <- *r
		}
		close(b)
	}()

	for c := range b {
		ls.Order = c.Order
		ls.Accrual = c.Accrual
		ls.Status = c.Status
		//fmt.Printf(":> %v\n", c)
	}

	return &ls
}
