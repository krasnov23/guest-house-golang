package handlers

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/justinas/nosurf"
	"github.com/krasnov23/guest-house-golang/internal/config"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"github.com/krasnov23/guest-house-golang/internal/render"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const portNumber = ":8082"

var app config.AppConfig

var session *scs.SessionManager

var pathToTemplates = "./../../templates"

var functions = template.FuncMap{
	"humanDate":  render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate":    render.Iterate,
	"add":        render.Add,
}

func TestMain(m *testing.M) {

	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// изменяем данное свойство когда выходим в продакшн
	app.InProduction = false

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	// будет отображать Дата + время + сообщение в консоли
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(app.MailChan)

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	// Конфигурационный файл
	app.TemplateCache = tc

	// в данном случае задаем параметр true т.к в любом из функций роутинга будет дергаться (например внутри Repo.Home) функция RenderTemplate
	// в которой если UseCache = false он будет искать функцию в render.go , а там указан путь var pathToTemplates = "./templates"
	// ( нам нужен тот что в этой переменной у нас pathToTemplates )
	app.UseCache = true

	listenForMail()

	// В данном параметре создается тестовый репозиторий для поиска данных из бд
	repo := NewTestRepo(&app)
	NewHandler(repo)

	// прокидываются данные для функции RenderTemplate
	render.NewRenderer(&app)

	// m.Run() – запускает все тесты (*testing.M), которые определены в текущем пакете, и возвращает код завершения (exit code):
	//0 – если все тесты прошли успешно.
	//1 (или другой ненулевой код) – если хотя бы один тест упал.
	// os.Exit() – завершает программу с этим кодом, передавая его в операционную систему.
	os.Exit(m.Run())
}

func listenForMail() {

	go func() {
		for {
			_ = <-app.MailChan
		}
	}()

}

func getRoutes() http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(WriteToConsole)
	//mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)
	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)
	mux.Get("/search-availability", Repo.Availability)
	mux.Get("/choose-room/{id}", Repo.ChooseRoom)
	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	mux.Get("/contact", Repo.Contact)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}

// Создание нашего собственного middleWare который выводит в консоль: по каждому запросу
func WriteToConsole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit the page")
		next.ServeHTTP(w, r)
	})
}

func NoSurf(next http.Handler) http.Handler {

	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

func SessionLoad(next http.Handler) http.Handler {

	return session.LoadAndSave(next)
}

// CreateTestTemplateCache creates a template cache as a map
func CreateTestTemplateCache() (map[string]*template.Template, error) {

	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
