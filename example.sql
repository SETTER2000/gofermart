SELECT * FROM "user";
SELECT * FROM "order";
SELECT * FROM balance;
SELECT * FROM "gofermart";

DROP TABLE public."order";
DROP TABLE public."user";
DROP TABLE public."gofermart";
-- DROP TABLE public."ro";

TRUNCATE TABLE public."order";
TRUNCATE TABLE public."user" CASCADE;

CREATE TABLE IF NOT EXISTS public.user
(
    user_id          UUID         NOT NULL DEFAULT uuid_generate_v1(),
    CONSTRAINT user_id_user PRIMARY KEY (user_id),
    login            VARCHAR(100) NOT NULL UNIQUE,
    encrypted_passwd VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.gofermart
(
    slug    VARCHAR(300) NOT NULL,
    url     VARCHAR      NOT NULL UNIQUE,
    user_id VARCHAR(300) NOT NULL,
    del     BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.order
(
    order_id NUMERIC PRIMARY KEY,
    user_id  uuid,
    foreign key (user_id) references public."user" (user_id)
        match simple on update no action on delete no action,
    time_in  TIMESTAMP(0)
);

-- New DB

CREATE TABLE IF NOT EXISTS public.gofermart
(
    slug    VARCHAR(300) NOT NULL,
    url     VARCHAR      NOT NULL UNIQUE,
    user_id VARCHAR(300) NOT NULL,
    del     BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.user
(
    user_id          UUID         NOT NULL DEFAULT uuid_generate_v1(),
    CONSTRAINT user_id_user PRIMARY KEY (user_id),
    login            VARCHAR(100) NOT NULL UNIQUE,
    encrypted_passwd VARCHAR(100) NOT NULL
);
--
-- REGISTERED — заказ зарегистрирован, но не начисление не рассчитано;
-- INVALID — заказ не принят к расчёту, и вознаграждение не будет начислено;
-- PROCESSING — расчёт начисления в процессе;
-- PROCESSED — расчёт начисления окончен;
--
-- CREATE TYPE state AS ENUM ('REGISTERED', 'INVALID','PROCESSING', 'PROCESSED') ;
--create types
DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'state') THEN
            CREATE TYPE state AS ENUM ('NEW', 'REGISTERED', 'INVALID','PROCESSING', 'PROCESSED');
        END IF;
        --more types here...
    END
$$;
CREATE TABLE IF NOT EXISTS public.order
(
    number      NUMERIC PRIMARY KEY,
    user_id     uuid,
    uploaded_at TIMESTAMP(0) WITH TIME ZONE,
    accrual     NUMERIC(8, 3),
    status      state,
    FOREIGN KEY (user_id) REFERENCES public."user" (user_id)
        MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- INSERT INTO "order"(number, user_id, uploaded_at, accrual, status) VALUES (424242424242, '0742f229-c951-11ed-9cea-e96171a9a322', now(), 500.65, 'NEW');
SELECT *
FROM "order";
DROP TABLE "order";

CREATE TABLE IF NOT EXISTS public.balance
(
--     id           SERIAL PRIMARY KEY,
    number       NUMERIC NOT NULL,
    user_id      uuid NOT NULL,
    sum          NUMERIC(8, 3) NOT NULL CHECK (sum > 0),
    processed_at TIMESTAMP(0) WITH TIME ZONE,
    FOREIGN KEY (number) REFERENCES public."order" (number)
        MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION,
    FOREIGN KEY (user_id) REFERENCES public."user" (user_id)
        MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION,
    PRIMARY KEY(number, user_id)
);
SELECT * FROM balance;
INSERT INTO "balance"(number, user_id, sum, processed_at) VALUES ('8376841766', '8fa5524e-c962-11ed-9cea-e96171a9a322', 65, now());
INSERT INTO "balance"(number, user_id, sum, processed_at) VALUES ('4242424242424242', '8fa5524e-c962-11ed-9cea-e96171a9a322', 65, now());
INSERT INTO "balance"(number, user_id, sum, processed_at) VALUES ('101725', '8fa5524e-c962-11ed-9cea-e96171a9a322', 65, now());
INSERT INTO "balance"(number, user_id, sum, processed_at) VALUES ('7733868', 'a689880b-c961-11ed-9cea-e96171a9a322', 65, now());
INSERT INTO "balance"(number, user_id, sum, processed_at) VALUES ('7733868', 'a689880b-c961-11ed-9cea-e96171a9a322', 65, now());
INSERT INTO "balance"(number, user_id, sum, processed_at) VALUES ('8376841766', 'a689880b-c961-11ed-9cea-e96171a9a322', 65, now());
SELECT * FROM balance;

create table ro
(
    time_in TIMESTAMP(0),
    name    VARCHAR(30)
);

select CURRENT_TIMESTAMP(0);

insert into ro (tt, name)
values (now(), 'Na');
SELECT *
FROM ro;


SELECT to_char('2022-03-31 17:39:23.5'::timestamp(0), 'YYYY-MM-DD"T"HH24:MI:SSOF:00'),
       to_char('2022-03-31 17:39:23.500'::timestamp(0), 'YYYY-MM-DD"T"HH24:MI:SSOF'),
       to_char('2022-03-31 17:39:23.5123456789'::timestamp(0), 'YYYY-MM-DD"T"HH24:MI:SS.MSOF'),
       to_char('2022-03-31 17:39:23.5123456789'::timestamp(0), 'YYYY-MM-DD"T"HH24:MI:SS.USOF')
           insert
into "order" (number, user_id, uploaded_at, accrual, status)
values ();