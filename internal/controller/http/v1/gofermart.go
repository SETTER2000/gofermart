package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/entity"
	"github.com/SETTER2000/gofermart/internal/usecase"
	"github.com/SETTER2000/gofermart/internal/usecase/repo"
	"github.com/SETTER2000/gofermart/pkg/log/logger"
	"github.com/SETTER2000/gofermart/scripts"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type shorturlRoutes struct {
	s   usecase.Gofermart
	l   logger.Interface
	cfg *config.Config
}

func newGofermartRoutes(handler chi.Router, s usecase.Gofermart, l logger.Interface, cfg *config.Config) {
	sr := &shorturlRoutes{s, l, cfg}
	handler.Route("/user", func(r chi.Router) {
		r.Get("/urls", sr.urls)
		r.Delete("/urls", sr.delUrls2)
		r.Post("/register", sr.register)
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

func (r *shorturlRoutes) shortLink(w http.ResponseWriter, req *http.Request) {
	gofermart := chi.URLParam(req, "key")
	data := entity.Gofermart{Config: r.cfg}
	data.Slug = gofermart
	sh, err := r.s.ShortLink(req.Context(), &data)
	if err != nil {
		r.l.Error(err, "http - v1 - shortLink")
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}
	// при запросе удалённого URL с помощью хендлера GET /{id} нужно вернуть статус 410 Gone
	if sh.Del {
		w.WriteHeader(http.StatusGone)
		return
	}
	fmt.Println("URL найден заношу в Location: ", sh.URL)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Encoding", "gzip")
	w.Header().Set("Location", sh.URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// GET /ping, который при запросе проверяет соединение с базой данных
// при успешной проверке хендлер должен вернуть HTTP-статус 200 OK
// при неуспешной — 500 Internal Server Error
func (r *shorturlRoutes) connect(res http.ResponseWriter, req *http.Request) {
	dsn, ok := os.LookupEnv("DATABASE_URI")
	if !ok || dsn == "" {
		dsn = r.cfg.Storage.ConnectDB
		if dsn == "" {
			r.l.Info("connect DSN string is empty: %v\n", dsn)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		db, err := pgx.Connect(req.Context(), os.Getenv("DATABASE_URI"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		defer db.Close(context.Background())

		fmt.Printf("connect... \n")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("connect... "))
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
func (r *shorturlRoutes) longLink(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	data := entity.Gofermart{Config: r.cfg}
	data.URL = string(body)
	//data.URL, _ = scripts.Trim(string(body), "")
	data.Slug = scripts.UniqueString()
	//data.UserID = req.Context().Value("access_token").(string)
	gofermart, err := r.s.LongLink(ctx, &data)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			data2 := entity.Gofermart{Config: r.cfg, URL: data.URL}
			//data2.URL = data.URL
			sh, err := r.s.ShortLink(ctx, &data2)
			if err != nil {
				r.l.Error(err, "http - v2 - shortLink")
				http.Error(res, fmt.Sprintf("%v", err), http.StatusBadRequest)
				return
			}
			gofermart = sh.Slug
			res.Header().Set("Content-Type", http.DetectContentType(body))
			res.WriteHeader(http.StatusConflict)
		} else {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	}
	d := scripts.GetHost(r.cfg.HTTP, gofermart)
	res.Header().Set("Content-Type", http.DetectContentType(body))
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(d))
}

// GET
func (r *shorturlRoutes) urls(res http.ResponseWriter, req *http.Request) {
	u := entity.User{}
	userID := req.Context().Value("access_token")
	if userID == nil {
		res.Write([]byte(fmt.Sprintf("Not access_token and user_id: %s", userID)))
	}
	u.UserID = fmt.Sprintf("%s", userID)
	user, err := r.s.UserAllLink(req.Context(), &u)
	if err != nil {
		r.l.Error(err, "http - v1 - shortLink")
		http.Error(res, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	obj, err := json.Marshal(user.Urls)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("%v", len(obj))
	res.Header().Set("Content-Type", "application/json")
	if string(obj) == "null" {
		res.WriteHeader(http.StatusNoContent)
	} else {
		res.WriteHeader(http.StatusOK)
	}
	res.Write(obj)
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
func (r *shorturlRoutes) shorten(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data := entity.Gofermart{Config: r.cfg}
	resp := entity.GofermartResponse{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	data.Slug = scripts.UniqueString()
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}
	//data.UserID = req.Context().Value(r.cfg.Cookie.AccessTokenName).(string)
	resp.URL, err = r.s.Shorten(ctx, &data)

	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			data2 := entity.Gofermart{Config: r.cfg}
			data2.URL = data.URL
			sh, err := r.s.ShortLink(ctx, &data2)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}
			resp.URL = sh.Slug
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)
		} else {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	}
	resp.URL = scripts.GetHost(r.cfg.HTTP, resp.URL)
	obj, err := json.Marshal(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(obj)
}

func (r *shorturlRoutes) register(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data := entity.Gofermart{Config: r.cfg}
	Authentication := entity.Authentication{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(body, &Authentication); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// если значения пустое
	if len(Authentication.Login) < 1 || len(Authentication.Password) < 1 {
		http.Error(res, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	var rs entity.Response
	var sr entity.ShortenResponse

	err = r.s.Register(ctx, &Authentication)
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			res.WriteHeader(http.StatusConflict)
			return
		}
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	sr.Slug = data.Slug
	sr.URL = scripts.GetHost(r.cfg.HTTP, data.Slug)
	rs = append(rs, sr)

	obj, err := json.Marshal(rs)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(obj)
}
func (r *shorturlRoutes) batch(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data := entity.Gofermart{Config: r.cfg}
	CorrelationOrigin := entity.CorrelationOrigin{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &CorrelationOrigin); err != nil {
		panic(err)
	}

	var rs entity.Response
	var sr entity.ShortenResponse
	for _, bt := range CorrelationOrigin {
		data.URL = bt.URL
		data.Slug = bt.Slug
		_, err = r.s.Shorten(ctx, &data)
		if err != nil {
			if errors.Is(err, repo.ErrAlreadyExists) {
				res.WriteHeader(http.StatusConflict)
				return
			}
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		sr.Slug = data.Slug
		sr.URL = scripts.GetHost(r.cfg.HTTP, data.Slug)
		rs = append(rs, sr)
	}

	obj, err := json.Marshal(rs)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(obj)
}

// Асинхронный (КАБУДА!) хендлер DELETE /api/user/urls,
// который принимает список идентификаторов сокращённых URL для удаления
// в формате: [ "a", "b", "c", "d", ...]
// В случае успешного приёма запроса хендлер должен возвращать HTTP-статус 202 Accepted.
// Фактический результат удаления может происходить позже — каким-либо
// образом оповещать пользователя об успешности или неуспешности не нужно.
func (r *shorturlRoutes) delUrls(res http.ResponseWriter, req *http.Request) {
	var slugs []string
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &slugs); err != nil {
		panic(err)
	}

	u := entity.User{}
	userID := req.Context().Value("access_token")
	if userID == nil {
		res.Write([]byte(fmt.Sprintf("Not access_token and user_id: %s", userID)))
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
		newWorker(r, req, fanOutCh, workerCh)
		workerChs = append(workerChs, workerCh)
	}

	// здесь fanIn
	for v := range fanIn(workerChs...) {
		r.l.Info("%s\n", v.UserID)
	}
	//-- end fanOut fanIn

	//-- not multithreading
	//for i := 0; i < len(slugs); i++ {
	//	fmt.Printf("SLUG#%d: %s\n", i, slugs[i])
	//	err = r.s.UserDelLink(req.Context(), &u)
	//}
	//if err != nil {
	//	r.l.Error(err, "http - v1 - delUrls")
	//	http.Error(res, fmt.Sprintf("%v", err), http.StatusBadRequest)
	//	return
	//}
	//-- end not multithreading

	res.WriteHeader(http.StatusAccepted)
	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte("Ok!"))
}

func (r *shorturlRoutes) delUrls2(res http.ResponseWriter, req *http.Request) {
	var slugs []string
	const workersCount = 10
	inputCh := make(chan entity.User)

	go func() {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &slugs); err != nil {
			panic(err)
		}
		u := entity.User{}
		userID := req.Context().Value("access_token")
		if userID == nil {
			res.Write([]byte(fmt.Sprintf("Not access_token and user_id: %s", userID)))
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
		newWorker(r, req, fanOutCh, workerCh)
		workerChs = append(workerChs, workerCh)
	}

	// здесь fanIn
	for v := range fanIn(workerChs...) {
		r.l.Info("%s\n", v.UserID)
	}

	res.WriteHeader(http.StatusAccepted)
	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte("Ok!"))
}

func newWorker(r *shorturlRoutes, req *http.Request, input, out chan entity.User) {
	go func() {
		us := entity.User{}
		for u := range input {
			fmt.Printf("UserID: %s, DelLink: %s count: %v ", u.UserID, u.DelLink, len(u.DelLink))
			r.s.UserDelLink(req.Context(), &u)
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
