package repo

import (
	"context"
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/entity"
	"github.com/SETTER2000/gofermart/scripts"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
)

const (
	driverName = "pgx"
)

type (
	producerSQL struct {
		db *sqlx.DB
	}

	consumerSQL struct {
		db *sqlx.DB
	}

	InSQL struct {
		cfg *config.Config
		r   *consumerSQL
		w   *producerSQL
	}
)

// NewInSQL слой взаимодействия с db в данном случаи с postgresql
func NewInSQL(cfg *config.Config) *InSQL {
	return &InSQL{
		cfg: cfg,
		// создаём новый потребитель
		r: NewSQLConsumer(cfg),
		// создаём новый производитель
		w: NewSQLProducer(cfg),
	}
}

// NewSQLProducer производитель
func NewSQLProducer(cfg *config.Config) *producerSQL {
	connect := Connect(cfg)
	return &producerSQL{
		db: connect,
	}
}

func (i *InSQL) Registry(ctx context.Context, a *entity.Authentication) error {
	if err := a.Validate(); err != nil {
		return err
	}
	if err := a.BeforeCreate(); err != nil {
		return err
	}
	err := i.w.db.QueryRow(
		"INSERT INTO public.user(login, encrypted_passwd) VALUES ($1,$2) RETURNING user_id",
		a.Login,
		a.EncryptPassword,
	).Scan(&a.ID)
	if err, ok := err.(*pgconn.PgError); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return NewConflictError("", "", ErrAlreadyExists)
		}
	}
	return nil
}

func (i *InSQL) Post(ctx context.Context, sh *entity.Gofermart) error {
	stmt, err := i.w.db.Prepare("INSERT INTO public.gofermart (slug, url, user_id) VALUES ($1,$2,$3)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(sh.Slug, sh.URL, sh.UserID)
	if err, ok := err.(*pgconn.PgError); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return NewConflictError("old url", "http://testiki", ErrAlreadyExists)
		}
	}
	return nil
}

func (i *InSQL) OrderIn(ctx context.Context, sh *entity.Gofermart) error {
	stmt, err := i.w.db.Prepare("INSERT INTO public.order (order_id, user_id) VALUES ($1,$2)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(sh.Order, sh.UserID)
	if err, ok := err.(*pgconn.PgError); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return NewConflictError("old url", "http://testiki", ErrAlreadyExists)
		}
		return err
	}
	return nil
}

func (i *InSQL) Put(ctx context.Context, sh *entity.Gofermart) error {
	return i.Post(ctx, sh)
}

// NewSQLConsumer потребитель
func NewSQLConsumer(cfg *config.Config) *consumerSQL {
	connect := Connect(cfg)
	return &consumerSQL{
		db: connect,
	}
}

func (i *InSQL) Get(ctx context.Context, sh *entity.Gofermart) (*entity.Gofermart, error) {
	var slug, url, id string
	var del bool
	rows, err := i.w.db.Query("SELECT slug, url, user_id, del FROM gofermart WHERE slug = $1 OR url = $2 ", sh.Slug, sh.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&slug, &url, &id, &del)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	//sh := entity.Gofermart{}
	sh.Slug = slug
	sh.URL = url
	sh.UserID = id
	sh.Del = del
	return sh, nil
}

// OrderGetByID поиск по ID ордера
func (i *InSQL) OrderGetByID(ctx context.Context, g *entity.Gofermart) (*entity.Gofermart, error) {
	var id, user string
	q := `SELECT order_id, user_id FROM "order" WHERE order_id=$1`
	rows, err := i.w.db.Queryx(q, g.Order)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &user)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	g.Order = id
	g.UserID = user
	return g, nil
}

func (i *InSQL) GetByLogin(ctx context.Context, l string) (*entity.Authentication, error) {
	var a entity.Authentication
	var userID, login, encrypt string

	q := `SELECT user_id, login, encrypted_passwd FROM "user" WHERE login=$1`
	rows, err := i.w.db.Queryx(q, l)
	//rows, err := i.w.db.Query("SELECT * FROM user WHERE login=$1", l)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&userID, &login, &encrypt)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	a.ID = userID
	a.Login = login
	a.EncryptPassword = encrypt
	return &a, nil
}

func (i *InSQL) GetByID(ctx context.Context, l string) (*entity.Authentication, error) {
	var a entity.Authentication
	var userID, login, encrypt string

	q := `SELECT user_id, login, encrypted_passwd FROM "user" WHERE user_id=$1`
	rows, err := i.w.db.Queryx(q, l)
	//rows, err := i.w.db.Query("SELECT * FROM user WHERE login=$1", l)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&userID, &login, &encrypt)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	if userID == "" {
		return nil, ErrNotFound
	}
	a.ID = userID
	a.Login = login
	a.EncryptPassword = encrypt
	return &a, nil
}

func (i *InSQL) GetAll(ctx context.Context, u *entity.User) (*entity.User, error) {
	var slug, url, id string
	q := `SELECT slug, url, user_id FROM gofermart WHERE user_id=$1 AND del=$2`
	rows, err := i.w.db.Queryx(q, u.UserID, false)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	l := entity.List{}
	for rows.Next() {
		err = rows.Scan(&slug, &url, &id)
		if err != nil {
			return nil, err
		}
		l.URL = url
		l.Slug = scripts.GetHost(i.cfg.HTTP, slug)
		u.Urls = append(u.Urls, l)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return u, nil
}
func (i *InSQL) Delete(ctx context.Context, u *entity.User) error {
	q := `UPDATE gofermart SET del = $1
	FROM (SELECT unnest($2::text[]) AS slug) AS data_table
	WHERE gofermart.slug = data_table.slug AND gofermart.user_id=$3`

	rows, err := i.w.db.Queryx(q, true, u.DelLink, u.UserID)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}

func Connect(cfg *config.Config) (db *sqlx.DB) {
	db, _ = sqlx.Open(driverName, cfg.ConnectDB)
	err := db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	n := 100
	db.SetMaxIdleConns(n)
	db.SetMaxOpenConns(n)
	schema := `
-- CREATE EXTENSION "uuid-ossp";

CREATE TABLE IF NOT EXISTS public.user
(
	user_id UUID NOT NULL DEFAULT uuid_generate_v1(),
  CONSTRAINT user_id_user PRIMARY KEY (user_id),
    login VARCHAR(100) NOT NULL UNIQUE,
    encrypted_passwd VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.gofermart
(
   slug    VARCHAR(300) NOT NULL,
   url     VARCHAR NOT NULL UNIQUE,
   user_id VARCHAR(300) NOT NULL,
   del BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.order (
  order_id NUMERIC PRIMARY KEY,
  user_id uuid,
  foreign key (user_id) references public."user" (user_id)
  match simple on update no action on delete no action
);


`
	db.MustExec(schema)
	if err != nil {
		panic(err)
	}
	return db
}

func (i *InSQL) Read() error {
	return nil
}
func (i *InSQL) Save() error {
	return nil
}
