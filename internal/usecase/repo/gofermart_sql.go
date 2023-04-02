package repo

import (
	"context"
	"database/sql"
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
	"math"
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

// OrderIn записать новый заказ
func (i *InSQL) OrderIn(ctx context.Context, o *entity.Order) error {
	tx, err := i.w.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO public.order (number, user_id, uploaded_at, status, accrual) VALUES ($1,$2, now(),$3,$4)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, o.Number, o.UserID, o.Status, o.Accrual)
	if err, ok := err.(*pgconn.PgError); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return NewConflictError("old url", "http://testiki", ErrAlreadyExists)
		}
		return err
	}

	return tx.Commit()
}

// OrderPostBalanceWithdraw запрос на списание средств
func (i *InSQL) OrderPostBalanceWithdraw(ctx context.Context, wd *entity.Withdraw) error {
	balance, err := i.Balance(ctx)
	if err != nil {
		log.Fatalf("balance error %e", err)
		return err
	}

	if balance.Current < wd.Sum {
		return NewConflictError("old url", "", ErrInsufficientFundsAccount)
	}

	tx, err := i.r.db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO balance (number, user_id, sum, processed_at) VALUES ($1,$2,$3, now())")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, wd.NumOrder, wd.UserID, wd.Sum)
	if err, ok := err.(*pgconn.PgError); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return NewConflictError("old url", "http://testiki", ErrAlreadyExists)
		}
		return err
	}

	return tx.Commit()
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

	sh.Slug = slug
	sh.URL = url
	sh.UserID = id
	sh.Del = del

	return sh, nil
}

// OrderGetByNumber поиск по ID ордера
func (i *InSQL) OrderGetByNumber(ctx context.Context, o *entity.Order) (*entity.OrderResponse, error) {
	or := entity.OrderResponse{}
	q := `SELECT number, user_id, uploaded_at, accrual, status FROM "order" WHERE number=:number`
	rows, err := i.w.db.NamedQuery(q, o)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.StructScan(&or)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return &or, nil
}

func (i *InSQL) GetByLogin(ctx context.Context, l string) (*entity.Authentication, error) {
	var a entity.Authentication
	var userID, login, encrypt string

	q := `SELECT user_id, login, encrypted_passwd FROM "user" WHERE login=$1`
	rows, err := i.w.db.Queryx(q, l)
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

// OrderListGetUserID  вернуть список заказов по UserID
func (i *InSQL) OrderListGetUserID(ctx context.Context, u *entity.User) (*entity.OrderList, error) {
	var number, userID, uploadedAt, status string
	var accrual float32

	// 2020-12-10T15:15:45+03:00
	q := `SELECT number, user_id,  uploaded_at, status, accrual FROM "order" WHERE user_id=$1 ORDER BY uploaded_at`

	rows, err := i.w.db.Queryx(q, u.UserID)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	or := entity.OrderResponse{}
	ol := entity.OrderList{}
	for rows.Next() {
		err = rows.Scan(&number, &userID, &uploadedAt, &status, &accrual)
		if err != nil {
			return nil, err
		}

		or.Number = number
		or.Status = status
		or.Accrual = accrual
		or.UploadedAt = uploadedAt
		ol = append(ol, or)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &ol, nil
}

// OrderListGetStatus вернуть список всех заказов в соответствие статуса
func (i *InSQL) OrderListGetStatus(ctx context.Context) (*entity.OrderList, error) {
	var number, userID, uploadedAt, status string
	var accrual float32
	processed := "PROCESSED"
	invalid := "INVALID"
	// 2020-12-10T15:15:45+03:00
	q := `SELECT number, user_id,  uploaded_at, status, accrual FROM "order" WHERE status!=$1 AND status!=$2 ORDER BY uploaded_at`

	rows, err := i.w.db.Queryx(q, processed, invalid)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	or := entity.OrderResponse{}
	ol := entity.OrderList{}
	for rows.Next() {
		err = rows.Scan(&number, &userID, &uploadedAt, &status, &accrual)
		if err != nil {
			return nil, err
		}

		or.Number = number
		or.Status = status
		or.Accrual = accrual
		or.UploadedAt = uploadedAt
		ol = append(ol, or)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &ol, nil
}

// Balance получение текущего баланса пользователя
func (i *InSQL) Balance(ctx context.Context) (*entity.Balance, error) {
	var current, withdrawn sql.NullFloat64

	q := `SELECT SUM(accrual) AS current, (SELECT SUM(sum) FROM "balance" WHERE user_id=$1) AS withdrawn FROM "order" WHERE user_id=$1`

	rows, err := i.w.db.Queryx(q, ctx.Value(i.cfg.Cookie.AccessTokenName).(string))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	b := entity.Balance{}
	for rows.Next() {
		err = rows.Scan(&current, &withdrawn)
		if err != nil {
			return nil, err
		}

		b.Current = float32(math.Round((current.Float64-withdrawn.Float64)*100) / 100)
		b.Withdraw = float32(math.Round(withdrawn.Float64*100) / 100)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &b, nil
}
func (i *InSQL) BalanceGetAll(ctx context.Context) (*entity.WithdrawalsList, error) {
	var number, userID, processedAt string
	var sum float32

	q := `SELECT number, user_id, sum, processed_at FROM "balance" WHERE user_id=$1 ORDER BY processed_at`
	us := ctx.Value(i.cfg.Cookie.AccessTokenName).(string)

	rows, err := i.w.db.Queryx(q, us)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	wd := entity.WithdrawResponse{}
	ol := entity.WithdrawalsList{}

	for rows.Next() {
		err = rows.Scan(&number, &userID, &sum, &processedAt)
		if err != nil {
			return nil, err
		}
		wd.NumOrder = number
		wd.Sum = sum
		wd.ProcessedAt = processedAt
		ol = append(ol, wd)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &ol, nil
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

// UpdateOrder обновление заказа
func (i *InSQL) UpdateOrder(ctx context.Context, ls *entity.LoyaltyStatus) error {
	q := `UPDATE "order" SET status=:status, accrual=:accrual WHERE number=:order`

	rows, err := i.w.db.NamedQuery(q, ls)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}

// UpdateOrderUserID обновление заказа по id пользователя
func (i *InSQL) UpdateOrderUserID(ctx context.Context, ls *entity.LoyaltyStatus) error {
	tx, err := i.w.db.Begin()
	ct := context.Background()
	userID := ctx.Value(i.cfg.Cookie.AccessTokenName).(string)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ct, "UPDATE \"order\" SET status=$1, accrual=$2 WHERE number=$3 AND user_id=$4")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(ls.Status, ls.Accrual, ls.Order, userID); err != nil {
		return err
	}

	return tx.Commit()
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
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS public.user
(
	user_id UUID NOT NULL DEFAULT uuid_generate_v1(),
  CONSTRAINT user_id_user PRIMARY KEY (user_id),
    login VARCHAR(100) NOT NULL UNIQUE,
    encrypted_passwd VARCHAR(100) NOT NULL
);
--create types
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'state') THEN
        CREATE TYPE state AS ENUM ('NEW', 'REGISTERED', 'INVALID','PROCESSING', 'PROCESSED');
    END IF;
    --more types here...
END$$;
CREATE TABLE IF NOT EXISTS public.order
(
    number      NUMERIC PRIMARY KEY,
    user_id     uuid,
    uploaded_at TIMESTAMP(0) WITH TIME ZONE,
    accrual     float4 DEFAULT 0,
    status      state,
    FOREIGN KEY (user_id) REFERENCES public."user" (user_id)
        MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE TABLE IF NOT EXISTS public.balance
(
    id           SERIAL PRIMARY KEY,
    number       NUMERIC NOT NULL,
    user_id      uuid NOT NULL,
    sum          float4 NOT NULL CHECK (sum >= 0),
    processed_at TIMESTAMP(0) WITH TIME ZONE,
--     FOREIGN KEY (number) REFERENCES public."order" (number)
--         MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION,
    FOREIGN KEY (user_id) REFERENCES public."user" (user_id)
        MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION
--     PRIMARY KEY(number, user_id, sum)
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
