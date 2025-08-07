package handlers

import (
	"bookings-udemy/internal/config"
	"bookings-udemy/internal/driver"
	"bookings-udemy/internal/forms"
	"bookings-udemy/internal/helpers"
	"bookings-udemy/internal/models"
	"bookings-udemy/internal/render"
	"bookings-udemy/internal/repository"
	"bookings-udemy/internal/repository/dbrepo"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

func NewRepo(app *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: app,
		DB:  dbrepo.NewPostgresDBRepo(db.SQL, app),
	}
}

func NewTestRepo(app *config.AppConfig) *Repository {
	return &Repository{
		App: app,
		DB:  dbrepo.NewTestingRepo(app),
	}
}

func NewHandler(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	m.DB.AllUsers()
	// Кладем Ip вновь зашедшего пользователя в сессию
	/*remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)*/

	err := render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		log.Fatal(err)
	}

}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["test"] = "Hello"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")

	stringMap["remote_ip"] = remoteIP

	render.Template(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Как система понимает "что это именно мы"?
// Куки → ID сессии → Данные
// Ваш браузер передает куку session=abc123
// Сервер проверяет, что abc123 есть в его хранилище
// Если есть — это вы, если нет — "чужой"
// Криптографическая защита
// Настоящий session_id подписан секретным ключом сервера
// Пример подписи: abc123.xyz456 (где xyz456 — HMAC-подпись)
// Подделать без ключа сервера невозможно
// Контекст — это "сейф"
// r.Context() содержит только ваши данные
// Другие пользователи физически не могут получить к ним доступ
// Это гарантирует архитектура Go и работа сервера
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// Session.Get достаёт из контекста все данные сессии
	// (которые middleware положил туда после проверки кук)
	//Ищет в них конкретное значение по ключу "reservation"
	//Пытается привести его к типу models.Reservation
	// Куки → session_id → данные — это происходит в middleware ДО попадания в ваш обработчик
	//В r.Context() попадают уже проверенные и загруженные данные
	//Сам session_id в контексте обычно НЕ хранится — там лежат уже готовые данные
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot find Room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	// Получение id сессии из куки
	/*cookie, err := r.Cookie("session")
	if err != nil {
		return
	}
	fmt.Println(cookie.Value)*/

	// преобразует к строке для нашей формы
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		// с помощью свойства ниже даем доступ к форме которая у нас есть в шаблоне
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	// Парсит данные формы из тела запроса (для POST)
	// Заполняет два поля структуры http.Request:
	// r.Form: Содержит данные как из URL (query parameters), так и из тела запроса (POST данные).
	// r.PostForm: Содержит только данные из тела запроса (POST данные).
	err := r.ParseForm()

	if err != nil {
		//helpers.ServerError(w, err)
		m.App.Session.Put(r.Context(), "error", "cannot parse data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data of room_id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	/*reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, errors.New("cannot get reservation from session"))
		return
	}*/

	/*reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")
	reservation.StartDate = startDate
	reservation.EndDate = endDate
	reservation.RoomID = roomID*/

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		RoomID:    roomID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	form := forms.New(r.PostForm)

	// проверяет что каждое из полей не пустое в противном случае добавляет ошибку в массив ошибок
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		http.Error(w, "my own error message", http.StatusSeeOther)

		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cant insert reservation to DB")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	hmtlMessage := fmt.Sprintf(`
		<strong> Reservation Confirmation </strong><br>
		Dear %s:, <br>
		This is confirm your reservation from %s to %s.
	`, reservation.FirstName,
		reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"))

	// отправка уведомления гостю
	msg := models.MailData{
		To:       reservation.Email,
		From:     "admin@gmail.com",
		Subject:  "Reservation confirmation",
		Content:  hmtlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cant insert room restrictions to DB")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// сохранение данных бронирования (reservation) в сессии пользователя.
	// В браузере мы не можем увидеть напрямую данные сессии так как сессии хранятся на сервере, но
	// мы может получить их из cookie по session_id
	m.App.Session.Put(r.Context(), "reservation", reservation)
	//render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{})

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {

	// Получаем из сессии значение по ключу reservation и приводим его к типу models.Reservation
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		m.App.ErrorLog.Println("cant get error from session")
		log.Println("cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	m.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})

}

// Данный метод отправки формы не будет работать если в нем не будет передаваться csrf токен,
// т.к в настройках нашего сервера задан csrf токен который должен в них отправляться
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, start)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, end)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	for _, i := range rooms {
		m.App.InfoLog.Println("ROOM:", i.ID, i.RoomName)
	}

	//Если нет свободных комнат на указанные даты
	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No rooms available")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})

	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// При данном методе генерируется реальная сессия в go
	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})

}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()

	if err != nil {

		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return

		/*if err != nil {
			helpers.ServerError(w, err)
			log.Println(err)
		}*/
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	endDate, err := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(roomID, startDate, endDate)

	if err != nil {

		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to DB",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return

	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, err := json.MarshalIndent(resp, "", "     ")

	if err != nil {
		helpers.ServerError(w, err)
		log.Println(err)
	}

	log.Println(string(out))

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(out)
	if err != nil {
		return
	}
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {

	// Получения id комнаты из реквеста
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "problem getting reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {

	// id,s(start),e (end)
	ID, err := strconv.Atoi(r.URL.Query().Get("id"))
	startDate := r.URL.Query().Get("s")
	endDate := r.URL.Query().Get("e")

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var res models.Reservation

	layout := "2006-01-02"

	sd, err := time.Parse(layout, startDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	ed, err := time.Parse(layout, endDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomByID(ID)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = ID
	res.StartDate = sd
	res.EndDate = ed

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

// ShowLogin -
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {

	// В шаблоне login.page.tmpl есть конструкции типа {{with .Form.Errors.Get "email"}}, которые требуют, чтобы .Form был не nil.
	// Если Form не передан, шаблон может либо:
	// Не отображать поля (если есть {{if .Form}} или {{with .Form}}),
	// Вызвать панику (если .Form — nil, но код пытается вызвать .Errors.Get).
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// https://go.dev/play/p/uKMMCzJWGsW - генерация пароля
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	//
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)

	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// Полностью удаляет данные сессии из хранилища (например, Redis, базы данных или файла).
	// Внутри, скорее всего, происходит удаление ключа сессии (например, session:<session_id>).
	_ = m.App.Session.Destroy(r.Context())

	// Генерирует новый идентификатор сессии (session ID), чтобы избежать атак типа session fixation.
	// Старая сессия уже уничтожена, но браузеру выдается новый session_id (через cookie).
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {

	reservations, err := m.DB.AllNewReservations()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {

	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {

	// делит роут по которому идет запрос по /
	exploded := strings.Split(r.RequestURI, "/")

	// берем ту часть где id и приводим к числу
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["year"] = year
	stringMap["month"] = month

	res, err := m.DB.GetReservationByID(id)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// делит роут по которому идет запрос по /
	exploded := strings.Split(r.RequestURI, "/")

	// берем ту часть где id и приводим к числу
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "flash", "changes saved")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {

	// Получение текущего времени
	now := time.Now()

	// Получение кверипараметра если такой будет передан
	if r.URL.Query().Get("y") != "" {
		year, err := strconv.Atoi(r.URL.Query().Get("y"))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		month, err := strconv.Atoi(r.URL.Query().Get("m"))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// Получение первого и последнего месяца в году
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {

		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-02")] = 0
			blockMap[d.Format("2006-01-02")] = 0
		}

		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)

		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				//reservationMap[y.StartDate.Format("2006-01-02")] = y.ReservationID
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-02")] = y.ReservationID
				}
			} else {
				blockMap[y.StartDate.Format("2006-01-02")] = y.ID
			}
		}

		for key, value := range reservationMap {
			if value != 0 {
				fmt.Println(key, value)
			}
		}

		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	/*log.Println("=== Форма данных (r.PostForm) ===")
	for key, values := range r.PostForm {
		log.Printf("%s: %v\n", key, values)
	}
	log.Println("================================")*/

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	rooms, err := m.DB.AllRooms()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, x := range rooms {
		// Получение заблокированных дат (тех дат у которых нет reservation_id).
		// Проходимся по нашей мапе и если в нашей форме нету value то есть мы сняли его с заблокированных дат
		// Если наш id в room_restrictions > 0, то мы должны удалить их c нашей базы
		// Достаем мапу формата ['2025-03-03' : room_restriction_id] из сессии и преобразуем в нужный для нас формат
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)

		for name, value := range curMap {
			if val, ok := curMap[name]; ok {
				// следим только за теми элементами где val > 0, и их нету в форме которая к нам пришла
				if val > 0 {
					// Если в форме нету name равного римувблок с айди и неймом то удаляем его с бд
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						// удаляем room_restriction по его айди
						err := m.DB.DeleteBlockById(value)

						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	// обработка новых блокировок, если name начинается с add_block_ то в базу вносятся новые блокировки
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block_") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			// вставляем блокировку по roomID и дате блокировки
			err := m.DB.InsertBlockForRoom(roomID, t)

			if err != nil {
				log.Println(err)
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}

func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.UpdateProcessedForReservation(id, 1)

	if err != nil {
		log.Println(err)
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "reservation mark as processed")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	_ = m.DB.DeleteReservation(id)

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}
