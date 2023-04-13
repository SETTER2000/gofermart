package scripts

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/SETTER2000/gofermart/config"
	"hash/fnv"
	"math/rand"
	"os"
	"time"
)

func FNV32a(text string) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(text))
	return algorithm.Sum32()
}

// RandBytes генерирует массив случайных байт. Размер массива передаётся параметром.
// Функция должна возвращать массив в виде строки в кодировке base64
func RandBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ``, err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
func GenerateString(n int) string {
	// generate string
	digits := "0123456789"
	//specials := "~=+%^*/()[]{}/!@#$?|"
	specials := "_"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" + digits + specials
	length := 3
	if n > length {
		length = n
	}

	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}
func UniqueString() string {
	return fmt.Sprintf("%v%s", time.Now().UnixNano(), GenerateString(3))
}

// GetHost формирует короткий URL
func GetHost(cfg config.HTTP, gofermart string) string {
	return fmt.Sprintf("%s/%s", cfg.BaseURL, gofermart)
}

// CheckEnvironFlag проверка значения переменной окружения и одноименного флага
// при отсутствие переменной окружения в самой среде или пустое значение этой переменной, проверяется
// значение флага с таким же именем, по сути сама переменная окружение отсутствовать не может в системе,
// идет лишь проверка значения в двух местах в начале в окружение, затем во флаге.
func CheckEnvironFlag(environName string, flagName string) bool {
	dsn, ok := os.LookupEnv(environName)
	if !ok || dsn == "" {
		dsn = flagName
		if dsn == "" {
			fmt.Printf("connect DSN string is empty: %v\n", dsn)
			return false
		}
	}
	return true
}
func TrimEmpty(s string) (string, error) {
	sz := len(s)
	var word string
	for i := 0; i < sz; i++ {
		if string(s[i]) != " " {
			word += string(s[i])
		}
	}
	return word, nil
}

// Trim удаляет первый и последний символ в строке s
// t - удаляется символ переданный в аргумент
// по умолчанию удаляет символ \n
func Trim(s string, t string) (string, error) {
	if s == "" {
		return s, fmt.Errorf("error arg s empty: %s", s)
	}
	sz := len(s)
	if sz > 0 && t != "" {
		if string(s[sz-1]) == t {
			s = s[:sz-1]
		}
		if string(s[0]) == t {
			s = s[1:]
		}
	}
	sz = len(s)
	if sz > 0 && s[sz-1] == '\n' {
		s = s[:sz-1]
	}
	if sz > 0 && s[0] == '\n' {
		s = s[1:]
	}

	return s, nil
}

func EncryptString(s string) (string, error) {
	salt := "poaleell"
	h := sha256.New()
	h.Write([]byte(s + salt))
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst), nil
}

func СheckPasswd() {
	var (
		data  []byte         // слайс случайных байт
		hash1 []byte         // хеш с использованием интерфейса hash.Hash
		hash2 [md5.Size]byte // хеш, возвращаемый функцией md5.Sum
	)

	// 1) генерация data длиной 512 байт
	data = make([]byte, 512)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}

	// 2) вычисление hash1 с использованием md5.New
	h := md5.New()
	h.Write(data)
	hash1 = h.Sum(nil)

	// 3) вычисление hash2 функцией md5.Sum
	hash2 = md5.Sum(data)

	// hash2[:] приводит массив байт к слайсу
	if bytes.Equal(hash1, hash2[:]) {
		fmt.Println("Всё правильно! Хеши равны")
	} else {
		fmt.Println("Что-то пошло не так")
	}
}
