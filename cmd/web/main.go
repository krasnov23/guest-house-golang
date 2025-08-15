package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/krasnov23/guest-house-golang/internal/config"
	"github.com/krasnov23/guest-house-golang/internal/driver"
	"github.com/krasnov23/guest-house-golang/internal/handlers"
	"github.com/krasnov23/guest-house-golang/internal/helpers"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"github.com/krasnov23/guest-house-golang/internal/render"
	"log"
	"net/http"
	"os"
	"time"
)

const portNumber = ":8082"

var app config.AppConfig

var session *scs.SessionManager

var infoLog *log.Logger

var errorLog *log.Logger

// main is the main function
func main() {

	db, err := run()

	if err != nil {
		log.Fatal(err)
	}

	defer db.SQL.Close()

	defer close(app.MailChan)
	listenForMail()

	/*from := "me@here.com"
	auth := smtp.PlainAuth("", from, "", "localhost")
	err = smtp.SendMail("localhost:1025", auth, from, []string{"you@there.com"}, []byte("hello world"))

	if err != nil {
		log.Println(err)
	}*/

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// Что я собираюсь добавлять в сессии
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// изменяем данное свойство когда выходим в продакшн
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	// будет отображать Дата + время + сообщение в консоли
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errLog

	session = scs.New()
	// 24 часа срок жизни сессии
	session.Lifetime = 24 * time.Hour
	// Если Persist = true, кука будет постоянной (persistent cookie), и сессия сохранится даже после закрытия браузера.
	session.Cookie.Persist = true
	// SameSite: http.SameSiteLaxMode кука будет отправляться только при безопасных запросах (например, GET) или при запросах с того же сайта
	session.Cookie.SameSite = http.SameSiteLaxMode
	// Если Secure = false, кука будет отправляться как по HTTP, так и по HTTPS. Если Secure установлено в true, кука будет отправляться только по HTTPS-соединениям
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Соединяемся с БД
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=app_db user=postgres password=postgres")

	if err != nil {
		log.Fatal("cannot connect to database")
	}
	log.Println("Connecting to database successfully")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	// Конфигурационный файл
	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandler(repo)

	// передаем конфигурационный файл в наш файл render, чтобы такие настройки как кэширование были видны в файле рендеринга
	render.NewRenderer(&app)

	helpers.NewHelpers(&app)

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	return db, nil
}
