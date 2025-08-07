package main

import (
	"github.com/justinas/nosurf"
	"github.com/krasnov23/guest-house-golang/internal/helpers"
	"log"
	"net/http"
)

// Создание нашего собственного middleWare который выводит в консоль: по каждому запросу
func WriteToConsole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Создание CSRF токена для защиты от POST запросов с посторонних сайтов
func NoSurf(next http.Handler) http.Handler {

	// создает новый CSRF middleware, оборачивая переданный обработчик
	csrfHandler := nosurf.New(next)

	// настраивает параметры куки, которая будет использоваться для хранения CSRF-токена.
	// HttpOnly: true — кука недоступна через JavaScript (защита от XSS-атак).
	// Path: "/" — кука будет отправляться на все пути сайта.
	// Secure: false — кука будет отправляться как по HTTP, так и по HTTPS
	// SameSite: http.SameSiteLaxMode кука будет отправляться только при безопасных запросах (например, GET) или при запросах с того же сайта
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Логируем ошибку CSRF
		log.Printf("CSRF token mismatch: form token=%s, cookie token=%s", r.FormValue("csrf_token"), nosurf.Token(r))
		http.Error(w, "CSRF token mismatch", http.StatusBadRequest)
	}))

	// csrfHandler имеет тип *nosurf.CSRFHandler,
	// который реализует интерфейс http.Handler. Это означает, что csrfHandler можно использовать как обычный обработчик HTTP-запросов.
	return csrfHandler
}

// После выполнения всех действий в текущем middleware (в данном случае — загрузки и сохранения сессии), управление передаётся этому обработчику.
// параметры, заданные в session = scs.New(), не будут применены, если middleware SessionLoad (или session.LoadAndSave) не используется.
// Без SessionLoad данные сессии не будут автоматически загружаться из хранилища.
// Это означает, что любые попытки получить данные сессии (например, через session.Get) вернут nil или ошибку, так как данные сессии не будут доступны в контексте запроса.
// Без SessionLoad изменения в сессии (например, через session.Put) не будут сохранены в хранилище.
// Это означает, что данные сессии будут "теряться" между запросами, и приложение не сможет поддерживать состояние пользователя.
// По сути, сессии перестанут работать. Вы не сможете: Хранить данные пользователя между запросами, Использовать сессии для аутентификации, хранения временных данных и других задач.
func SessionLoad(next http.Handler) http.Handler {
	// Эта функция загружает данные
	// сессии из хранилища (например, из cookie или базы данных) перед вызовом основного обработчика (next)
	// и сохраняет изменения в сессии после завершения работы основного обработчика.
	// переменная session указана глобально в main
	return session.LoadAndSave(next)
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "You need to authenticate")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		}

		next.ServeHTTP(w, r)
	})

}
