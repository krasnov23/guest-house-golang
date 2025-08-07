package helpers

import (
	"fmt"
	"github.com/krasnov23/guest-house-golang/internal/config"
	"net/http"
	"runtime/debug"
)

var app *config.AppConfig

func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error: %d\n", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Хотя в вашей функции запрос не изменяется, в общем случае указатель позволяет модифицировать оригинальный объект запроса
// (например, добавлять или изменять поля в middleware).
// Использование *http.Request делает функцию совместимой с другими стандартными функциями и middleware, которые ожидают указатель.
func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")

	return exists
}
