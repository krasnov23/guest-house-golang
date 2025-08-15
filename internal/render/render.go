package render

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/krasnov23/guest-house-golang/internal/config"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"add":        Add,
}

var app *config.AppConfig

var pathToTemplates = "./templates"

// Прокидываем доступ к
func NewRenderer(a *config.AppConfig) {
	app = a
}

func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func Iterate(count int) []int {
	var i int
	var items []int

	for i = 0; i < count; i++ {
		items = append(items, i)
	}

	return items
}

func Add(a, b int) int {
	return a + b
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

// добавляет данные для всех шаблонов
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {

	// возвращает строковое значение для данного ключа, а затем удаляет его из данных сеанса. Статус данных сеанса будет изменен. Нулевое значение
	// для строки ("") возвращается, если ключ не существует или значение не может быть присвоен тип string.
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	// nosurf.Token(r) возвращает валидный токен.
	td.CSRFToken = nosurf.Token(r)

	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}

	return td
}

// RenderTemplate renders a template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {

	// объявляется переменная которая будет вида "название шаблона" => класс содержащий содержимое шаблона
	// "about.page.tmpl" =>  type Template struct

	var tc map[string]*template.Template

	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	// в переменную t помещается шаблон по его названию
	t, ok := tc[tmpl]
	if !ok {
		// log.Fatal завершает выполнение программы с кодом выхода 1 (аварийное завершение).
		// Это означает, что при вызове log.Fatal программа немедленно останавливается, и управление не возвращается к тесту.
		//log.Fatal("could not find template: ", tmpl)
		return errors.New("could not find template: " + tmpl)
	}

	//  Рендеринг сначала выполняется в буфер, чтобы:
	// Проверить на ошибки до отправки клиенту.
	// Избежать частичной записи в ResponseWriter при ошибках.
	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	// Применяет шаблон t к данным td.
	// Результат (HTML) записывается в buf
	_ = t.Execute(buf, td)

	// Если шаблон выполнен успешно, данные из буфера отправляются клиенту с помощью buf.WriteTo(w)
	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("error writing template to browser", err)
	}

	return nil
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {

	myCache := map[string]*template.Template{}

	// Поиск всех шаблонов с расширением *.page.tmpl
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {

		// filepath.Base извлекает имя файла из пути (например, "/templates/home.page.tmpl" → "home.page.tmpl").
		name := filepath.Base(page)

		// template.New(name) — создает новый шаблон с именем name.
		// добавляет пользовательские функции (например, для форматирования дат), если они определены в functions.
		//  парсит содержимое файла page и привязывает его к шаблону.
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// Поиск файлов макетов (.layout.tmpl)
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		// Если найдены файлы макетов (len(matches) > 0), они парсятся и добавляются к основному шаблону.
		// Например, home.page.tmpl может наследовать от base.layout.tmpl.
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		// Сохранение шаблона в кэш
		myCache[name] = ts
	}

	return myCache, nil
}
