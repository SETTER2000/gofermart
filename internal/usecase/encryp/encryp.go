package encryp

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const (
	secretSecret = "RtsynerpoGIYdab_s234r"
)

var x interface{} = "access_token" //прочитать значение можно так: var keyToken string = x.(string)

type Encrypt struct{}

//	func RequireAuthentication(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			if !isAuthenticated(r) {
//				http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
//				return
//			}
//			// Assuming authentication passed, run the original handler
//			next.ServeHTTP(w, r)
//		})
//	}
func (e *Encrypt) isAuthenticated(r *http.Request) bool {
	at, err := r.Cookie("access_token")
	if err == http.ErrNoCookie {
		return false
	}
	// если кука обнаружена, то расшифровываем токен,
	// содержащийся в ней, и проверяем подпись
	_, err = e.DecryptToken(at.Value, secretSecret)
	if err != nil {
		fmt.Printf("error decrypt cookie: %e", err)
		return false
	}
	return true
}
func (e *Encrypt) isCookie(r *http.Request) bool {
	_, err := r.Cookie("access_token")
	return err != http.ErrNoCookie
}

// RequireAuthentication - middleware, которая устанавливает симметрично подписанную
// и зашифрованную куку устанавливается любому запросу не имеющему соответствующую куку
// или не прошедшая идентификацию, в куке зашифрован сгенерированный идентификатор пользователя
func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		en := Encrypt{}
		// если не обнаружена кука в запросе ...
		if !en.isAuthenticated(r) {
			http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
			return
		}
		//Предполагая, что аутентификация прошла, запустить исходный обработчик
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		en := Encrypt{}
		//idUser := ""
		//at, err := r.Cookie("access_token")
		// если не обнаружена кука в запросе ...
		if !en.isCookie(r) {
			ctx := r.Context()
			//en := Encrypt{}
			// ...создать подписанный секретным ключом токен,
			//token, err := en.EncryptToken(secretSecret)
			//if err != nil {
			//	fmt.Printf("Encrypt error: %v\n", err)
			//}
			// ...установить куку с именем access_token,
			// а в качестве значения установить зашифрованный,
			// подписанный токен
			http.SetCookie(w, &http.Cookie{
				Name:  "access_token",
				Path:  "/",
				Value: "",
				//Value: token,
				//Expires: time.Now().Add(time.Nanosecond * time.Duration(sessionLifeNanos)),
			})

			// декодируем token
			//idUser, err := en.DecryptToken(token, secretSecret)
			//if err != nil {
			//	fmt.Printf(" Decrypt error: %v\n", err)
			//}
			//ctx = context.WithValue(ctx, x, idUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
			//http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
			//return
		}
		//ctx = context.WithValue(ctx, x, idUser)
		//Предполагая, что аутентификация прошла, запустить исходный обработчик
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EncryptionKeyCookie - middleware
func EncryptionKeyCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		en := Encrypt{}
		idUser := ""
		at, err := r.Cookie("access_token")
		// если куки нет, то ничего не делаем, просто выходим
		if err == http.ErrNoCookie {
			ctx = context.WithValue(ctx, x, idUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
			// создать токен
			//token, err := en.EncryptToken(secretSecret)
			//if err != nil {
			//	fmt.Printf("Encrypt error: %v\n", err)
			//}
			////sessionLifeNanos := 100000000000
			//http.SetCookie(w, &http.Cookie{
			//	Name:  "access_token",
			//	Path:  "/",
			//	Value: token,
			//	//Expires: time.Now().Add(time.Nanosecond * time.Duration(sessionLifeNanos)),
			//})
			//
			//idUser, err = en.DecryptToken(token, secretSecret)
			//if err != nil {
			//	fmt.Printf(" Decrypt error: %v\n", err)
			//}
			//ctx = context.WithValue(ctx, x, idUser)
			//next.ServeHTTP(w, r.WithContext(ctx))
			//return
		}
		// если кука есть, то расшифровываем её и проверяем подпись
		idUser, err = en.DecryptToken(at.Value, secretSecret)
		// ...если подпись не соответствует, то очищаем куку и выходим
		if err != nil {
			fmt.Printf("Decrypt token error: %v\n", err)
			http.SetCookie(w, &http.Cookie{
				Name:  "access_token",
				Path:  "/",
				Value: "",
			})
			ctx = context.WithValue(ctx, x, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		// .. если подпись валидная,
		// то расшифрованный id юзера кладём в контекст в переменную access_token
		// т.е. должно получиться, что зашифрованный и подписанный токен,
		// всегда, содержит в себе id юзера и лишь попадая в систему при запросе от клиента
		// расшифровывается и кладётся в текущий контекст запроса.
		// Важная деталь!!
		// ID юзера попадает в токен, только после авторизации!
		// Так что если ID отсутствует в токене, значит юзер аноним.
		ctx = context.WithValue(ctx, x, idUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EncryptToken шифрование и подпись
// data - данные для кодирования
// secretKey - ключ для шифрования,
// из него создаётся ключ с помощью которого можно шифровать и расшифровать данные
// возвращает зашифрованную строку/токен
func (e *Encrypt) EncryptToken(secretKey string, data string) (string, error) {
	//data := scripts.UniqueString() //
	src := []byte(data) // данные, которые хотим зашифровать
	if len(src) < 1 {
		return "", fmt.Errorf("empty data arg")
	}
	// ключ шифрования, будем использовать AES256,
	// создав ключ длиной 32 байта (256 бит)
	key := sha256.Sum256([]byte(secretKey))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return "", err
	}

	// создаём вектор инициализации
	nonce := key[len(key)-aesgcm.NonceSize():]
	dst := aesgcm.Seal(nil, nonce, src, nil) // зашифровываем
	return fmt.Sprintf("%x", dst), nil
}

// DecryptToken расшифровать токен
// data - данные для расшифровки
// secretKey - ключ с помощью которого шифровались данные
// возвращает расшифрованную строку
func (e *Encrypt) DecryptToken(data string, secretKey string) (string, error) {
	// 1) получите ключ из password, используя sha256.Sum256
	key := sha256.Sum256([]byte(secretKey))

	// 2) создайте aesblock и aesgcm
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	// создаём вектор инициализации
	// 3) получите вектор инициализации aesgcm.NonceSize() байт с конца ключа
	nonce := key[len(key)-aesgcm.NonceSize():]

	// 4) декодируйте сообщение msg в двоичный формат
	encrypted, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}

	// расшифровываем
	// 5) расшифруйте и выведите данные
	decrypted, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		fmt.Printf("Chiper фонит!\n")
		return "", err
	}
	return string(decrypted), nil
}

func CheckToken(msg string) bool {
	var (
		data []byte // декодированное сообщение с подписью
		id   uint32 // значение идентификатора
		sign []byte // HMAC-подпись от идентификатора
	)
	data, err := hex.DecodeString(msg)
	if err != nil {
		panic(err)
	}
	key := []byte(secretSecret)
	//*****
	// 2) получите идентификатор из первых четырёх байт,
	//    используйте функцию binary.BigEndian.Uint32
	id = binary.BigEndian.Uint32(data[:4])
	// 3) вычислите HMAC-подпись sign для этих четырёх байт
	h := hmac.New(sha256.New, key)
	h.Write(data[:4])
	sign = h.Sum(nil)

	if hmac.Equal(sign, data[4:]) {
		fmt.Println("Подпись подлинная. ID:", id)
		return true
	}

	fmt.Println("Подпись неверна. Где-то ошибка...")
	return false
}

func SessionCreated(w http.ResponseWriter, r *http.Request, data string) error {
	//ctx := r.Context()
	en := Encrypt{}
	// ...создать подписанный секретным ключом токен,
	token, err := en.EncryptToken(secretSecret, data)
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

	_, err = en.DecryptToken(token, secretSecret)
	if err != nil {
		fmt.Printf(" Decrypt error: %v\n", err)
		return err
	}
	//context.WithValue(ctx, x, idUser)
	//next.ServeHTTP(w, r.WithContext(ctx))
	return nil
}
