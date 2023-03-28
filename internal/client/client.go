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
	//fmt.Printf("CONNECT ACCRUAL status: %d  %s\n", resp.StatusCode, link)
	json.Unmarshal(body, &ls)
	if resp.StatusCode == 204 && ls.Status == "" {
		ls.Status = "NEW"
		//fmt.Printf("LoyaltyPoints::%v\n", ls)
		return &ls, nil
	}

	return &ls, nil
}

func (a *AClient) Start() error {
	ctx := context.Background()

	//b := make(chan entity.LoyaltyStatus)
	//a.s.OrderListUserID(ctx, &u)

	ol, err := a.s.OrderListAll(ctx)
	if err != nil {
		return err
	}
	var l entity.LoyaltyStatus
	//var ls []entity.LoyaltyStatus
	lCh := make(chan entity.LoyaltyStatus, 1)
	for i, o := range *ol {
		time.Sleep(1 * time.Second)
		fmt.Printf("Index: %d %v\n", i, o)
		go func(o entity.OrderResponse) {
			l.Status = o.Status
			l.Accrual = o.Accrual
			l.Order = o.Number
			fmt.Printf("ORDER:: %v\n", l)
			lCh <- l
			l = *a.Run(lCh)
			a.s.OrderUpdate(ctx, &l)
		}(o)
	}
	return nil
}
func (a *AClient) Run(lCh chan entity.LoyaltyStatus) *entity.LoyaltyStatus {
	fmt.Printf("Index: %v\n", "lCh**DDDDD")
	//fmt.Printf("Index: %v\n", <-lCh)
	//wg := sync.WaitGroup{}
	//var ar []entity.LoyaltyStatus
	var ls entity.LoyaltyStatus
	//sl := []string{
	//	"101725",
	//	"7733868",
	//	"48200117"}
	//for _, j := range sl {
	//ls.Order = order
	//ls.Status = "NEW"
	//ar = append(ar, order)
	//}

	//a := make(chan int, 3)
	b := make(chan entity.LoyaltyStatus, 1)
	//a <- 1
	//a <- 2
	//a <- 3
	//for _, l := range orders {

	go func() {
		s := <-lCh
		for {
			//wg.Add(1)
			//for i := 0; i < len(ar); i++ {
			time.Sleep(time.Second)
			r, err := a.LoyaltyFind(s.Order)
			if err != nil {
				fmt.Printf("ERROR to loop:: %e", err)
				break
			}

			if r.Status == "PROCESSED" || r.Status == "INVALID" {
				b <- *r
				fmt.Printf("NEW DATA:: %v\n", *r)
				//TODO Update тут

				break
			}
			//fmt.Printf("Loop\n")
			b <- *r
		}
		close(b)
		//wg.Done()
	}()
	//}(l)
	//}
	//close(a)
	//cCh := gc.Union2(a, b)
	//
	for c := range b {

		ls.Order = c.Order
		ls.Accrual = c.Accrual
		ls.Status = c.Status
		fmt.Printf(":> %v\n", c)
	}
	//wg.Wait()
	//for c := range cCh {
	//	fmt.Printf("Loyalty_XXX: %v\n", c)
	//}
	fmt.Printf("EXIT OPROS:: %v %v %v\n", ls.Order, ls.Accrual, ls.Status)
	return &ls
}
