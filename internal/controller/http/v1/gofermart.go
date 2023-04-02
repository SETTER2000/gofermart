package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/client"
	"github.com/SETTER2000/gofermart/internal/entity"
	"github.com/SETTER2000/gofermart/internal/usecase"
	"github.com/SETTER2000/gofermart/internal/usecase/encryp"
	"github.com/SETTER2000/gofermart/internal/usecase/repo"
	"github.com/SETTER2000/gofermart/pkg/log/logger"
	"github.com/SETTER2000/gofermart/scripts/luna"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"time"
)

type gofermartRoutes struct {
	s   usecase.Gofermart
	l   logger.Interface
	cfg *config.Config
	c   *client.AClient
}

func newGofermartRoutes(handler chi.Router, s usecase.Gofermart, l logger.Interface, cfg *config.Config, c *client.AClient) {
	sr := &gofermartRoutes{s, l, cfg, c}
	handler.Route("/user", func(r chi.Router) {
		r.Get("/orders", sr.handleUserOrdersGet)
		r.Get("/balance", sr.handleUserBalanceGet)
		r.Get("/withdrawals", sr.handleUserWithdrawalsGet)
		r.Post("/register", sr.handleUserCreate)
		r.Post("/login", sr.handleUserLogin)
		r.Post("/orders", sr.handleUserOrders)
		r.Post("/balance/withdraw", sr.handleUserBalanceWithdraw)
	})
}

// GET /ping, который при запросе проверяет соединение с базой данных
// при успешной проверке хендлер должен вернуть HTTP-статус 200 OK
// при неуспешной — 500 Internal Server Error
func (sr *gofermartRoutes) connect(w http.ResponseWriter, r *http.Request) {
	dsn, ok := os.LookupEnv("DATABASE_URI")
	if !ok || dsn == "" {
		dsn = sr.cfg.Storage.ConnectDB
		if dsn == "" {
			sr.l.Info("connect DSN string is empty: %v\n", dsn)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		db, err := pgx.Connect(r.Context(), os.Getenv("DATABASE_URI"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer db.Close(context.Background())
		sr.respond(w, r, http.StatusOK, "connect... ")
	}
}

// @Summary     Return JSON empty
// @Description Redirect to log URL
// @ID          Регистрация пользователя
// @Tags  	    gofermart
// @Accept      json
// @Produce     json
// @Success     200 {object} response — пользователь успешно зарегистрирован и аутентифицирован
// @Failure     400 {object} response — неверный формат запроса
// @Failure     409 {object} response — логин уже занят
// @Failure     500 {object} response — внутренняя ошибка сервера
// @Router      /user/register [post]
func (sr *gofermartRoutes) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	// Регистрация производится по паре логин/пароль.
	// Каждый логин должен быть уникальным. После успешной регистрации
	// должна происходить автоматическая аутентификация пользователя.
	// Для передачи аутентификационных данных используйте
	// механизм cookies или HTTP-заголовок Authorization.
	ctx := r.Context()
	a := &entity.Authentication{Config: sr.cfg}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(body, &a); err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}
	err = sr.s.Register(ctx, a)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			sr.error(w, r, http.StatusConflict, err)
			return
		}
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}
	sr.SessionCreated(w, r, a.ID)
	a.Sanitize()
	sr.respond(w, r, http.StatusOK, a)
}

// @Summary     Return JSON empty
// @ID          Аутентификация пользователя
// @Tags  	    gofermart
// @Accept      json
// @Produce     json
// @Success     200 {object} response — пользователь успешно аутентифицирован
// @Failure     400 {object} response — неверный формат запроса
// @Failure     401 {object} response — неверная пара логин/пароль
// @Failure     500 {object} response — внутренняя ошибка сервера
// @Router      /user/login [post]
func (sr *gofermartRoutes) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	a := &entity.Authentication{Config: sr.cfg}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(body, &a); err != nil {
		sr.error(w, r, http.StatusInternalServerError, err)
		return
	}
	if err := a.Validate(); err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}
	if err := a.BeforeCreate(); err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}
	u, err := sr.s.UserFindByLogin(r.Context(), a.Login) // will return the user by login
	if err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}
	if !u.ComparePassword(a.Password) {
		sr.error(w, r, http.StatusUnauthorized, ErrIncorrectLoginOrPass)
		return
	}
	sr.SessionCreated(w, r, u.ID)
	sr.respond(w, r, http.StatusOK, nil)
}

// IsAuthenticated проверка авторизирован ли пользователь, по сути проверка токена на пригодность
// TODO возможно здесь нужно реализовать проверку времени жизни токена
func (sr *gofermartRoutes) IsAuthenticated(w http.ResponseWriter, r *http.Request) (string, error) {
	var e encryp.Encrypt
	ctx := r.Context()
	at, err := r.Cookie("access_token")
	if err == http.ErrNoCookie {
		return "", err
	}
	// если кука обнаружена, то расшифровываем токен,
	// содержащийся в ней, и проверяем подпись
	dt, err := e.DecryptToken(at.Value, sr.cfg.SecretKey)
	if err != nil {
		fmt.Printf("error decrypt cookie: %e", err)
		return "", err
	}

	_, err = sr.s.UserFindByID(r.Context(), dt)
	if err != nil {
		return "", err
	}
	return ctx.Value(sr.cfg.Cookie.AccessTokenName).(string), nil
}

// @Summary     Return JSON empty
// @Description Загрузка номера заказа
// @ID          handleUserOrders
// @Tags  	    gofermart
// @Accept      text
// @Produce     text
// @Success     200 {object} response — номер заказа уже был загружен этим пользователем
// @Success     202 {object} response — новый номер заказа принят в обработку
// @Failure     400 {object} response — неверный формат запроса
// @Failure     401 {object} response — пользователь не аутентифицирован
// @Failure     409 {object} response — номер заказа уже был загружен другим пользователем
// @Failure     422 {object} response — неверный формат номера заказа
// @Failure     500 {object} response — внутренняя ошибка сервера
// @Router      /user/orders [post]
func (sr *gofermartRoutes) handleUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := sr.IsAuthenticated(w, r)
	if err != nil {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}
	// проверка правильного формата запроса
	if !sr.ContentTypeCheck(w, r, "text/plain") {
		sr.respond(w, r, http.StatusBadRequest, "неверный формат запроса")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sr.respond(w, r, http.StatusInternalServerError, nil)
		return
	}

	order := string(body)
	o := entity.Order{Config: sr.cfg}
	o.Number, err = strconv.Atoi(order)
	if err != nil {
		sr.respond(w, r, http.StatusBadRequest, "неверный формат запроса")
		return
	}
	// проверка формата номера заказа
	if !luna.Luna(o.Number) { // цветы, цветы
		sr.respond(w, r, http.StatusUnprocessableEntity, "неверный формат номера заказа")
		return
	}

	// взаимодействие с системой расчёта начислений баллов лояльности
	lp, err := sr.c.LoyaltyFind(order)
	if err != nil {
		sr.l.Error(err, "http - v1 - accrualClient")
	}
	ctx := r.Context()
	if lp.Status != "PROCESSED" && lp.Status != "INVALID" {
		lCh := make(chan entity.LoyaltyStatus, 1)
		// входные значения кладём в inputCh
		go func(l entity.LoyaltyStatus) {
			lCh <- l
			l = *sr.c.Run(lCh)
			sr.s.OrderUpdateUserID(ctx, &l)
			close(lCh)
		}(*lp)
	}
	o.Accrual = lp.Accrual
	o.Status = lp.Status
	o.UserID = userID

	_, err = sr.s.OrderAdd(ctx, &o)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			data2 := entity.Order{Config: sr.cfg, Number: o.Number}
			or, err := sr.s.OrderFindByID(ctx, &data2)
			if err != nil {
				sr.l.Error(err, "http - v1 - handleUserOrders")
				sr.respond(w, r, http.StatusBadRequest, nil)
				return
			}

			if or.UserID != o.UserID {
				sr.respond(w, r, http.StatusConflict, "номер заказа уже был загружен другим пользователем")
			}
			sr.respond(w, r, http.StatusOK, nil)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("новый номер заказа принят в обработку"))
}

// @Summary     Return JSON
// @Description Запрос на списание средств
// @ID          handleUserBalanceWithdraw
// @Tags  	    gofermart
// @Accept      json
// @Produce     json
// @Success     200 {object} response — успешная обработка запроса
// @Failure     401 {object} response — пользователь не авторизован
// @Failure     402 {object} response — на счету недостаточно средств
// @Failure     422 {object} response — неверный формат номера заказа
// @Failure     500 {object} response — внутренняя ошибка сервера
// @Router      /user/balance/withdraw [post]
func (sr *gofermartRoutes) handleUserBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := sr.IsAuthenticated(w, r)
	if err != nil {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}

	wd := entity.Withdraw{}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(body, &wd); err != nil {
		sr.error(w, r, http.StatusInternalServerError, err)
		return
	}

	o := entity.Order{}
	o.Number, _ = strconv.Atoi(wd.NumOrder)
	if !luna.Luna(o.Number) { // цветы, цветы
		sr.respond(w, r, http.StatusUnprocessableEntity, "неверный формат номера заказа")
		return
	}

	// добавить ордер
	err = sr.redirectToOrderAdd(w, r, wd.NumOrder)
	if err != nil {
		sr.error(w, r, http.StatusConflict, err)
		return
	}

	wd.Order = &o
	err = sr.s.OrderBalanceWithdrawAdd(ctx, &wd) // списание
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			sr.error(w, r, http.StatusConflict, err)
			return
		} else if errors.Is(err, repo.ErrInsufficientFundsAccount) {
			// TODO 402 — на счету недостаточно средств;
			sr.error(w, r, http.StatusPaymentRequired, err)
			return
		}
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}

	sr.respond(w, r, http.StatusOK, "успешная обработка запроса")
}
func (sr *gofermartRoutes) redirectToOrderAdd(w http.ResponseWriter, r *http.Request, order string) error {
	link := sr.cfg.HTTP.BaseURL + "/api/user/orders"

	// конструируем контекст с Timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	// функция cancel() позволяет при необходимости остановить операции
	defer cancel()
	// собираем запрос с контекстом
	req, _ := http.NewRequestWithContext(ctx, "POST", link, bytes.NewBufferString(order))
	// конструируем клиент
	client := &http.Client{}

	at, err := r.Cookie("access_token")
	// если куки нет, то ничего не делаем
	if err == http.ErrNoCookie {
		return fmt.Errorf("error cookie, empty cookie")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("error cookiejar query add order: %e", err)
	} else {
		client.Jar = jar
	}

	req.AddCookie(at)
	req.Header.Set("Content-Type", "text/plain")
	// отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error post query add order: %e", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return fmt.Errorf("status %d: вы не можете использовать данный номер заказа ", resp.StatusCode)
	}

	return nil
}

// @Summary     Return JSON empty
// @Description Получение списка загруженных номеров заказов
// @ID          handleUserOrdersGet
// @Tags  	    gofermart
// @Accept      text
// @Produce     text
// @Success     200 {object} response — успешная обработка запроса
// @Success     204 {object} response — нет данных для ответа
// @Failure     401 {object} response — пользователь не аутентифицирован
// @Failure     500 {object} response — внутренняя ошибка сервера
// @Router      /user/orders [get]
func (sr *gofermartRoutes) handleUserOrdersGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := entity.User{}

	userID, err := sr.IsAuthenticated(w, r)
	if err != nil {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}

	u.UserID = userID
	ol, err := sr.s.OrderListUserID(ctx, &u)
	if err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
	}

	if len(*ol) < 1 {
		sr.respond(w, r, http.StatusNoContent, "нет данных для ответа")
	}

	sr.respond(w, r, http.StatusOK, ol)
}

// @Summary     Return JSON
// @Description Получение текущего баланса пользователя
// @ID          handleUserBalanceGet
// @Tags  	    gofermart
// @Accept      json
// @Produce     json
// @Success     200 {object} response — успешная обработка запроса
// @Failure     401 {object} response — пользователь не аутентифицирован
// @Failure     500 {object} response — внутренняя ошибка сервера
// @Router      /user/balance [get]
func (sr *gofermartRoutes) handleUserBalanceGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := sr.IsAuthenticated(w, r)
	if err != nil {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}

	b, err := sr.s.FindBalance(ctx)
	if err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
	}

	sr.respond(w, r, http.StatusOK, b)
}

// @Summary     Return JSON
// @Description Получение информации о выводе средств с накопительного счёта пользователем.
// Хендлер доступен только авторизованному пользователю. Факты выводов в выдаче должны быть
// отсортированы по времени вывода от самых старых к самым новым. Формат даты — RFC3339.
// @ID          handleUserWithdrawalsGet
// @Tags  	    gofermart
// @Accept      json
// @Produce     json
// @Success     200 {object} response — успешная обработка запроса
// @Success     204 {object} response — нет данных для ответа
// @Failure     401 {object} response — пользователь не аутентифицирован
// @Failure     500 {object} response — внутренняя ошибка сервера
// @Router      /api/user/withdrawals [get]
func (sr *gofermartRoutes) handleUserWithdrawalsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := sr.IsAuthenticated(w, r)
	if err != nil {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}

	ol, err := sr.s.FindWithdrawalsList(ctx)
	if err != nil {
		sr.error(w, r, http.StatusInternalServerError, err)
	}

	if len(*ol) < 1 {
		sr.respond(w, r, http.StatusNoContent, "нет ни одного списания")
	}

	sr.respond(w, r, http.StatusOK, ol)
}

// ContentTypeCheck проверка соответствует ли content-type запроса endpoint
func (sr *gofermartRoutes) ContentTypeCheck(w http.ResponseWriter, r *http.Request, t string) bool {
	return r.Header.Get("Content-Type") == t
}

func (sr *gofermartRoutes) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	sr.respond(w, r, code, map[string]string{"error": err.Error()})
}
func (sr *gofermartRoutes) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (sr *gofermartRoutes) SessionCreated(w http.ResponseWriter, r *http.Request, data string) error {
	en := encryp.Encrypt{}
	// ...создать подписанный секретным ключом токен,
	token, err := en.EncryptToken(sr.cfg.SecretKey, data)
	if err != nil {
		fmt.Printf("Encrypt error: %v\n", err)
		return err
	}
	// ...установить куку с именем access_token,
	// а в качестве значения установить зашифрованный,
	// подписанный токен
	http.SetCookie(w, &http.Cookie{
		Name:    "access_token",
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(time.Minute * 60),
	})

	_, err = en.DecryptToken(token, sr.cfg.SecretKey)
	if err != nil {
		fmt.Printf(" Decrypt error: %v\n", err)
		return err
	}

	return nil
}
