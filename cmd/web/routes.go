package main

import (
	"bookings-udemy/internal/config"
	"bookings-udemy/internal/handlers"
	"github.com/bmizerany/pat"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
)

// mux - создаем мультиплексер в который прокидываем роутинг, c помощью пакета pat
func patRoutes(app *config.AppConfig) http.Handler {
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Repo.Home))
	mux.Get("/about", http.HandlerFunc(handlers.Repo.About))

	return mux
}

// создаем мультплексер с помощью chi
func routes(app *config.AppConfig) http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(WriteToConsole)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)
	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)
	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)
	mux.Get("/contact", handlers.Repo.Contact)

	// http.Dir("./static/"): Это функция, которая создаёт файловую систему, указывая на папку ./static/ в вашем проекте.
	// Она позволяет Go получить доступ к файлам в этой папке (например, изображениям, CSS, JavaScript).
	// http.FileServer: Это функция, которая создаёт HTTP-обработчик для обслуживания статических
	// файлов из указанной файловой системы. В данном случае она будет обслуживать файлы из папки ./static/.
	fileServer := http.FileServer(http.Dir("./static/"))
	// mux.Handle: Это метод, который регистрирует обработчик для определённого маршрута.
	// В данном случае он регистрирует обработчик для всех запросов, начинающихся с /static/.
	// static/*: Это шаблон маршрута, который соответствует всем URL-адресам, начинающимся с /static/.
	// Например: static/css/style.css, static/images/logo.png
	// http.StripPrefix("/static", fileServer): Это функция, которая удаляет префикс /static из URL-адреса перед передачей запроса в fileServer
	// Например: Если запрошен URL /static/css/style.css, то после удаления префикса fileServer будет искать файл css/style.css в папке ./static/.
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		//mux.Use(Auth)

		mux.Get("/dashboard", handlers.Repo.AdminDashboard)
		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)
		mux.Get("/process-reservation/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", handlers.Repo.AdminDeleteReservation)

		mux.Get("/reservations/{src}/{id}/show", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)
	})

	return mux
}
