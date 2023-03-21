package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/entity"
	"github.com/SETTER2000/gofermart/internal/usecase"
	"github.com/SETTER2000/gofermart/internal/usecase/encryp"
	"github.com/SETTER2000/gofermart/internal/usecase/repo"
	"github.com/SETTER2000/gofermart/pkg/log/logger"
	"github.com/SETTER2000/gofermart/scripts"
	"github.com/SETTER2000/gofermart/scripts/luna"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type gofermartRoutes struct {
	s   usecase.Gofermart
	l   logger.Interface
	cfg *config.Config
}

func newGofermartRoutes(handler chi.Router, s usecase.Gofermart, l logger.Interface, cfg *config.Config) {
	sr := &gofermartRoutes{s, l, cfg}
	handler.Route("/user", func(r chi.Router) {
		r.Delete("/urls", sr.delUrls2)
		r.Get("/urls", sr.urls)
		r.Get("/orders", sr.handleUserOrdersGet)
		r.Post("/register", sr.handleUserCreate)
		r.Post("/login", sr.handleUserLogin)
		r.Post("/orders", sr.handleUserOrders)
		r.Post("/balance/withdraw", sr.handleUserBalanceWithdraw)
	})
	handler.Route("/shorten", func(r chi.Router) {
		r.Post("/", sr.shorten) // POST /
		r.Post("/batch", sr.batch)
	})
}

// @Summary     Return short URL
// @Description Redirect to long URL
// @ID          ShortLink
// @Tags  	    gofermart
// @Accept      text
// @Produce     text
// @Success     307 {object} string
// @Failure     500 {object} response
// @Router      /{key} [get]

func (sr *gofermartRoutes) shortLink(w http.ResponseWriter, r *http.Request) {
	short := chi.URLParam(r, "key")
	data := entity.Gofermart{Config: sr.cfg}
	data.Slug = short
	sh, err := sr.s.ShortLink(r.Context(), &data)
	if err != nil {
		sr.l.Error(err, "http - v1 - shortLink")
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}
	// при запросе удалённого URL с помощью хендлера GET /{id} нужно вернуть статус 410 Gone
	if sh.Del {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Encoding", "gzip")
	w.Header().Set("Location", sh.URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
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

// @Summary     Return short URL
// @Description Redirect to long URL
// @ID          longLink
// @Tags  	    gofermart
// @Accept      text
// @Produce     text
// @Success     201 {object} string
// @Failure     500 {object} response
// @Router      / [post]
func (sr *gofermartRoutes) longLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		sr.error(w, r, http.StatusInternalServerError, err)
		return
	}
	data := entity.Gofermart{Config: sr.cfg}
	data.URL = string(body)
	//data.URL, _ = scripts.Trim(string(body), "")
	data.Slug = scripts.UniqueString()
	//data.UserID = r.Context().Value("access_token").(string)
	gofermart, err := sr.s.LongLink(ctx, &data)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			data2 := entity.Gofermart{Config: sr.cfg, URL: data.URL}
			//data2.URL = data.URL
			sh, err := sr.s.ShortLink(ctx, &data2)
			if err != nil {
				sr.l.Error(err, "http - v2 - shortLink")
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
				return
			}
			gofermart = sh.Slug
			w.Header().Set("Content-Type", http.DetectContentType(body))
			w.WriteHeader(http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	d := scripts.GetHost(sr.cfg.HTTP, gofermart)
	//w.Header().Set("Content-Type", http.DetectContentType(body))
	//w.WriteHeader(http.StatusCreated)
	//w.Write([]byte(d))
	sr.respond(w, r, http.StatusCreated, d)
}

// GET
func (sr *gofermartRoutes) urls(w http.ResponseWriter, r *http.Request) {
	u := entity.User{}
	userID := r.Context().Value("access_token")
	if userID == nil {
		w.Write([]byte(fmt.Sprintf("Not access_token and user_id: %s", userID)))
	}
	u.UserID = fmt.Sprintf("%s", userID)
	user, err := sr.s.UserAllLink(r.Context(), &u)
	if err != nil {
		sr.l.Error(err, "http - v1 - shortLink")
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}

	obj, err := json.Marshal(user.Urls)
	if err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if string(obj) == "null" {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write(obj)
}

// @Summary     Return JSON short URL
// @Description Redirect to long URL
// @ID          shorten
// @Tags  	    gofermart
// @Accept      json
// @Produce     json
// @Success     307 {object} string
// @Failure     500 {object} response
// @Router      /shorten [post]
func (sr *gofermartRoutes) shorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := entity.Gofermart{Config: sr.cfg}
	resp := entity.GofermartResponse{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data.Slug = scripts.UniqueString()
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}
	//data.UserID = r.Context().Value(sr.cfg.Cookie.AccessTokenName).(string)
	resp.URL, err = sr.s.Shorten(ctx, &data)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			data2 := entity.Gofermart{Config: sr.cfg}
			data2.URL = data.URL
			sh, err := sr.s.ShortLink(ctx, &data2)
			if err != nil {
				sr.error(w, r, http.StatusBadRequest, err)
			}
			resp.URL = sh.Slug
			sr.respond(w, r, http.StatusConflict, resp)
		} else {
			sr.error(w, r, http.StatusBadRequest, err)
			return
		}
	}
	resp.URL = scripts.GetHost(sr.cfg.HTTP, resp.URL)
	obj, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(obj)
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
func (sr *gofermartRoutes) IsAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	var e encryp.Encrypt
	at, err := r.Cookie("access_token")
	if err == http.ErrNoCookie {
		return false
	}
	// если кука обнаружена, то расшифровываем токен,
	// содержащийся в ней, и проверяем подпись
	dt, err := e.DecryptToken(at.Value, sr.cfg.SecretKey)
	if err != nil {
		fmt.Printf("error decrypt cookie: %e", err)
		return false
	}
	//fmt.Printf("User ID расшифрованный из токена:: %s\n", dt)
	_, err = sr.s.UserFindByID(r.Context(), dt)
	return err == nil
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
	// TODO менять
	// проверка аутентификации
	if !sr.IsAuthenticated(w, r) {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}
	// проверка правильного формата запроса
	if !sr.ContentTypeCheck(w, r, "text/plain") {
		sr.respond(w, r, http.StatusBadRequest, "неверный формат запроса")
		return
	}
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sr.respond(w, r, http.StatusInternalServerError, nil)
		return
	}
	o := entity.Order{Config: sr.cfg}
	o.Number = strings.TrimSpace(string(body))
	order, _ := strconv.Atoi(o.Number)
	if !luna.Luna(order) { // цветы, цветы
		sr.respond(w, r, http.StatusUnprocessableEntity, "неверный формат номера заказа")
		return
	}
	o.UserID = ctx.Value("access_token").(string)
	o.Status = "REGISTERED"
	strAcc := "500"
	o.Accrual = strings.TrimSpace(strAcc)

	_, err = sr.s.OrderAdd(ctx, &o)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			data2 := entity.Order{Config: sr.cfg, Number: o.Number}
			o2, err := sr.s.OrderFindByID(ctx, &data2)
			if err != nil {
				sr.l.Error(err, "http - v1 - handleUserOrders")
				sr.respond(w, r, http.StatusBadRequest, nil)
				return
			}

			if o2.UserID != o.UserID {
				sr.respond(w, r, http.StatusConflict, "номер заказа уже был загружен другим пользователем")
			}
			sr.respond(w, r, http.StatusOK, nil)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	//w.Header().Set("Content-Type", http.DetectContentType(body))
	//w.WriteHeader(http.StatusAccepted)
	//w.Write([]byte("Новый номер заказа принят в обработку!"))

	w.Header().Set("Content-Type", "text/plain")
	//w.Header().Add("Content-Encoding", "gzip")
	//w.Header().Set("Location", gofermart)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("новый номер заказа принят в обработку"))
	//sr.respond(w, r, http.StatusAccepted, gofermart)
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
	// проверка аутентификации
	if !sr.IsAuthenticated(w, r) {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}
	// проверка правильного формата запроса
	if !sr.ContentTypeCheck(w, r, "text/plain") {
		sr.respond(w, r, http.StatusBadRequest, "неверный формат запроса")
		return
	}
	o := entity.Order{Config: sr.cfg}
	wd := entity.Withdraw{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(body, &wd); err != nil {
		sr.error(w, r, http.StatusInternalServerError, err)
		return
	}
	o.Number = strings.TrimSpace(wd.Order)
	order, _ := strconv.Atoi(o.Number)
	if !luna.Luna(order) { // цветы, цветы
		sr.respond(w, r, http.StatusUnprocessableEntity, "неверный формат номера заказа")
		return
	}
	o.UserID = ctx.Value("access_token").(string)
	o.Status = "REGISTERED"
	strAcc := "500"
	o.Accrual = strings.TrimSpace(strAcc)

	err = sr.s.BalanceWithdraw(ctx, &wd)
	//_, err = sr.s.OrderAdd(ctx, &o)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			data2 := entity.Order{Config: sr.cfg, Number: o.Number}
			o2, err := sr.s.OrderFindByID(ctx, &data2)
			if err != nil {
				sr.l.Error(err, "http - v1 - handleUserOrders")
				sr.respond(w, r, http.StatusBadRequest, nil)
				return
			}

			if o2.UserID != o.UserID {
				sr.respond(w, r, http.StatusConflict, "номер заказа уже был загружен другим пользователем")
			}
			sr.respond(w, r, http.StatusOK, nil)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	sr.respond(w, r, http.StatusOK, "в разработке")
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
	// проверка аутентификации
	if !sr.IsAuthenticated(w, r) {
		sr.respond(w, r, http.StatusUnauthorized, nil)
		return
	}
	ctx := r.Context()
	u := entity.User{}
	u.UserID = ctx.Value("access_token").(string)
	ol, err := sr.s.OrderList(ctx, &u)
	if err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
	}
	if len(*ol) < 1 {
		sr.respond(w, r, http.StatusNoContent, "нет данных для ответа")
	}
	sr.respond(w, r, http.StatusOK, ol)
}

// batch
func (sr *gofermartRoutes) batch(w http.ResponseWriter, r *http.Request) {
	data := entity.Gofermart{Config: sr.cfg}
	CorrelationOrigin := entity.CorrelationOrigin{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}
	if err = json.Unmarshal(body, &CorrelationOrigin); err != nil {
		panic(err)
	}

	var rs entity.Response
	var gr entity.GoferResponse
	for _, bt := range CorrelationOrigin {
		data.URL = bt.URL
		data.Slug = bt.Slug
		_, err = sr.s.Shorten(r.Context(), &data)
		if err != nil {
			if errors.Is(err, repo.ErrAlreadyExists) {
				sr.error(w, r, http.StatusConflict, err)
				return
			}
			sr.error(w, r, http.StatusBadRequest, err)
			return
		}
		gr.Slug = data.Slug
		gr.URL = scripts.GetHost(sr.cfg.HTTP, data.Slug)
		rs = append(rs, gr)
	}

	//obj, err := json.Marshal(rs)
	//if err != nil {
	//	sr.error(w, r, http.StatusBadRequest, err)
	//	return
	//}
	//w.Header().Set("Content-Type", "application/json")
	//w.WriteHeader(http.StatusCreated)
	//w.Write(obj)
	sr.respond(w, r, http.StatusCreated, rs)
}

// Асинхронный хендлер DELETE /api/user/urls,
// который принимает список идентификаторов сокращённых URL для удаления
// в формате: [ "a", "b", "c", "d", ...]
// В случае успешного приёма запроса хендлер должен возвращать HTTP-статус 202 Accepted.
// Фактический результат удаления может происходить позже — каким-либо
// образом оповещать пользователя об успешности или неуспешности не нужно.
func (sr *gofermartRoutes) delUrls(w http.ResponseWriter, r *http.Request) {
	var slugs []string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sr.error(w, r, http.StatusBadRequest, err)
		return
	}
	if err = json.Unmarshal(body, &slugs); err != nil {
		panic(err)
	}

	u := entity.User{}
	userID := r.Context().Value("access_token")
	if userID == nil {
		w.Write([]byte(fmt.Sprintf("Not access_token and user_id: %s", userID)))
	}
	u.UserID = fmt.Sprintf("%s", userID)
	u.DelLink = slugs

	//-- fanOut fanIn - multithreading
	const workersCount = 16
	inputCh := make(chan entity.User)
	// входные значения кладём в inputCh
	go func(u entity.User) {
		inputCh <- u
		close(inputCh)
	}(u)
	// здесь fanOut
	fanOutChs := fanOut(inputCh, workersCount)
	workerChs := make([]chan entity.User, 0, workersCount)
	for _, fanOutCh := range fanOutChs {
		workerCh := make(chan entity.User)
		newWorker(sr, r, fanOutCh, workerCh)
		workerChs = append(workerChs, workerCh)
	}

	// здесь fanIn
	for v := range fanIn(workerChs...) {
		sr.l.Info("%s\n", v.UserID)
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Ok!"))
	sr.respond(w, r, http.StatusAccepted, "")
}

// ContentTypeCheck проверка соответствует ли content-type запроса endpoint
func (sr *gofermartRoutes) ContentTypeCheck(w http.ResponseWriter, r *http.Request, t string) bool {
	return r.Header.Get("Content-Type") == t
}
func (sr *gofermartRoutes) delUrls2(w http.ResponseWriter, r *http.Request) {
	var slugs []string
	const workersCount = 10
	inputCh := make(chan entity.User)

	go func() {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &slugs); err != nil {
			panic(err)
		}
		u := entity.User{}
		userID := r.Context().Value("access_token")
		if userID == nil {
			w.Write([]byte(fmt.Sprintf("Not access_token and user_id: %s", userID)))
		}
		u.UserID = fmt.Sprintf("%s", userID)
		u.DelLink = slugs
		inputCh <- u
		close(inputCh)
	}()

	// здесь fanOut
	fanOutChs := fanOut(inputCh, workersCount)
	workerChs := make([]chan entity.User, 0, workersCount)
	for _, fanOutCh := range fanOutChs {
		workerCh := make(chan entity.User)
		newWorker(sr, r, fanOutCh, workerCh)
		workerChs = append(workerChs, workerCh)
	}

	// здесь fanIn
	for v := range fanIn(workerChs...) {
		sr.l.Info("%s\n", v.UserID)
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Ok!"))
}

func newWorker(sr *gofermartRoutes, r *http.Request, input, out chan entity.User) {
	go func() {
		us := entity.User{}
		for u := range input {
			fmt.Printf("UserID: %s, DelLink: %s count: %v ", u.UserID, u.DelLink, len(u.DelLink))
			sr.s.UserDelLink(r.Context(), &u)
			out <- us
		}
		close(out)
	}()
	time.Sleep(50 * time.Millisecond)
}
func fanIn(inputChs ...chan entity.User) chan entity.User {
	// один выходной канал, куда сливаются данные из всех каналов
	outCh := make(chan entity.User)

	go func() {
		wg := &sync.WaitGroup{}

		for _, inputCh := range inputChs {
			wg.Add(1)

			go func(inputCh chan entity.User) {
				defer wg.Done()
				for item := range inputCh {
					outCh <- item
				}
			}(inputCh)
		}

		wg.Wait()
		close(outCh)
	}()

	return outCh
}

func fanOut(inputCh chan entity.User, n int) []chan entity.User {
	chs := make([]chan entity.User, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan entity.User)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan entity.User) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			num, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- num
		}
	}()

	return chs
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
	//ctx := r.Context()
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
	//context.WithValue(ctx, x, idUser)
	//next.ServeHTTP(w, r.WithContext(ctx))
	return nil
}
