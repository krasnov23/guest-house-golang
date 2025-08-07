package render

import (
	"fmt"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()

	if err != nil {
		t.Errorf("Error getting session: %v", err)
	}

	session.Put(r.Context(), "flash", "123")
	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("Failed")
	}

}

func TestRenderTemplate(t *testing.T) {
	// данный путь указывается потому что мы находимся в папке render, перезаписывае значение в данной переменной
	pathToTemplates = "./../../templates"

	// Создаем кэш шаблонов
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	// прокидываем его в наши конфиги чтобы он уже был готов при запуске RenderTemplate
	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	ww := httptest.NewRecorder()

	err = Template(ww, r, "home.page.tmpl", &models.TemplateData{})

	if err != nil {
		t.Error("error writing template in browser")
	}

	err = Template(ww, r, "unExisted.page.tmpl", &models.TemplateData{})

	if err == nil {
		t.Error("rendered template should have failed")
	}

}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

// В функции getSession() заголовок X-Session передается в session.Load, чтобы загрузить сессию.
// Если заголовок отсутствует, сессия не будет загружена, и данные из сессии не будут доступны.
func getSession() (*http.Request, error) {

	// эмуляция http запроса
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return nil, err
	}

	//sessionID := r.Header.Get("X-Session")
	//fmt.Printf("X-Session header value: '%s'\n", sessionID)

	// Из запроса извлекается текущий контекст (context.Context).
	// Контекст используется для передачи данных между middleware и обработчиками запросов. В данном случае, он будет использоваться для загрузки сессии.
	ctx := r.Context()

	// Если распечатаем ctx тут он будет без сессиий
	fmt.Println(ctx)

	// Заголовок X-Session — это пользовательский HTTP-заголовок, который используется для передачи идентификатора сессии.
	// это функция, которая загружает сессию из контекста. Обычно это часть библиотеки для работы с сессиями
	// извлекает значение заголовка X-Session из запроса. Этот заголовок может содержать идентификатор сессии
	// обновляет контекст, добавляя в него данные сессии.
	// в целом второй аргумент здесь не важен можно положить что угодно но тест будет проходить,
	// потому что библиотека для управления сессиями создает новую сессию, если идентификатор сессии неверный или отсутствует.
	ctx, err = session.Load(ctx, r.Header.Get("X-Session"))

	// Если распечатаем тут он уже будет с сессией внутри
	fmt.Println(ctx)

	if err != nil {
		return nil, err
	}

	// Создается новый запрос на основе старого, но с обновленным контекстом (ctx).
	// Контекст запроса теперь содержит данные сессии, которые могут быть использованы в дальнейшей обработке запроса.
	r = r.WithContext(ctx)

	return r, nil

}

func TestCreateTemplateCache(t *testing.T) {

	pathToTemplates = "./../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

}
