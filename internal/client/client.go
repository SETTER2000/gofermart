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
	req, _ := http.NewRequestWithContext(a.ctx, "GET", link, nil)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к клиенту Accrual:: %e", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
	lCh := make(chan entity.LoyaltyStatus, 1)

	for _, o := range *ol {
		time.Sleep(1 * time.Second)
		go func(o entity.OrderResponse) {
			l.Status = o.Status
			l.Accrual = o.Accrual
			l.Order = o.Number
			lCh <- l
			l = *a.Run(lCh)
			a.s.OrderUpdate(ctx, &l) // обновить заказ
		}(o)
	}
	return nil
}
func (a *AClient) Run(lCh chan entity.LoyaltyStatus) *entity.LoyaltyStatus {
	var ls entity.LoyaltyStatus
	b := make(chan entity.LoyaltyStatus, 1)

	go func() {
		s := <-lCh
		for {
			time.Sleep(time.Second)
			r, err := a.LoyaltyFind(s.Order)
			if err != nil {
				fmt.Printf("ERROR to loop:: %e", err)
				break
			}

			if r.Status == "PROCESSED" || r.Status == "INVALID" {
				b <- *r
				fmt.Printf("NEW DATA:: %v\n", *r)
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
		fmt.Printf(":> %v\n", c)
	}

	fmt.Printf("EXIT OPROS:: %v %v %v\n", ls.Order, ls.Accrual, ls.Status)
	return &ls
}
