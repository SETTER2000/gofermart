package scripts

import (
	"fmt"
	"github.com/SETTER2000/shorturl/config"
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
func GetHost(cfg config.HTTP, shorturl string) string {
	return fmt.Sprintf("%s/%s", cfg.BaseURL, shorturl)
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
