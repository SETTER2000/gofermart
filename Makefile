test:
	go test -v -count 1 ./...

test100:
	go test -v -count 100 ./...

race:
	go test -v -race -count 1 ./...

# Название файла с точкой входа
MAIN=main.go

# Путь где создать бинарник
APP_DIR=cmd/gophermart

# Наименование бинарника
BIN_NAME=gophermart

# Подключение к базе данных
DB=postgres://gofermart:DB_gf-2023@127.0.0.1:5432/gofermart?sslmode=disable

# HELP по флагам
short_h:
	./$(APP_DIR)/$(BIN_NAME) -h

# Запустить сервис gofermart сконфигурировав его от файла json (config/config.json)
conf:
	./$(APP_DIR)/$(BIN_NAME) -c config/config_tt.json

# Запустить сервис gofermart с протоколом HTTPS
hs:
	sudo ./$(APP_DIR)/$(BIN_NAME) -s

# Запустить сервис gofermart и с протоколом HTTPS в фоновом режиме
hsf:
	sudo ./$(APP_DIR)/$(BIN_NAME) -s >/dev/null &

# Запустить сервис gofermart с подключением к DB
# FILE_STORAGE_PATH=;DATABASE_DSN=postgres://gofermart:DBshorten-2023@127.0.0.1:5432/gofermart?sslmode=disable
run:
	./$(APP_DIR)/$(BIN_NAME) -d $(DB)

# Скомпилировать и запустить бинарник сервиса gofermart (gofermart) с подключением к DB и запечёнными аргументами сборки
short:
	go build -o $(APP_DIR)/$(BIN_NAME) $(APP_DIR)/$(MAIN)
	./$(APP_DIR)/$(BIN_NAME)
# Запустить перед запуском сервера в отдельном терминале (для OS Linux)
accrual:
	./cmd/accrual/accrual_linux_amd64 -a localhost:8088
# Скомпилировать и запустить бинарник сервиса gofermart (gofermart) с подключением к DB
short_d:
	go build -o $(APP_DIR)/$(BIN_NAME) $(APP_DIR)/$(MAIN)
	./$(APP_DIR)/$(BIN_NAME) -d $(DB)

cover:
	go test -v -count 1 -race -coverpkg=./... -coverprofile=$(COVER_OUT) ./...
	go tool cover -func $(COVER_OUT)
	go tool cover -html=$(COVER_OUT)
	rm $(COVER_OUT)

cover1:
	go test -v -count 1  -coverpkg=./... -coverprofile=cover.out.tmp ./...
	cat cover.out.tmp | grep -v mocks/*  > cover.out2.tmp
	cat cover.out2.tmp | grep -v log/*  > $(COVER_OUT)
	go tool cover -func $(COVER_OUT)
	go tool cover -html=$(COVER_OUT)
	rm cover.out.tmp cover.out2.tmp
	rm $(COVER_OUT)

# Запустить сервис с документацией
# Доступен здесь: http://rooder.ru:6060/pkg/github.com/SETTER2000/gofermart/?m=all
godoc:
	godoc -http rooder.ru:6060

# Запустить сервис с документацией в фоновом режиме
doc:
	godoc -http=rooder.ru:6060  -play >/dev/null &
